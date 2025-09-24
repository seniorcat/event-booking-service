package db

import (
	"context"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"laschool.ru/event-booking-service/internal/config"
	"laschool.ru/event-booking-service/pkg/container"
)

const DIDatabase = "db"

func init() {
	container.Register(func(builder *container.Builder, _ map[string]interface{}) error {
		return builder.Add(container.Def{
			Name: DIDatabase,
			Build: func(ctn container.Container) (interface{}, error) {
				cfg := ctn.Get(config.DIConfig).(*config.Config)

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				db, err := sqlx.Open("pgx", cfg.Database.DSN)
				if err != nil {
					return nil, err
				}
				// Проверим соединение
				if err := db.PingContext(ctx); err != nil {
					return nil, err
				}
				return db, nil
			},
		})
	})
}
