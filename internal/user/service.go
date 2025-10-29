package user

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"laschool.ru/event-booking-service/internal/jwtutil"
)

type Service interface {
	Register(ctx context.Context, u *User) (int64, error)
	Login(ctx context.Context, email, password string) (string, error)
}

type service struct {
	repo   Repository
	secret string
	ttl    time.Duration
}

func NewService(repo Repository, secret string, ttl time.Duration) Service {

	return &service{repo: repo, secret: secret, ttl: ttl}
}

func (s *service) Register(ctx context.Context, user *User) (int64, error) {
	return s.repo.Create(ctx, user)
}

func (s *service) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid password")
	}

	token, err := jwtutil.GenerateJWT(user.ID, s.secret, s.ttl)
	if err != nil {
		return "", fmt.Errorf("generate token error: %w", err)
	}

	return token, nil
}
