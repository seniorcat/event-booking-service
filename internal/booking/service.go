package booking

import (
	"context"
	"errors"
	"log"
	"time"

	"laschool.ru/event-booking-service/internal/notification"
)

type Service interface {
	Create(ctx context.Context, b *Booking, eventCapacity int) (int64, error)
	Get(ctx context.Context, id int64) (*Booking, error)
	ListByEvent(ctx context.Context, eventID int64, limit, offset int) ([]Booking, error)
	Cancel(ctx context.Context, id int64) error
}

type service struct {
	repo      Repository
	publisher notification.Publisher
}

func NewService(repo Repository, publisher notification.Publisher) Service {
	return &service{repo: repo, publisher: publisher}
}

func (s *service) Create(ctx context.Context, b *Booking, eventCapacity int) (int64, error) {
	if b.EventID == 0 || b.UserID == 0 {
		return 0, errors.New("event_id and user_id are required")
	}
	if b.Seats <= 0 {
		return 0, errors.New("seats must be positive")
	}
	used, err := s.repo.CountConfirmedSeats(ctx, b.EventID)
	if err != nil {
		return 0, err
	}
	if used+b.Seats > eventCapacity {
		return 0, errors.New("not enough seats")
	}
	bookingID, err := s.repo.Create(ctx, b)
	if err != nil {
		return 0, err
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic in publisher: %v", r)
			}
		}()
		event := &notification.NotificationEvent{
			Event:      "booking.created",
			BookingID:  bookingID,
			UserID:     b.UserID,
			EventID:    b.EventID,
			OccurredAt: time.Now().UTC(),
		}
		if err := s.publisher.Publish(context.Background(), event); err != nil {
			log.Printf("failed to publish booking.created: %v", err)
		}

	}()
	return bookingID, nil
}

func (s *service) Get(ctx context.Context, id int64) (*Booking, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) ListByEvent(ctx context.Context, eventID int64, limit, offset int) ([]Booking, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListByEvent(ctx, eventID, limit, offset)
}

func (s *service) Cancel(ctx context.Context, id int64) error {
	if id == 0 {
		return errors.New("id is required")
	}

	err := s.repo.Cancel(ctx, id)
	if err != nil {
		return err
	}

	// Получаем booking для данных уведомления
	b, err := s.repo.GetByID(ctx, id)
	if err == nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("panic in publisher: %v", r)
				}
			}()

			event := &notification.NotificationEvent{
				Event:      "booking.cancelled",
				BookingID:  b.ID,
				UserID:     b.UserID,
				EventID:    b.EventID,
				OccurredAt: time.Now().UTC(),
			}

			if err := s.publisher.Publish(context.Background(), event); err != nil {
				log.Printf("failed to publish booking.cancelled: %v", err)
			}
		}()
	}

	return nil
}
