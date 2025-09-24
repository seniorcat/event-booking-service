package event

import (
	"context"
	"testing"
	"time"
)

type repoStub struct{}

func (repoStub) Create(ctx context.Context, e *Event) (int64, error)          { return 1, nil }
func (repoStub) GetByID(ctx context.Context, id int64) (*Event, error)        { return &Event{ID: id}, nil }
func (repoStub) List(ctx context.Context, limit, offset int) ([]Event, error) { return nil, nil }
func (repoStub) Update(ctx context.Context, e *Event) error                   { return nil }
func (repoStub) Delete(ctx context.Context, id int64) error                   { return nil }

func TestService_Create_Validation(t *testing.T) {
	svc := NewService(repoStub{})
	_, err := svc.Create(context.Background(), &Event{Title: "", Capacity: 10, StartsAt: time.Now(), EndsAt: time.Now().Add(time.Hour)})
	if err == nil {
		t.Fatal("expected error for empty title")
	}

	_, err = svc.Create(context.Background(), &Event{Title: "A", Capacity: 0, StartsAt: time.Now(), EndsAt: time.Now().Add(time.Hour)})
	if err == nil {
		t.Fatal("expected error for capacity <= 0")
	}

	_, err = svc.Create(context.Background(), &Event{Title: "A", Capacity: 10, StartsAt: time.Now().Add(time.Hour), EndsAt: time.Now()})
	if err == nil {
		t.Fatal("expected error for invalid dates")
	}
}

func TestService_Create_Success(t *testing.T) {
	svc := NewService(repoStub{})
	id, err := svc.Create(context.Background(), &Event{Title: "A", Capacity: 10, StartsAt: time.Now(), EndsAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero id")
	}
}
