package booking

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Create(ctx context.Context, b *Booking) (int64, error)
	GetByID(ctx context.Context, id int64) (*Booking, error)
	ListByEvent(ctx context.Context, eventID int64, limit, offset int) ([]Booking, error)
	Cancel(ctx context.Context, id int64) error
	CountConfirmedSeats(ctx context.Context, eventID int64) (int, error)
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, b *Booking) (int64, error) {
	const q = `INSERT INTO bookings (event_id, user_id, seats, status) VALUES ($1,$2,$3,'confirmed') RETURNING id`
	var id int64
	if err := r.db.QueryRowxContext(ctx, q, b.EventID, b.UserID, b.Seats).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *repository) GetByID(ctx context.Context, id int64) (*Booking, error) {
	const q = `SELECT id, event_id, user_id, seats, status, created_at FROM bookings WHERE id=$1`
	var b Booking
	if err := r.db.GetContext(ctx, &b, q, id); err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *repository) ListByEvent(ctx context.Context, eventID int64, limit, offset int) ([]Booking, error) {
	const q = `SELECT id, event_id, user_id, seats, status, created_at FROM bookings WHERE event_id=$1 ORDER BY id DESC LIMIT $2 OFFSET $3`
	var list []Booking
	if err := r.db.SelectContext(ctx, &list, q, eventID, limit, offset); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *repository) Cancel(ctx context.Context, id int64) error {
	const q = `UPDATE bookings SET status='cancelled' WHERE id=$1`
	_, err := r.db.ExecContext(ctx, q, id)
	return err
}

func (r *repository) CountConfirmedSeats(ctx context.Context, eventID int64) (int, error) {
	const q = `SELECT COALESCE(SUM(seats),0) FROM bookings WHERE event_id=$1 AND status='confirmed'`
	var total int
	if err := r.db.GetContext(ctx, &total, q, eventID); err != nil {
		return 0, err
	}
	return total, nil
}
