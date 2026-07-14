package main

import (
	"context"
	"log"
	"net/http"

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

	meowAPIBaseURL := cfg.MeowAPIBaseURL
	if meowAPIBaseURL == "" {
		meowAPIBaseURL, err = st.GetSetting(ctx, "meow_api_base_url")
		if err != nil {
			log.Fatal(err)
		}
	}
	meowClient := meow.New(meowAPIBaseURL, cfg.MeowTimeout)
	router := httpapi.NewRouter(httpapi.Dependencies{
		Store:      st,
		Config:     cfg,
		MeowClient: meowClient,
	})
	log.Printf("meowbridge starting on %s", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, router); err != nil {
		log.Fatal(err)
	}
}
