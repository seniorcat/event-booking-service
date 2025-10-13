package user

import "time"

type User struct {
	ID        int64     `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Name      string    `db:"name" json:"name"`
	Password  string    `db:"password_hash" json:"password_hash"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
