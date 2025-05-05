package config

import (
	"time"
)

type Config struct {
	Postgres struct {
		Host     string
		Port     string
		User     string
		Password string
		DBName   string
		SSLMode  string
		GRPCPort string `mapstructure:"GRPC_PORT"`
	}
	Server struct {
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

	// Server
	cfg.Server.Port = "8081"

	// Auth
	cfg.Auth.AccessTokenDuration = 15 * time.Minute
	cfg.Auth.RefreshTokenDuration = 360 * time.Hour // 15 дней
	cfg.Auth.SecretKey = "your-secret-key"
	cfg.Migrations.Enable = false
	return cfg
}
