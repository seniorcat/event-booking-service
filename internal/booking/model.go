package booking

import "time"

type Booking struct {
	ID        int64     `db:"id" json:"id"`
	EventID   int64     `db:"event_id" json:"event_id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	Seats     int       `db:"seats" json:"seats"`
	Status    string    `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
