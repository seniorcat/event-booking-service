package booking

import (
	"github.com/jmoiron/sqlx"
	"laschool.ru/event-booking-service/internal/db"
	"laschool.ru/event-booking-service/pkg/container"
)

const (
	DIBookingRepo    = "booking-repository"
	DIBookingService = "booking-service"
)

func init() {
	container.Register(func(builder *container.Builder, _ map[string]interface{}) error {
		if err := builder.Add(container.Def{
			Name: DIBookingRepo,
			Build: func(ctn container.Container) (interface{}, error) {
				database := ctn.Get(db.DIDatabase).(*sqlx.DB)
				return NewRepository(database), nil
			},
		}); err != nil {
			return err
		}
		return builder.Add(container.Def{
			Name: DIBookingService,
			Build: func(ctn container.Container) (interface{}, error) {
				repo := ctn.Get(DIBookingRepo).(Repository)
				return NewService(repo), nil
			},
		})
	})
}
