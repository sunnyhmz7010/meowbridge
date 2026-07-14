package config

import (
	"testing"
	"time"
)

func TestLoadDoesNotRequireAdminPasswordAfterBootstrap(t *testing.T) {
	t.Setenv("ADMIN_PASSWORD", "")
	t.Setenv("MEOW_API_BASE_URL", "https://push.example.test")
	t.Setenv("JWT_SECRET", "jwt-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AdminPassword != "" {
		t.Fatalf("AdminPassword = %q", cfg.AdminPassword)
	}
}

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("ADMIN_PASSWORD", "secret-password")
	t.Setenv("MEOW_API_BASE_URL", "https://push.example.test")
	t.Setenv("DATABASE_PATH", "")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("JWT_SECRET", "jwt-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DatabasePath != "meowbridge.db" {
		t.Fatalf("DatabasePath = %q", cfg.DatabasePath)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if cfg.LogRetentionDays != 14 {
		t.Fatalf("LogRetentionDays = %d", cfg.LogRetentionDays)
	}
	if cfg.MeowTimeout != 10*time.Second {
		t.Fatalf("MeowTimeout = %s", cfg.MeowTimeout)
	}
}
