package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Service interface {
	Get(ctx context.Context, key string, target interface{}) (bool, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeletePatern(ctx context.Context, pattern string) error
	GetProtected(ctx context.Context, key string, calculate func() (interface{}, error), baseTTL time.Duration) (interface{}, error)
	GetWithLock(ctx context.Context, key string, calculate func() (interface{}, error), baseTTL time.Duration) (interface{}, error)
	WithJitter(baseTTL time.Duration) (time.Duration, error)
}

type service struct {
	redis *redis.Client
}

type ProtectionConfig struct {
	LockTTl        time.Duration
	MaxRetries     int
	BaseRetryDelay time.Duration
	JitterProcent  float64
}

func NewService(redis *redis.Client, lockTTLfromConf int) Service {
	return &service{
		redis: redis,
	}
}

func (s *service) Get(ctx context.Context, key string, target interface{}) (bool, error) {
	data, err := s.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("redis get failed: %w", err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return false, fmt.Errorf("cache data unmarshal failed: %w", err)
	}
	return true, nil
}

func (s *service) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	serialized, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache data marshal failed: %w", err)
	}
	if err := s.redis.Set(ctx, key, serialized, ttl).Err(); err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}
	return nil
}

func (s *service) Delete(ctx context.Context, key string) error {
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis delete failed: %w", err)
	}
	return nil
}

func (s *service) DeletePatern(ctx context.Context, pattern string) error {
	var cursor uint64
	var allKeys []string
	var err error

	for {
		var keys []string
		keys, cursor, err = s.redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("redis scan failed:%w", err)
		}
		allKeys = append(allKeys, keys...)
		if cursor == 0 {
			break
		}
	}
	batchSize := 100
	for i := 0; i < len(allKeys); i += batchSize {
		end := i + batchSize
		if end > len(allKeys) {
			end = len(allKeys)
		}
		batch := allKeys[i:end]
		if err := s.redis.Del(ctx, batch...).Err(); err != nil {
			return fmt.Errorf("redis batch delete failed: %w", err)
		}
	}

	return nil
}

func (s *service) GetProtected(ctx context.Context, key string, calculate func() (interface{}, error), baseTTL time.Duration) (interface{}, error) {
	var target interface{}
	found, err := s.Get(ctx, key, &target)
	if err != nil {
		return nil, fmt.Errorf("GetProtected redis error: %w", err)
	}
	if found {
		return target, nil
	}
	return s.GetWithLock(ctx, key, calculate, baseTTL)
}

func (s *service) GetWithLock(ctx context.Context, key string, calculate func() (interface{}, error), baseTTL time.Duration) (interface{}, error) {
	return nil, nil
}

func (s *service) WithJitter(baseTTL time.Duration) (time.Duration, error) {
	return 0, nil
}
