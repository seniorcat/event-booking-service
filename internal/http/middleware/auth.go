// internal/http/middleware/auth.go
package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"laschool.ru/event-booking-service/internal/config"
	"laschool.ru/event-booking-service/internal/jwtutil"
	"laschool.ru/event-booking-service/pkg/container"
)

type contextKey string

const UserIDKey contextKey = "userID"

// NewAuthMiddleware достаёт cfg из контейнера ОДИН РАЗ и возвращает middleware.
func NewAuthMiddleware() (func(http.Handler) http.Handler, error) {
	ctn, err := container.Instance(nil, nil)
	if err != nil {
		return nil, err
	}
	cfg := ctn.Get(config.DIConfig).(*config.Config)
	secret := cfg.JWT.Secret

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "missing or invalid Authorization header", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			claims, err := jwtutil.ValidateJWT(tokenStr, secret)
			if err != nil {
				if strings.Contains(err.Error(), "expired") || errors.Is(err, jwt.ErrTokenExpired) {
					http.Error(w, "token expired", http.StatusForbidden)
				} else {
					http.Error(w, "invalid token", http.StatusUnauthorized)
				}
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}, nil
}
