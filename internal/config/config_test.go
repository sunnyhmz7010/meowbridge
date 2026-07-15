package config

import (
	"testing"
	"time"
)

func TestLoadDoesNotRequireAdminPasswordAfterBootstrap(t *testing.T) {
	t.Setenv("ADMIN_PASSWORD", "")
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
	t.Setenv("DATABASE_PATH", "")
	t.Setenv("HTTP_ADDR", ":9000")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("JWT_SECRET", "jwt-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPPort != "8080" {
		t.Fatalf("HTTPPort = %q", cfg.HTTPPort)
	}
	if cfg.LogRetentionDays != 14 {
		t.Fatalf("LogRetentionDays = %d", cfg.LogRetentionDays)
	}
	if cfg.MeowTimeout != 10*time.Second {
		t.Fatalf("MeowTimeout = %s", cfg.MeowTimeout)
	}
}

func TestLoadUsesHTTPPortWithoutColon(t *testing.T) {
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("JWT_SECRET", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.HTTPPort != "9090" {
		t.Fatalf("HTTPPort = %q", cfg.HTTPPort)
	}
}

func TestLoadRejectsHTTPPortWithColon(t *testing.T) {
	t.Setenv("HTTP_PORT", ":9090")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil")
	}
}
