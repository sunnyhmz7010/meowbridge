package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sunnyhmz7010/meowbridge/internal/meow"
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

func TestNewMeowClientUsesPersistedSettingAfterBootstrap(t *testing.T) {
	ctx := context.Background()
	firstCalls := 0
	firstServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		firstCalls++
		w.WriteHeader(http.StatusOK)
	}))
	defer firstServer.Close()
	secondCalls := 0
	secondServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secondCalls++
		w.WriteHeader(http.StatusOK)
	}))
	defer secondServer.Close()

	st, err := store.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = st.Close() }()
	if err := st.Migrate(ctx); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", MeowAPIBaseURL: firstServer.URL, LogRetentionDays: 14}); err != nil {
		t.Fatalf("Bootstrap first: %v", err)
	}
	if err := st.Bootstrap(ctx, store.BootstrapOptions{MeowAPIBaseURL: secondServer.URL, LogRetentionDays: 14}); err != nil {
		t.Fatalf("Bootstrap second: %v", err)
	}

	client := newMeowClient(st, time.Second)
	if _, err := client.Push(ctx, meow.PushRequest{Nickname: "sunny", Title: "test", Msg: "persisted", MsgType: "text"}); err != nil {
		t.Fatalf("Push: %v", err)
	}
	if firstCalls != 1 || secondCalls != 0 {
		t.Fatalf("push calls: persisted = %d, env = %d", firstCalls, secondCalls)
	}
}
