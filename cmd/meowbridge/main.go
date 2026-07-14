package main

import (
	"log"

	"github.com/sunnyhmz7010/meowbridge/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("meowbridge starting on %s", cfg.HTTPAddr)
}
