package cache_test

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"laschool.ru/event-booking-service/internal/cache"
	config "laschool.ru/event-booking-service/internal/config"
)

var svc cache.Service
var rdb redis.Client

func TestMain(m *testing.M) {
	configFile := filepath.Join("..", "..", "int-tests", "config.test.yaml")
	cfg, err := config.Load(configFile)
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		panic("failed to connect to Redis: " + err.Error())
	}
	svc = cache.NewService(rdb, time.Duration(cfg.Redis.LockTTL)*time.Second)

	// Очистка базы перед тестами
	if err := rdb.FlushDB(ctx).Err(); err != nil {
		log.Printf("warning: failed to flush DB: %v", err)
	}
	code := m.Run()

	// Очистка базы после тестов
	if err := rdb.FlushDB(ctx).Err(); err != nil {
		log.Printf("warning: failed to flush DB: %v", err)
	}

	os.Exit(code)
}

func TestIntegration_SetGetDelete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	key := "it:" + t.Name()
	t.Cleanup(func() { _ = svc.Delete(context.Background(), key) })
	value := map[string]string{"hello": "world"}
	require.NoError(t, svc.Set(ctx, key, value, time.Minute))
	var got map[string]string
	found, err := svc.Get(ctx, key, &got)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, value, got)
	require.NoError(t, svc.Delete(ctx, key))
	var newGot map[string]string
	found, err = svc.Get(ctx, key, &newGot)
	require.NoError(t, err)
	require.False(t, found)
}

func TestIntegration_DeletePattern(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	keys := []string{
		"test:key1",
		"test:key2",
		"test:key3",
		"test:key1:var1",
		"other:key1",
		"other:key2",
	}
	for _, key := range keys {
		err := rdb.Set(ctx, key, "value", 0).Err()
		require.NoError(t, err, "failed to set test key")
	}
	for _, key := range keys {
		exist, err := rdb.Exists(ctx, key).Result()
		require.NoError(t, err, "failed to check exist")
		require.Equal(t, int64(1), exist, "Key should exist before deletion")

	}
}
