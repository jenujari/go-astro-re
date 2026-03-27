package main

import (
	"context"
	"log"
	"os"

	"local.io/go-astro-re/internal/config"
	"local.io/go-astro-re/internal/storage/postgres"
)

func main() {
	configPath := "config/config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := postgres.Open(cfg.Database)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := postgres.ApplyMigrations(context.Background(), db, "sql/migrations"); err != nil {
		log.Fatalf("apply migrations: %v", err)
	}

	log.Println("migrations applied")
}
