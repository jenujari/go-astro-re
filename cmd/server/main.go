package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"local.io/go-astro-re/internal/application"
	"local.io/go-astro-re/internal/config"
	mockastro "local.io/go-astro-re/internal/infrastructure/astrology/mock"
	httpiface "local.io/go-astro-re/internal/interfaces/http"
	"local.io/go-astro-re/internal/observability"
	"local.io/go-astro-re/internal/rules"
	_ "local.io/go-astro-re/internal/rules/bootstrap"
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

	repo := postgres.NewRepository(db)
	service := application.NewEvaluationService(
		mockastro.NewBuilder(),
		rules.DefaultRegistry(),
		application.NewAggregator(),
		repo,
		logger,
		metrics,
		cfg.Engine.PersistResults,
	)

	server := httpiface.NewServer(service, logger)
	httpServer := &http.Server{
		Addr:              cfg.Server.HTTPAddr,
		Handler:           server.Routes(),
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
	}

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-rootCtx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
	}()

	logger.Info("server starting", "addr", cfg.Server.HTTPAddr, "persist_results", cfg.Engine.PersistResults, "workers", cfg.Engine.WorkerCount)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped", "at", time.Now().UTC())
}
