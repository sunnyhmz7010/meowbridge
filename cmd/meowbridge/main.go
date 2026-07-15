package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/sunnyhmz7010/meowbridge/internal/config"
	"github.com/sunnyhmz7010/meowbridge/internal/httpapi"
	"github.com/sunnyhmz7010/meowbridge/internal/meow"
	"github.com/sunnyhmz7010/meowbridge/internal/store"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	st, err := store.Open(ctx, cfg.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	if err := st.Migrate(ctx); err != nil {
		log.Fatal(err)
	}
	if err := st.Bootstrap(ctx, store.BootstrapOptions{
		AdminPassword:    cfg.AdminPassword,
		MeowAPIBaseURL:   cfg.MeowAPIBaseURL,
		LogRetentionDays: cfg.LogRetentionDays,
	}); err != nil {
		log.Fatal(err)
	}

	meowClient := newMeowClient(st, cfg.MeowTimeout)
	router := httpapi.NewRouter(httpapi.Dependencies{
		Store:      st,
		Config:     cfg,
		MeowClient: meowClient,
	})
	log.Printf("meowbridge starting on %s", cfg.HTTPAddr)
	if err := newHTTPServer(cfg.HTTPAddr, router).ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

type meowSettingStore interface {
	GetSetting(ctx context.Context, key string) (string, error)
}

func newMeowClient(st meowSettingStore, timeout time.Duration) *meow.Client {
	return meow.NewWithBaseURLProvider(func(ctx context.Context) (string, error) {
		return st.GetSetting(ctx, "meow_api_base_url")
	}, timeout)
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
