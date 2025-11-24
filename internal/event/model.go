package event

import "time"

type Event struct {
	ID          int64     `db:"id" json:"id"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	Location    string    `db:"location" json:"location"`
	StartsAt    time.Time `db:"starts_at" json:"starts_at"`
	EndsAt      time.Time `db:"ends_at" json:"ends_at"`
	Capacity    int       `db:"capacity" json:"capacity"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// CreateEventRequest Модель запроса на создание события
type CreateEventRequest struct {
	Title       string    `json:"title" example:"Concert: The Rusty Cats"`
	Description string    `json:"description" example:"Live concert in the park"`
	Location    string    `json:"location" example:"Central Park"`
	StartsAt    time.Time `json:"starts_at" example:"2026-01-15T18:00:00Z"`
	EndsAt      time.Time `json:"ends_at" example:"2026-01-15T21:00:00Z"`
	Capacity    int       `json:"capacity" example:"100"`
}
