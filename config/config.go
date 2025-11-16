package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server ServerConfig
	DB     DBConfig
}

type ServerConfig struct {
	Port           int    `env:"SERVER_PORT"`
	ReadTimeoutMs  int    `env:"READ_TIMEOUT_MS"`
	WriteTimeoutMs int    `env:"WRITE_TIMEOUT_MS"`
	LogLevel       string `env:"LOG_LEVEL"`
}

type DBConfig struct {
	URL             string        `env:"DATABASE_URL"`
	MigrationsPath  string        `env:"MIGRATIONS_PATH"`
	MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS"`
	MaxConnLifetime time.Duration `env:"DB_MAX_CONN_LIFETIME"`
}

func MustLoad() *Config {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("config load error: %v", err)
	}

	return &cfg
}
