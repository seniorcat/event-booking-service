package booking

import (
	"context"
	"testing"
)

type repoStub struct {
	used int
}

func (r repoStub) Create(ctx context.Context, b *Booking) (int64, error) { return 1, nil }
func (r repoStub) GetByID(ctx context.Context, id int64) (*Booking, error) {
	return &Booking{ID: id}, nil
}
func (r repoStub) ListByEvent(ctx context.Context, eventID int64, limit, offset int) ([]Booking, error) {
	return nil, nil
}
func (r repoStub) Cancel(ctx context.Context, id int64) error { return nil }
func (r repoStub) CountConfirmedSeats(ctx context.Context, eventID int64) (int, error) {
	return r.used, nil
}

func TestService_Create_CapacityExceeded(t *testing.T) {
	svc := NewService(repoStub{used: 9})
	_, err := svc.Create(context.Background(), &Booking{EventID: 1, UserID: 1, Seats: 2}, 10)
	if err == nil {
		t.Fatal("expected not enough seats error")
	}
}

func TestService_Create_Success(t *testing.T) {
	svc := NewService(repoStub{used: 5})
	id, err := svc.Create(context.Background(), &Booking{EventID: 1, UserID: 1, Seats: 3}, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero id")
	}
}
