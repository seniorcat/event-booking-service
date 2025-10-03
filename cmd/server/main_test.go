package main_test

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/pressly/goose/v3"
	"gopkg.in/yaml.v3"
	httprouter "laschool.ru/event-booking-service/internal/http"
	"laschool.ru/event-booking-service/internal/http/middleware"
)

var (
	db     *sql.DB
	server http.Handler
)

type TestConfig struct {
	Database struct {
		DSN string `yaml:"dsn"`
	} `yaml:"database"`
}

// loadTestConfig загружает YAML конфиг
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

// getProjectRoot возвращает путь к корню проекта
func getProjectRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to get cwd: %v", err))
	}

	return filepath.Join(cwd, "..", "..")
}

func TestMain(m *testing.M) {
	root := getProjectRoot()

	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = filepath.Join(root, "int-tests", "config.test.yaml")
	}

	cfg, err := loadTestConfig(configFile)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	db, err = sql.Open("pgx", cfg.Database.DSN)
	if err != nil {
		panic(fmt.Sprintf("db open failed: %v", err))
	}

	if err := goose.Up(db, filepath.Join(root, "deploy", "migrations")); err != nil {
		panic(fmt.Sprintf("migrations failed: %v", err))
	}

	mux := httprouter.NewRouter()
	loggingMux := middleware.LoggingMiddleware(mux)
	server = middleware.PanicMiddleware(loggingMux)

	code := m.Run()

	if err := goose.Reset(db, filepath.Join(root, "deploy", "migrations")); err != nil {
		panic(fmt.Sprintf("cleanup failed: %v", err))
	}

	os.Exit(code)
}
