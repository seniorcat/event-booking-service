package handlers

import (
	"context"

	"laschool.ru/event-booking-service/internal/booking"
	"laschool.ru/event-booking-service/internal/cache"
	"laschool.ru/event-booking-service/pkg/container"
)

func GetListBookingsWithCache(ctx context.Context, ctn container.Container, eventID int64, limit, offset int) ([]booking.Booking, error) {
	bsvc := ctn.Get(booking.DIBookingService).(booking.Service)
	cacheService := ctn.Get(cache.DICacheService).(cache.Service)

}
