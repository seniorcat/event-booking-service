package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Service interface {
	Get(ctx context.Context, key string, target interface{}) (bool, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeletePatern(ctx context.Context, pattern string) error
}

type service struct {
	redis *redis.Client
}

func NewService(redis *redis.Client) Service {
	return &service{redis: redis}
}

func (s *service) Get(ctx context.Context, key string, target interface{}) (bool, error) {
	return true, nil
}

func (s *service) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}

func (s *service) Delete(ctx context.Context, key string) error {
	return nil
}

func (s *service) DeletePatern(ctx context.Context, pattern string) error {
	return nil
}
