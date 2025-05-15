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
	Logger struct {
		LogLevel    string   `yaml:"log_level"`
		Development bool     `yaml:"development"`
		Encoding    string   `yaml:"encoding"`
		OutputPaths []string `yaml:"output_paths"`
	} `yaml:"logger"`
	GRPC struct {
		Port string `yaml:"port"`
	} `yaml:"grpc"`
}

func Load() *Config {
	cfg := &Config{}

	// Postgres configuration
	cfg.Postgres.Host = "localhost"
	cfg.Postgres.Port = "5432"
	cfg.Postgres.User = "postgres"
	cfg.Postgres.Password = "postgres"
	cfg.Postgres.DBName = "PG"
	cfg.Postgres.SSLMode = "disable"
	cfg.Postgres.GRPCPort = "50052"

	// Server configuration
	cfg.Server.Port = "8081"

	// Auth configuration
	cfg.Auth.AccessTokenDuration = 15 * time.Minute
	cfg.Auth.RefreshTokenDuration = 360 * time.Hour
	cfg.Auth.SecretKey = "your-secret-key"

	// Logger configuration
	cfg.Logger = struct {
		LogLevel    string   `yaml:"log_level"`
		Development bool     `yaml:"development"`
		Encoding    string   `yaml:"encoding"`
		OutputPaths []string `yaml:"output_paths"`
	}{
		LogLevel:    "debug",
		Development: true,
		Encoding:    "console",
	}

	// GRPC configuration
	cfg.GRPC.Port = "50051"

	cfg.Migrations.Enable = false
	return cfg
}
