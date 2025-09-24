package config

import (
	"laschool.ru/event-booking-service/pkg/container"
)

const DIConfig = "config"

func init() {
	container.Register(func(builder *container.Builder, _ map[string]interface{}) error {
		return builder.Add(container.Def{
			Name: DIConfig,
			Build: func(ctn container.Container) (interface{}, error) {
				return LoadConfig()
			},
		})
	})
}
