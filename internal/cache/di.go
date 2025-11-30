package cache

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"laschool.ru/event-booking-service/pkg/container"
)

const DICacheService = "cache_servive"

func init() {
	container.Register(func(builder *container.Builder, _ map[string]interface{}) error {
		return builder.Add(container.Def{
			Name: DICacheService,
			Build: func(ctn container.Container) (interface{}, error) {
				redisRaw := ctn.Get(DIRedis)
				redisClient, ok := redisRaw.(*redis.Client)
				if !ok {
					return nil, fmt.Errorf("expected *redis.Client, got %T", redisRaw)
				}
				return NewService(redisClient), nil
			},
		})
	})
}
