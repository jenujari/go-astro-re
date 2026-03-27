package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"local.io/go-astro-re/internal/application"
	"local.io/go-astro-re/internal/domain"
	mockastro "local.io/go-astro-re/internal/infrastructure/astrology/mock"
	httpiface "local.io/go-astro-re/internal/interfaces/http"
	"local.io/go-astro-re/internal/observability"
	"local.io/go-astro-re/internal/rules"
	_ "local.io/go-astro-re/internal/rules/bootstrap"
)

func TestServerComponent_EvaluateFlow(t *testing.T) {
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
	server := httpiface.NewServer(service, logger)

	payload := map[string]any{
		"datetime":            "2025-01-01T12:00:00Z",
		"timezone":            "Asia/Kolkata",
		"calculation_profile": "default",
		"worker_count":        4,
		"location": map[string]any{
			"name":         "Delhi",
			"latitude":     28.6139,
			"longitude":    77.2090,
			"country_code": "IN",
		},
		"metadata": map[string]string{
			"source": "component-test",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/evaluations", bytes.NewReader(body)).WithContext(context.Background())
	rr := httptest.NewRecorder()

	server.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var report domain.EvaluationReport
	if err := json.Unmarshal(rr.Body.Bytes(), &report); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if len(report.RuleResults) != 5 {
		t.Fatalf("expected 5 rule results, got %d", len(report.RuleResults))
	}
	if report.Summary.CategoryTotals == nil {
		t.Fatal("expected category totals to be present")
	}
}
