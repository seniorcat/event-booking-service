package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	pb "laschool.ru/event-booking-service/api/v1/booking"
	"laschool.ru/event-booking-service/internal/booking"
	"laschool.ru/event-booking-service/internal/cache"
	"laschool.ru/event-booking-service/internal/event"
)

type bookingGRPCServer struct {
	pb.UnimplementedBookingServiceServer
	bookingService booking.Service
	eventService   event.Service
	cacheService   cache.Service
}

func NewGRPCBookingServer(
	bookingService booking.Service,
	eventService event.Service,
	cacheService cache.Service,
) pb.BookingServiceServer {
	return &bookingGRPCServer{
		bookingService: bookingService,
		eventService:   eventService,
		cacheService:   cacheService,
	}
}

func (s *bookingGRPCServer) CreateBooking(ctx context.Context, req *pb.CreateBookingRequest) (*pb.CreateBookingResponse, error) {
	if req.EventId == 0 {
		return nil, status.Error(codes.InvalidArgument, "event_id is required")
	}
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.Seats <= 0 {
		return nil, status.Error(codes.InvalidArgument, "seats must be positive")
	}

	e, err := s.eventService.Get(ctx, req.EventId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "event not found")
	}
	newBooking := &booking.Booking{
		EventID:   req.EventId,
		UserID:    req.UserId,
		Seats:     int(req.Seats),
		Status:    "confirmed",
		CreatedAt: time.Now(),
	}
	id, err := s.bookingService.Create(ctx, newBooking, e.Capacity)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create booking: "+err.Error())
	}
	go func(b *booking.Booking, bookingID int64, eventID int64) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// Удаляем кэш для этого события
		pattern := fmt.Sprintf("event:%d:bookings*", eventID)
		if err := s.cacheService.DeletePattern(ctx, pattern); err != nil {
			log.Printf("WARNING: Cache invalidation failed for event %d: %v", eventID, err)
		} else {
			log.Printf("Cache invalidated for event %d after booking creation", eventID)
		}

		bookingKey := fmt.Sprintf("booking:%d", bookingID)
		if err := s.cacheService.Set(ctx, bookingKey, b, 30*time.Minute); err != nil {
			log.Printf("WARNING: Failed to cache booking %d: %v", bookingID, err)
		} else {
			log.Printf("Booking %d cached successfully", bookingID)
		}
	}(newBooking, id, req.EventId)

	return &pb.CreateBookingResponse{Id: id}, nil
}

func (s *bookingGRPCServer) CancelBooking(ctx context.Context, req *pb.CancelBookingRequest) (*emptypb.Empty, error) {
	if req.BookingId == 0 {
		return nil, status.Error(codes.InvalidArgument, "booking_id is required")
	}
	if err := s.bookingService.Cancel(ctx, req.BookingId); err != nil {
		return nil, status.Error(codes.Internal, "failed to cancel booking: "+err.Error())
	}
	bookingID := req.BookingId
	go func(id int64) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Удаляем все связанное с бронированиями
		_ = s.cacheService.DeletePattern(ctx, "event:*:bookings*")
		_ = s.cacheService.DeletePattern(ctx, fmt.Sprintf("booking:%d", id))
		_ = s.cacheService.DeletePattern(ctx, "stats:bookings*")
	}(bookingID)
	return &emptypb.Empty{}, nil
}

func (s *bookingGRPCServer) GetBooking(ctx context.Context, req *pb.GetBookingRequest) (*pb.Booking, error) {
	if req.BookingId == 0 {
		return nil, status.Error(codes.InvalidArgument, "booking_id is required")
	}
	b, err := s.bookingService.Get(ctx, req.BookingId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "booking is not found")
	}
	return &pb.Booking{
		Id:        b.ID,
		EventId:   b.EventID,
		UserId:    b.UserID,
		Seats:     int32(b.Seats),
		Status:    stringToProtoStatus(b.Status),
		CreatedAt: timestamppb.New(b.CreatedAt),
	}, nil
}

func (s *bookingGRPCServer) ListBookings(
	req *pb.ListBookingsRequest,
	stream pb.BookingService_ListBookingsServer,
) error {
	// ===== ВАЛИДАЦИЯ =====
	if req.EventId == 0 {
		return status.Error(codes.InvalidArgument, "event_id is required")
	}

	// Validate и установить defaults для limit
	limit := int(req.Limit)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Validate и установить defaults для offset
	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	log.Printf("ListBookings: event_id=%d, limit=%d, offset=%d", req.EventId, limit, offset)

	// ===== КЭШИРОВАНИЕ =====
	cacheKey := fmt.Sprintf("event:%d:bookings:limit:%d:offset:%d", req.EventId, limit, offset)

	// Функция для получения данных из БД
	calculateFunc := func() (interface{}, error) {
		return s.bookingService.ListByEvent(
			stream.Context(),
			req.EventId,
			limit,
			offset,
		)
	}

	// Попытаться получить из кэша, иначе из БД
	data, err := s.cacheService.GetProtected(
		stream.Context(),
		cacheKey,
		calculateFunc,
		5*time.Minute,
	)
	if err != nil {
		log.Printf("ERROR: Failed to get bookings: %v", err)
		return status.Error(codes.Internal, "failed to list bookings")
	}

	// ===== РАСПАРСИТЬ JSON =====
	var bookings []booking.Booking
	if err := json.Unmarshal(data, &bookings); err != nil {
		log.Printf("ERROR: Failed to unmarshal bookings: %v", err)
		return status.Error(codes.Internal, "failed to parse bookings")
	}

	log.Printf("Got %d bookings from cache/DB", len(bookings))

	// ===== ОТПРАВИТЬ ЧЕРЕЗ STREAM =====
	for _, b := range bookings {
		// Преобразовать в proto
		protoBooking := &pb.Booking{
			Id:        b.ID,
			EventId:   b.EventID,
			UserId:    b.UserID,
			Seats:     int32(b.Seats),
			Status:    stringToProtoStatus(b.Status),
			CreatedAt: timestamppb.New(b.CreatedAt),
		}

		// Отправить клиенту
		if err := stream.Send(protoBooking); err != nil {
			log.Printf("ERROR: Failed to send booking %d: %v", b.ID, err)
			return status.Error(codes.Internal, "failed to send booking")
		}

		log.Printf("Sent booking %d", b.ID)
	}

	// ===== УСПЕХ =====
	log.Printf("ListBookings completed successfully, sent %d bookings", len(bookings))
	return nil
}

func stringToProtoStatus(status string) pb.BookingStatus {
	switch status {
	case "confirmed":
		return pb.BookingStatus_BOOKING_STATUS_CONFIRMED
	case "cancelled":
		return pb.BookingStatus_BOOKING_STATUS_CANCELLED
	case "pending":
		return pb.BookingStatus_BOOKING_STATUS_PENDING
	default:
		return pb.BookingStatus_BOOKING_STATUS_UNSPECIFIED
	}
}
