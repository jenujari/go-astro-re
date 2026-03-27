package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"local.io/go-astro-re/internal/application"
	"local.io/go-astro-re/internal/config"
	mockastro "local.io/go-astro-re/internal/infrastructure/astrology/mock"
	cliiface "local.io/go-astro-re/internal/interfaces/cli"
	"local.io/go-astro-re/internal/observability"
	"local.io/go-astro-re/internal/rules"
	_ "local.io/go-astro-re/internal/rules/bootstrap"
	"local.io/go-astro-re/internal/storage/postgres"
)

func main() {
	configPath := "config/config.yaml"
	args := os.Args[1:]
	if len(args) >= 2 && args[0] == "-config" {
		configPath = args[1]
		args = args[2:]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := observability.NewLogger(cfg.Logging.Level)
	metrics := observability.NoopMetrics{}

	var db *sql.DB
	if cfg.Database.Enabled {
		db, err = postgres.Open(cfg.Database)
		if err != nil {
			logger.Error("open database", "error", err)
			os.Exit(1)
		}
		defer db.Close()
	}

	service := application.NewEvaluationService(
		mockastro.NewBuilder(),
		rules.DefaultRegistry(),
		application.NewAggregator(),
		postgres.NewRepository(db),
		logger,
		metrics,
		cfg.Engine.PersistResults,
	)

	runner := cliiface.NewRunner(service, os.Stdout)
	if err := runner.Run(context.Background(), args); err != nil {
		logger.Error("cli failed", "error", err)
		os.Exit(1)
	}
}
