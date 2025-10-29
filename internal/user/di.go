package user

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"laschool.ru/event-booking-service/internal/config"
	"laschool.ru/event-booking-service/internal/db"
	"laschool.ru/event-booking-service/pkg/container"
)

const (
	DIUserRepo    = "user-repository"
	DIUserService = "user-service"
)

func init() {
	container.Register(func(builder *container.Builder, _ map[string]interface{}) error {
		if err := builder.Add(container.Def{
			Name: DIUserRepo,
			Build: func(ctn container.Container) (interface{}, error) {
				database := ctn.Get(db.DIDatabase).(*sqlx.DB)
				return NewRepository(database), nil
			},
		}); err != nil {
			return err
		}

		return builder.Add(container.Def{
			Name: DIUserService,
			Build: func(ctn container.Container) (interface{}, error) {
				fmt.Println("Building user service")
				repo := ctn.Get(DIUserRepo).(Repository)
				cfg := ctn.Get(config.DIConfig).(*config.Config)
				return NewService(repo, cfg.JWT.Secret, cfg.JWT.TTL), nil
			},
		})
	})
}
