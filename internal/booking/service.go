package booking

import (
	"context"
	"errors"
)

type Service interface {
	Create(ctx context.Context, b *Booking, eventCapacity int) (int64, error)
	Get(ctx context.Context, id int64) (*Booking, error)
	ListByEvent(ctx context.Context, eventID int64, limit, offset int) ([]Booking, error)
	Cancel(ctx context.Context, id int64) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
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
	return s.repo.Create(ctx, b)
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
	return s.repo.Cancel(ctx, id)
}
