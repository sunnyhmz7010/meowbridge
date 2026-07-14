package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AdminPassword    string
	DatabasePath     string
	HTTPAddr         string
	JWTSecret        string
	MeowAPIBaseURL   string
	LogRetentionDays int
	MeowTimeout      time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		AdminPassword:    os.Getenv("ADMIN_PASSWORD"),
		DatabasePath:     envOrDefault("DATABASE_PATH", "meowbridge.db"),
		HTTPAddr:         envOrDefault("HTTP_ADDR", ":8080"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		MeowAPIBaseURL:   os.Getenv("MEOW_API_BASE_URL"),
		LogRetentionDays: intEnvOrDefault("LOG_RETENTION_DAYS", 14),
		MeowTimeout:      10 * time.Second,
	}
	if cfg.AdminPassword == "" {
		return Config{}, errors.New("ADMIN_PASSWORD is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}
	if cfg.MeowAPIBaseURL == "" {
		return Config{}, errors.New("MEOW_API_BASE_URL is required")
	}
	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func intEnvOrDefault(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
