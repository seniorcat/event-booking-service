package cache

import (
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"laschool.ru/event-booking-service/internal/config"
	"laschool.ru/event-booking-service/pkg/container"
)

const DICacheService = "cache_servive"

func init() {
	container.Register(func(builder *container.Builder, _ map[string]interface{}) error {
		return builder.Add(container.Def{
			Name: DICacheService,
			Build: func(ctn container.Container) (interface{}, error) {
				redisRaw := ctn.Get(DIRedis)
				cfg := ctn.Get(config.DIConfig).(*config.Config)
				redisClient, ok := redisRaw.(*redis.Client)
				if !ok {
					return nil, fmt.Errorf("expected *redis.Client, got %T", redisRaw)
				}
				return NewService(redisClient, time.Duration(cfg.Redis.LockTTL)), nil
			},
		})
	})
}
