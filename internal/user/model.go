package user

import "time"

type User struct {
	ID        int64     `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Name      string    `db:"name" json:"name"`
	Password  string    `db:"password_hash" json:"password_hash"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// LoginRequest represents payload for login endpoint
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse contains JWT token returned after successful login
type AuthResponse struct {
	Token string `json:"token"`
}

// RegisterRequest модель запроса для регистрации пользователя
type RegisterRequest struct {
	Name     string `json:"name" example:"Alice"`
	Email    string `json:"email" example:"alice@example.com"`
	Password string `json:"password" example:"password123"`
}

// RegisterResponse модель ответа после успешной регистрации пользователя
type RegisterResponse struct {
	ID      int64  `json:"id" example:"1"`
	Message string `json:"message" example:"user registered successfully"`
}
