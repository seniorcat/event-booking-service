package event

import (
	"github.com/jmoiron/sqlx"
	"laschool.ru/event-booking-service/internal/db"
	"laschool.ru/event-booking-service/pkg/container"
)

const (
	DIEventRepo    = "event-repository"
	DIEventService = "event-service"
)

func init() {
	container.Register(func(builder *container.Builder, _ map[string]interface{}) error {
		if err := builder.Add(container.Def{
			Name: DIEventRepo,
			Build: func(ctn container.Container) (interface{}, error) {
				database := ctn.Get(db.DIDatabase).(*sqlx.DB)
				return NewRepository(database), nil
			},
		}); err != nil {
			return err
		}
		return builder.Add(container.Def{
			Name: DIEventService,
			Build: func(ctn container.Container) (interface{}, error) {
				repo := ctn.Get(DIEventRepo).(Repository)
				return NewService(repo), nil
			},
		})
	})
}
