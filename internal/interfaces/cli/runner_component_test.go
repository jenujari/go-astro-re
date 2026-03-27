package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"

	"local.io/go-astro-re/internal/application"
	"local.io/go-astro-re/internal/domain"
	mockastro "local.io/go-astro-re/internal/infrastructure/astrology/mock"
	cliiface "local.io/go-astro-re/internal/interfaces/cli"
	"local.io/go-astro-re/internal/observability"
	"local.io/go-astro-re/internal/rules"
	_ "local.io/go-astro-re/internal/rules/bootstrap"
)

func TestRunnerComponent_Run(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := application.NewEvaluationService(
		mockastro.NewBuilder(),
		rules.DefaultRegistry(),
		application.NewAggregator(),
		nil,
		logger,
		observability.NoopMetrics{},
		false,
	)

	var out bytes.Buffer
	runner := cliiface.NewRunner(service, &out)
	err := runner.Run(context.Background(), []string{
		"-datetime", "2025-01-01T12:00:00Z",
		"-timezone", "Asia/Kolkata",
		"-location-name", "Delhi",
		"-country-code", "IN",
		"-lat", "28.6139",
		"-lon", "77.2090",
		"-profile", "default",
		"-workers", "2",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var report domain.EvaluationReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if len(report.RuleResults) != 5 {
		t.Fatalf("expected 5 rule results, got %d", len(report.RuleResults))
	}
}
