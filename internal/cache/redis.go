package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"laschool.ru/event-booking-service/internal/config"
	"laschool.ru/event-booking-service/pkg/container"
)

const DIRedis = "redis"

func init() {
	container.Register(func(builder *container.Builder, _ map[string]interface{}) error {
		return builder.Add(container.Def{
			Name: DIRedis,
			Build: func(ctn container.Container) (interface{}, error) {
				cfg := ctn.Get(config.DIConfig).(*config.Config)
				rdb := redis.NewClient(&redis.Options{
					Addr:     cfg.Redis.Address,
					Password: cfg.Redis.Password,
					DB:       cfg.Redis.DB,

					DialTimeout:  time.Duration(cfg.Redis.DialTimeout) * time.Second,
					ReadTimeout:  time.Duration(cfg.Redis.ReadTimeout) * time.Second,
					WriteTimeout: time.Duration(cfg.Redis.WriteTimeout) * time.Second,
					PoolTimeout:  time.Duration(cfg.Redis.PoolTimeout) * time.Second,

					MaxRetries:      cfg.Redis.MaxRetries,
					MinRetryBackoff: time.Duration(cfg.Redis.MinRetryBackoff) * time.Second,
					MaxRetryBackoff: time.Duration(cfg.Redis.MaxRetryBackoff) * time.Second,

					PoolSize:     cfg.Redis.PoolSize,
					MinIdleConns: cfg.Redis.MinIdleConns,

					ConnMaxIdleTime: time.Duration(cfg.Redis.ConnMaxIdleTime) * time.Second,
					ConnMaxLifetime: time.Duration(cfg.Redis.ConnMaxLifetime) * time.Second,
				})

				// Проверим соединение
				if err := pingWithRetry(rdb, rdb.Options().MaxRetries, time.Duration(cfg.Redis.MinRetryBackoff)*time.Second); err != nil {
					return nil, fmt.Errorf("failed to connect to Redis after %d attempts: %w", cfg.Redis.MaxRetries, err)
				}
				return rdb, nil
			},
		})
	})
}

func pingWithRetry(rdb *redis.Client, maxRetries int, retryDelay time.Duration) error {

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := rdb.Ping(ctx).Err(); err != nil {
			cancel()
			return nil
		} else {
			lastErr = err
			cancel()

			if i < maxRetries-1 {
				time.Sleep(retryDelay)
			}
		}
	}
	return lastErr
}
