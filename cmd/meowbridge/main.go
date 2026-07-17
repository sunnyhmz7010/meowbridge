package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/sunnyhmz7010/meowbridge/internal/config"
	"github.com/sunnyhmz7010/meowbridge/internal/httpapi"
	"github.com/sunnyhmz7010/meowbridge/internal/meow"
	"github.com/sunnyhmz7010/meowbridge/internal/store"
	"github.com/sunnyhmz7010/meowbridge/internal/token"
)

const meowAPIBaseURL = "https://api.chuckfang.com"

func main() {
	// 初始化结构化日志
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

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
		LogRetentionDays: cfg.LogRetentionDays,
	}); err != nil {
		log.Fatal(err)
	}
	jwtSecret, err := ensureJWTSecret(ctx, st, cfg.JWTSecret)
	if err != nil {
		log.Fatal(err)
	}
	cfg.JWTSecret = jwtSecret

	meowClient := newMeowClient(cfg.MeowTimeout)
	router := httpapi.NewRouter(httpapi.Dependencies{
		Store:      st,
		Config:     cfg,
		MeowClient: meowClient,
	})
	addr := httpAddrFromPort(cfg.HTTPPort)
	log.Printf("meowbridge starting on %s", addr)
	if err := newHTTPServer(addr, router).ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

type settingStore interface {
	GetSetting(ctx context.Context, key string) (string, error)
	SetSetting(ctx context.Context, key, value string) error
}

func newMeowClient(timeout time.Duration) *meow.Client {
	return meow.New(meowAPIBaseURL, timeout)
}

func httpAddrFromPort(port string) string {
	return ":" + port
}

func ensureJWTSecret(ctx context.Context, st settingStore, configured string) (string, error) {
	if configured != "" {
		if err := st.SetSetting(ctx, "jwt_secret", configured); err != nil {
			return "", err
		}
		return configured, nil
	}
	persisted, err := st.GetSetting(ctx, "jwt_secret")
	if err == nil {
		return persisted, nil
	}
	if err != store.ErrNotFound {
		return "", err
	}
	generated, err := token.Generate()
	if err != nil {
		return "", err
	}
	if err := st.SetSetting(ctx, "jwt_secret", generated); err != nil {
		return "", err
	}
	return generated, nil
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
