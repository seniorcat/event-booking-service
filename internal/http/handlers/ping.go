package handlers

import (
	"context"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"laschool.ru/event-booking-service/internal/config"
)

// PingHandler godoc
// @Summary      Проверка доступности
// @Description  Возвращает pong, если сервис работает
// @Tags         health
// @Produce      plain
// @Success      200  "pong"
// @Router       /ping [get]
func PingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

// HealthHandler godoc
// @Summary      Проверка состояния сервиса
// @Description  Проверяет состояние конфигурации и базы данных
// @Tags         health
// @Produce      plain
// @Success      200  "ok"
// @Failure      503  "config load failed | db open failed"
// @Router       /health [get]
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfig()
	if err != nil {
		http.Error(w, "config load failed", http.StatusServiceUnavailable)
		return
	}
	database, err := sqlx.Open("pgx", cfg.Database.DSN)
	if err != nil {
		http.Error(w, "db open failed", http.StatusServiceUnavailable)
		return
	}
	defer database.Close()
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := database.PingContext(ctx); err != nil {
		http.Error(w, "db not ready", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
