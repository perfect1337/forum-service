package config

import (
	"time"
)

// Добавляем явное объявление структуры
type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	GRPCPort string
}

type Config struct {
	Postgres PostgresConfig // Теперь используем явный тип
	Server   struct {
		Port string
	}
	Auth struct {
		AccessTokenDuration  time.Duration
		RefreshTokenDuration time.Duration
		SecretKey            string
	}

	Migrations struct {
		Enable bool
	}
}

func Load() *Config {
	cfg := &Config{}

	// Postgres
	cfg.Postgres.Host = "localhost"
	cfg.Postgres.Port = "5432"
	cfg.Postgres.User = "postgres"
	cfg.Postgres.Password = "postgres"
	cfg.Postgres.DBName = "PG"
	cfg.Postgres.SSLMode = "disable"
	cfg.Postgres.GRPCPort = "50052"

	// Server
	cfg.Server.Port = "8081"

	// Auth
	cfg.Auth.AccessTokenDuration = 15 * time.Minute
	cfg.Auth.RefreshTokenDuration = 360 * time.Hour
	cfg.Auth.SecretKey = "your-secret-key"
	cfg.Migrations.Enable = false
	return cfg
}
