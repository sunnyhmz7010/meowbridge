package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/sunnyhmz7010/meowbridge/internal/store"
)

func TestNewHTTPServerConfiguresTimeouts(t *testing.T) {
	handler := http.NewServeMux()
	server := newHTTPServer(":8080", handler)

	if server.Addr != ":8080" || server.Handler != handler {
		t.Fatalf("server address or handler was not configured")
	}
	if server.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("ReadHeaderTimeout = %s", server.ReadHeaderTimeout)
	}
	if server.ReadTimeout != 15*time.Second {
		t.Fatalf("ReadTimeout = %s", server.ReadTimeout)
	}
	if server.WriteTimeout != 15*time.Second {
		t.Fatalf("WriteTimeout = %s", server.WriteTimeout)
	}
	if server.IdleTimeout != 60*time.Second {
		t.Fatalf("IdleTimeout = %s", server.IdleTimeout)
	}
}

func TestHTTPAddrFromPortAddsColon(t *testing.T) {
	if got := httpAddrFromPort("9090"); got != ":9090" {
		t.Fatalf("httpAddrFromPort() = %q", got)
	}
}

func TestEnsureJWTSecretUsesProvidedValueAndPersistsIt(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = st.Close() }()
	if err := st.Migrate(ctx); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	secret, err := ensureJWTSecret(ctx, st, "provided-secret")
	if err != nil {
		t.Fatalf("ensureJWTSecret: %v", err)
	}
	if secret != "provided-secret" {
		t.Fatalf("secret = %q", secret)
	}
	stored, err := st.GetSetting(ctx, "jwt_secret")
	if err != nil || stored != "provided-secret" {
		t.Fatalf("stored jwt_secret = %q, %v", stored, err)
	}
}

func TestEnsureJWTSecretGeneratesAndReusesPersistedValue(t *testing.T) {
	ctx := context.Background()
	st, err := store.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = st.Close() }()
	if err := st.Migrate(ctx); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	first, err := ensureJWTSecret(ctx, st, "")
	if err != nil {
		t.Fatalf("ensureJWTSecret first: %v", err)
	}
	if len(first) < 32 {
		t.Fatalf("generated secret too short: %q", first)
	}
	second, err := ensureJWTSecret(ctx, st, "")
	if err != nil {
		t.Fatalf("ensureJWTSecret second: %v", err)
	}
	if second != first {
		t.Fatalf("second secret = %q, want %q", second, first)
	}
}
