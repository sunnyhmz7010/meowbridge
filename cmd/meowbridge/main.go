package main

import (
	"context"
	"log"

	"github.com/sunnyhmz7010/meowbridge/internal/config"
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

	log.Printf("meowbridge starting on %s", cfg.HTTPAddr)
}
