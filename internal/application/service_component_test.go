package application_test

import (
	"context"
	"io"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"local.io/go-astro-re/internal/application"
	"local.io/go-astro-re/internal/domain"
	mockastro "local.io/go-astro-re/internal/infrastructure/astrology/mock"
	"local.io/go-astro-re/internal/observability"
	"local.io/go-astro-re/internal/rules"
	_ "local.io/go-astro-re/internal/rules/bootstrap"
)

func TestEvaluationServiceComponent_DeterministicEvaluation(t *testing.T) {
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

	input := domain.AstrologyInput{
		DateTime: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		Timezone: "Asia/Kolkata",
		Location: domain.Location{
			Name:        "Delhi",
			Latitude:    28.6139,
			Longitude:   77.2090,
			CountryCode: "IN",
		},
		CalculationProfile: "default",
		Metadata:           map[string]string{"source": "component-test"},
	}

	reportA, err := service.Evaluate(context.Background(), input, 3)
	if err != nil {
		t.Fatalf("Evaluate() first call error = %v", err)
	}
	reportB, err := service.Evaluate(context.Background(), input, 3)
	if err != nil {
		t.Fatalf("Evaluate() second call error = %v", err)
	}

	if len(reportA.RuleResults) != 5 {
		t.Fatalf("expected 5 rule results, got %d", len(reportA.RuleResults))
	}
	if !reflect.DeepEqual(reportA.Summary, reportB.Summary) {
		t.Fatalf("expected deterministic summary, got %+v and %+v", reportA.Summary, reportB.Summary)
	}
	if len(reportA.PartialFailures) != 0 {
		t.Fatalf("expected no partial failures, got %v", reportA.PartialFailures)
	}
}
