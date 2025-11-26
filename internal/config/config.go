package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Server struct {
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type Database struct {
	DSN         string `yaml:"dsn"`
	AutoMigrate bool   `yaml:"auto_migrate"`
}

type JWT struct {
	Secret string        `yaml:"secret"`
	TTL    time.Duration `yaml:"ttl"`
}

type Redis struct {
	Address  string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"rdb"`

	DialTimeout  int `yaml:"dial_timeout"`
	ReadTimeout  int `yaml:"read_timeout"`
	WriteTimeout int `yaml:"write_timeout"`
	PoolTimeout  int `yaml:"pool_timeout"`

	MaxRetries      int `yaml:"max_retries"`
	MinRetryBackoff int `yaml:"min_retry_backoff"`
	MaxRetryBackoff int `yaml:"max_retry_backoff"`

	PoolSize     int `yaml:"pool_size"`
	MinIdleConns int `yaml:"min_idle_conns"`

	ConnMaxIdleTime int `yaml:"conn_max_idle_time"`
	ConnMaxLifetime int `yaml:"conn_max_lifetime"`
}

type Config struct {
	Server   Server   `yaml:"server"`
	Database Database `yaml:"database"`
	JWT      JWT      `yaml:"jwt"`
	Redis    Redis    `yaml:"redis"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LoadConfig загружает конфигурацию, читая путь из переменной окружения CONFIG_PATH
// с дефолтом на "config.yaml" в корне приложения.
func LoadConfig() (*Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = "config.yaml"
	}
	return Load(path)
}
