package notification

import (
	"laschool.ru/event-booking-service/internal/config"
	"laschool.ru/event-booking-service/pkg/container"
)

const (
	DIPublisher = "notification-publisher"
	DIConsumer  = "notification-consumer"
)

func init() {
	container.Register(func(builder *container.Builder, _ map[string]interface{}) error {
		// Регистрируем Publisher
		if err := builder.Add(container.Def{
			Name: DIPublisher,
			Build: func(ctn container.Container) (interface{}, error) {
				cfg := ctn.Get(config.DIConfig).(*config.Config)
				return NewPublisher(cfg.RabbitMQ)
			},
		}); err != nil {
			return err
		}

		// Регистрируем Consumer
		return builder.Add(container.Def{
			Name: DIConsumer,
			Build: func(ctn container.Container) (interface{}, error) {
				cfg := ctn.Get(config.DIConfig).(*config.Config)
				return NewConsumer(cfg.RabbitMQ)
			},
		})
	})
}
