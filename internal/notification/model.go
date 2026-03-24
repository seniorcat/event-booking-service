package notification

import "time"

type NotificationEvent struct {
	Event      string    `json:"event"`
	BookingID  int64     `json:"booking_id"`
	UserID     int64     `json:"user_id"`
	EventID    int64     `json:"event_id"`
	OccurredAt time.Time `json:"occurred_at"`
}
