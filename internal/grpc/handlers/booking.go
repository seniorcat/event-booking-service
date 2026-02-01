package handlers

import (
	"context"

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
	// if req.EventId == 0 || req.UserId == 0 {
	// 	return nil, fmt.Errorf("event_id and user_id are required")
	// }
	// newBooking := &booking.Booking{
	// 	EventID:   req.EventId,
	// 	UserID:    req.UserId,
	// 	Seats:     int(req.Seats),
	// 	Status:    "confirmed",
	// 	CreatedAt: time.Now(),
	// }

	// id, err := s.service.Create(ctx, newBooking, newBooking.Seats)

	return nil, nil
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
