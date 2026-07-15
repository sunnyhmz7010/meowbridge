package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AdminPassword    string
	DatabasePath     string
	HTTPPort         string
	JWTSecret        string
	LogRetentionDays int
	MeowTimeout      time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		AdminPassword:    os.Getenv("ADMIN_PASSWORD"),
		DatabasePath:     "meowbridge.db",
		HTTPPort:         envOrDefault("HTTP_PORT", "8080"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		LogRetentionDays: intEnvOrDefault("LOG_RETENTION_DAYS", 14),
		MeowTimeout:      10 * time.Second,
	}
	if err := validateHTTPPort(cfg.HTTPPort); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func validateHTTPPort(port string) error {
	if strings.Contains(port, ":") {
		return errors.New("HTTP_PORT must not include ':'")
	}
	value, err := strconv.Atoi(port)
	if err != nil || value <= 0 || value > 65535 {
		return fmt.Errorf("HTTP_PORT must be a valid TCP port")
	}
	return nil
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
