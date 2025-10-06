package main_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	httprouter "laschool.ru/event-booking-service/internal/http"
	"laschool.ru/event-booking-service/internal/http/middleware"
	"laschool.ru/event-booking-service/pkg/container"
)

var (
	dbConn *sql.DB
	server http.Handler
)

type Server struct {
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type TestConfig struct {
	Database struct {
		DSN string `yaml:"dsn"`
	} `yaml:"database"`
	Server Server `yaml:"server"`
}

// loadTestConfig читает YAML конфиг
func loadTestConfig(path string) (*TestConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg TestConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func initTestDI(configFile string) (*container.Container, error) {
	os.Setenv("CONFIG_PATH", configFile)

	ctn, err := container.Instance(nil, nil) // возвращает di.Container
	if err != nil {
		return nil, err
	}

	return &ctn, nil // возвращаем указатель на структуру
}

func TestMain(m *testing.M) {
	configFile := filepath.Join("..", "..", "int-tests", "config.test.yaml")

	// читаем тестовый конфиг для миграций
	cfg, err := loadTestConfig(configFile)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// подключаемся к базе
	dbConn, err = sql.Open("pgx", cfg.Database.DSN)
	if err != nil {
		panic(fmt.Sprintf("db open failed: %v", err))
	}

	// запускаем миграции
	if err := goose.Up(dbConn, filepath.Join("..", "..", "deploy", "migrations")); err != nil {
		panic(fmt.Sprintf("migrations failed: %v", err))
	}

	// инициализация DI
	ctn, err := initTestDI(configFile)
	if err != nil {
		panic(fmt.Sprintf("di init failed: %v", err))
	}

	// собираем маршруты
	mux := httprouter.NewRouter()
	loggingMux := middleware.LoggingMiddleware(mux)
	server = middleware.PanicMiddleware(loggingMux)

	// запуск тестов
	code := m.Run()

	// откат миграций
	if err := goose.Reset(dbConn, filepath.Join("..", "..", "deploy", "migrations")); err != nil {
		panic(fmt.Sprintf("cleanup failed: %v", err))
	}

	// удаляем контейнер
	ctn.DeleteWithSubContainers()

	os.Exit(code)
}

func doRequest(t *testing.T, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		err := json.NewEncoder(&buf).Encode(body)
		require.NoError(t, err)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	return w
}

func createEvent(t *testing.T, capacity int) int64 {
	resp := doRequest(t, "POST", "/events", map[string]any{
		"title":       "Test event",
		"description": "some desc",
		"location":    "online",
		"capacity":    capacity,
		"starts_at":   "2025-10-01T10:00:00Z",
		"ends_at":     "2025-10-01T12:00:00Z",
	})
	require.Equal(t, http.StatusCreated, resp.Code)

	var data struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&data))
	return data.ID
}

func createBooking(t *testing.T, eventID int64, seats int) *httptest.ResponseRecorder {
	return doRequest(t, "POST", "/bookings", map[string]any{
		"event_id": eventID,
		"user_id":  1,
		"seats":    seats,
	})
}

// =======================
// 🔹 Тесты
// =======================

func TestEventCRUD(t *testing.T) {
	// Create
	eventID := createEvent(t, 10)

	// Get
	resp := doRequest(t, "GET", fmt.Sprintf("/events/%d", eventID), nil)
	require.Equal(t, http.StatusOK, resp.Code)

	// Update
	resp = doRequest(t, "PUT", fmt.Sprintf("/events/%d", eventID), map[string]any{
		"title":       "Updated title",
		"description": "new desc",
		"location":    "offline",
		"capacity":    20,
		"starts_at":   "2025-10-01T11:00:00Z",
		"ends_at":     "2025-10-01T13:00:00Z",
	})
	require.Equal(t, http.StatusNoContent, resp.Code)

	// Delete
	resp = doRequest(t, "DELETE", fmt.Sprintf("/events/%d", eventID), nil)
	require.Equal(t, http.StatusNoContent, resp.Code)

	// Ensure deleted
	resp = doRequest(t, "GET", fmt.Sprintf("/events/%d", eventID), nil)
	require.Equal(t, http.StatusNotFound, resp.Code)
}

func TestBookingCapacity(t *testing.T) {
	eventID := createEvent(t, 5)

	// ok: 3 места
	resp := createBooking(t, eventID, 3)
	require.Equal(t, http.StatusCreated, resp.Code)

	// ok: ещё 2 места
	resp = createBooking(t, eventID, 2)
	require.Equal(t, http.StatusCreated, resp.Code)

	// fail: превышаем лимит
	resp = createBooking(t, eventID, 1)
	require.Equal(t, http.StatusBadRequest, resp.Code)
}
