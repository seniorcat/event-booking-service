package event

import (
	"context"
	"errors"
	"time"
)

type Service interface {
	Create(ctx context.Context, e *Event) (int64, error)
	Get(ctx context.Context, id int64) (*Event, error)
	List(ctx context.Context, limit, offset int) ([]Event, error)
	Update(ctx context.Context, e *Event) error
	Delete(ctx context.Context, id int64) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, e *Event) (int64, error) {
	if e.Title == "" {
		return 0, errors.New("title is required")
	}
	if e.StartsAt.IsZero() || e.EndsAt.IsZero() || !e.EndsAt.After(e.StartsAt) {
		return 0, errors.New("invalid dates")
	}
	if e.Capacity <= 0 {
		return 0, errors.New("capacity must be positive")
	}
	return s.repo.Create(ctx, e)
}

func (s *service) Get(ctx context.Context, id int64) (*Event, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, limit, offset int) ([]Event, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset)
}

func (s *service) Update(ctx context.Context, e *Event) error {
	if e.ID == 0 {
		return errors.New("id is required")
	}
	if e.UpdatedAt.IsZero() {
		e.UpdatedAt = time.Now()
	}
	return s.repo.Update(ctx, e)
}

func (s *service) Delete(ctx context.Context, id int64) error {
	if id == 0 {
		return errors.New("id is required")
	}
	return s.repo.Delete(ctx, id)
}
