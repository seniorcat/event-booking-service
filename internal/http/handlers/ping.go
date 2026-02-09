package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/sarulabs/di/v2"
)

type HealthController struct {
	container di.Container
}

func NewHealthController(ctn di.Container) *HealthController {
	return &HealthController{container: ctn}
}

func (h *HealthController) getDatabase() (*sqlx.DB, error) {
	dbInstance, err := h.container.SafeGet("db")
	if err != nil {
		return nil, err
	}

	db, ok := dbInstance.(*sqlx.DB)
	if !ok {
		return nil, err
	}

	return db, nil
}

func (h *HealthController) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	db, err := h.getDatabase()
	if err != nil {
		h.writeError(w, "database error: "+err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		h.writeError(w, "database ping failed:"+err.Error())
	}
	h.writeSuccess(w)
}

func (h *HealthController) writeError(w http.ResponseWriter, errorMsg string) {
	w.WriteHeader(http.StatusServiceUnavailable)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "not ready",
		"error":     errorMsg,
		"service":   "event-booking",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HealthController) writeSuccess(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ready",
		"service":   "event-booking",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"checks": map[string]string{
			"database": "connected",
		},
	})
}

func (h *HealthController) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "healthy",
		"service":   "event-booking",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// PingHandler godoc
// @Summary      Проверка доступности
// @Description  Возвращает pong, если сервис работает
// @Tags         health
// @Produce      plain
// @Success      200  "pong"
// @Router       /ping [get]
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
