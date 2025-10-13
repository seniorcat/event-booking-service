package user

import (
	"github.com/jmoiron/sqlx"
	"laschool.ru/event-booking-service/internal/db"
	"laschool.ru/event-booking-service/pkg/container"
)

const (
	DIUserRepo = "user-repository"
)

func init() {
	container.Register(func(builder *container.Builder, _ map[string]interface{}) error {
		return builder.Add(container.Def{
			Name: DIUserRepo,
			Build: func(ctn container.Container) (interface{}, error) {
				database := ctn.Get(db.DIDatabase).(*sqlx.DB)
				return NewRepository(database), nil
			},
		})
	})
}
