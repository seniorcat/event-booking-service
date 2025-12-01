package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	retryDelayConst    = 50 * time.Millisecond
	maxRetryDelayConst = 2 * time.Second
	maxRetriesConst    = 10
)

type Service interface {
	Get(ctx context.Context, key string, target interface{}) (bool, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
	GetProtected(ctx context.Context, key string, calculate func() (interface{}, error), baseTTL time.Duration) ([]byte, error)
	GetWithLock(ctx context.Context, key string, calculate func() (interface{}, error), baseTTL time.Duration) (interface{}, error)
	WithJitter(baseTTL time.Duration) time.Duration
}

type service struct {
	redis         *redis.Client
	LockTTL       time.Duration
	retryDelay    time.Duration
	maxRetryDelay time.Duration
	maxRetries    int
}

func NewService(redis *redis.Client, lockTTL time.Duration) Service {
	return &service{
		redis:         redis,
		LockTTL:       lockTTL,
		retryDelay:    retryDelayConst,
		maxRetryDelay: maxRetryDelayConst,
		maxRetries:    maxRetriesConst,
	}
}

func (s *service) Get(ctx context.Context, key string, target interface{}) (bool, error) {

	if target == nil {
		return false, errors.New("target cannot be nil")
	}
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

func (s *service) DeletePattern(ctx context.Context, pattern string) error {
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

func (s *service) GetProtected(ctx context.Context, key string, calculate func() (interface{}, error), baseTTL time.Duration) ([]byte, error) {
	data, err := s.redis.Get(ctx, key).Bytes()
	if err == nil && len(data) > 0 {
		return data, nil
	}

	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("GetProtected redis error: %w", err)
	}

	// Кэш промах - используем GetWithLock
	result, err := s.GetWithLock(ctx, key, calculate, baseTTL)
	if err != nil {
		return nil, err
	}

	// Сериализуем результат обратно в JSON
	return json.Marshal(result)
}

func (s *service) GetWithLock(ctx context.Context, key string, calculate func() (interface{}, error), baseTTL time.Duration) (interface{}, error) {
	lockKey := key + ":lock"
	gotLock, err := s.tryAcquireLock(ctx, lockKey)
	if err != nil {
		return nil, err
	}
	if gotLock {
		defer s.releaseLock(ctx, lockKey)
		return s.calculateAndStore(ctx, key, calculate, baseTTL)
	}
	return s.waitForCalculation(ctx, key, calculate, baseTTL)
}

func (s *service) tryAcquireLock(ctx context.Context, lockKey string) (bool, error) {
	success, err := s.redis.SetNX(ctx, lockKey, "1", s.LockTTL).Result()
	if err != nil {
		return false, fmt.Errorf("acquire lock failed:%w", err)
	}
	return success, nil
}

func (s *service) releaseLock(ctx context.Context, lockKey string) {
	if err := s.redis.Del(ctx, lockKey).Err(); err != nil {
		log.Printf("WARNING: Failed to release lock %s: %v", lockKey, err)
	}
}

func (s *service) calculateAndStore(ctx context.Context, key string, calculate func() (interface{}, error), baseTTL time.Duration) (interface{}, error) {
	data, err := calculate()
	if err != nil {
		return nil, fmt.Errorf("calculate function failed: %w", err)
	}
	ttl := s.WithJitter(baseTTL)
	if err := s.Set(ctx, key, data, ttl); err != nil {
		log.Printf("WARNING: Failed to cache data for key %s: %v", key, err)
	}

	return data, nil
}

func (s *service) waitForCalculation(ctx context.Context, key string, calculate func() (interface{}, error), baseTTL time.Duration) (interface{}, error) {
	currentDelay := s.retryDelay // Локальная копия

	for i := 0; i < s.maxRetries; i++ {
		select {
		case <-time.After(currentDelay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		var result interface{}
		found, err := s.Get(ctx, key, &result)
		if err != nil {
			return nil, err
		}
		if found {
			return result, nil
		}

		// Изменяем локальную переменную, а не поле структуры
		currentDelay = time.Duration(float64(currentDelay) * 1.5)
		if currentDelay > s.maxRetryDelay {
			currentDelay = s.maxRetryDelay
		}
	}

	log.Printf("WARNING: Cache stampede protection timeout for key %s, calculating ourselves", key)
	return s.calculateAndStore(ctx, key, calculate, baseTTL)
}

func (s *service) WithJitter(baseTTL time.Duration) time.Duration {
	if baseTTL <= 0 {
		return baseTTL
	}
	// Добавочный jitter: ±10% от baseTTL
	jitterRange := float64(baseTTL) * 0.1
	jitter := time.Duration(rand.Int63n(int64(2*jitterRange)) - int64(jitterRange))

	result := baseTTL + jitter

	// Гарантируем что TTL не станет отрицательным
	if result < time.Second {
		return time.Second
	}
	return result
}
