package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	go func(b *booking.Booking, bookingID int64, eventID int64) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// Удаляем кэш для этого события
		pattern := fmt.Sprintf("event:%d:bookings*", req.EventId)
		if err := s.cacheService.DeletePattern(ctx, pattern); err != nil {
			log.Printf("WARNING: Cache invalidation failed for event %d: %v", eventID, err)
		} else {
			log.Printf("Cache invalidated for event %d after booking creation", req.EventId)
		}

		bookingKey := fmt.Sprintf("booking:%d", bookingID)
		if err := s.cacheService.Set(ctx, bookingKey, newBooking, 30*time.Minute); err != nil {
			log.Printf("WARNING: Failed to cache booking %d: %v", bookingID, err)
		} else {
			log.Printf("Booking %d cached successfully", bookingID)
		}
	}(newBooking, id, req.EventId)

	return &pb.CreateBookingResponse{Id: id}, nil
}

func (s *bookingGRPCServer) CancelBooking(ctx context.Context, req *pb.CancelBookingRequest) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *bookingGRPCServer) GetBooking(ctx context.Context, req *pb.GetBookingRequest) (*pb.Booking, error) {
	return nil, nil
}

func (s *bookingGRPCServer) ListBookings(req *pb.ListBookingsRequest, stream pb.BookingService_ListBookingsServer) error {
	return nil
}

// func stringToProtoStatus(status string) pb.BookingStatus {
// 	switch status {
// 	case "confirmed":
// 		return pb.BookingStatus_BOOKING_STATUS_CONFIRMED
// 	case "cancelled":
// 		return pb.BookingStatus_BOOKING_STATUS_CANCELLED
// 	case "pending":
// 		return pb.BookingStatus_BOOKING_STATUS_PENDING
// 	default:
// 		return pb.BookingStatus_BOOKING_STATUS_UNSPECIFIED
// 	}
// }
