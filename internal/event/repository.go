package event

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	Create(ctx context.Context, e *Event) (int64, error)
	GetByID(ctx context.Context, id int64) (*Event, error)
	List(ctx context.Context, limit, offset int) ([]Event, error)
	Update(ctx context.Context, e *Event) error
	Delete(ctx context.Context, id int64) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, e *Event) (int64, error) {
	const q = `
        INSERT INTO events (title, description, location, starts_at, ends_at, capacity)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
    `
	var id int64
	if err := r.db.QueryRowxContext(ctx, q, e.Title, e.Description, e.Location, e.StartsAt, e.EndsAt, e.Capacity).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *repository) GetByID(ctx context.Context, id int64) (*Event, error) {
	const q = `SELECT id, title, description, location, starts_at, ends_at, capacity, created_at, updated_at FROM events WHERE id=$1`
	var e Event
	if err := r.db.GetContext(ctx, &e, q, id); err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *repository) List(ctx context.Context, limit, offset int) ([]Event, error) {
	const q = `
        SELECT id, title, description, location, starts_at, ends_at, capacity, created_at, updated_at
        FROM events
        ORDER BY starts_at DESC
        LIMIT $1 OFFSET $2
    `
	var events []Event
	if err := r.db.SelectContext(ctx, &events, q, limit, offset); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *repository) Update(ctx context.Context, e *Event) error {
	const q = `
        UPDATE events
        SET title=$1, description=$2, location=$3, starts_at=$4, ends_at=$5, capacity=$6, updated_at=NOW()
        WHERE id=$7
    `
	_, err := r.db.ExecContext(ctx, q, e.Title, e.Description, e.Location, e.StartsAt, e.EndsAt, e.Capacity, e.ID)
	return err
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM events WHERE id=$1`
	_, err := r.db.ExecContext(ctx, q, id)
	return err
}
