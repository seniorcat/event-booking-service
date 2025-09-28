package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	"laschool.ru/event-booking-service/internal/config"
	"laschool.ru/event-booking-service/internal/db"
	httprouter "laschool.ru/event-booking-service/internal/http"
	"laschool.ru/event-booking-service/internal/http/middleware"
	di "laschool.ru/event-booking-service/pkg/container"
)

func main() {
	fmt.Println("Booking Service started...")

	// загружаем конфиг
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// при необходимости прогоняем миграции (если флаг в конфиге)
	if cfg.Database.AutoMigrate {
		ctn, err := di.Instance(nil, nil)
		if err != nil {
			log.Fatalf("di init failed: %v", err)
		}
		database := ctn.Get(db.DIDatabase).(*sqlx.DB)
		if err := db.RunMigrations(context.Background(), database, "deploy/migrations"); err != nil {
			log.Fatalf("migrations failed: %v", err)
		}
	}

	// маршруты
	mux := httprouter.NewRouter()
	// логирование сервера
	loggingMux := middleware.LoggingMiddleware(mux)
	muxWithLogAndPanic := middleware.PanicMiddleware(loggingMux)

	// старт сервера
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      muxWithLogAndPanic,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	log.Printf("server listening on %d...", cfg.Server.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
