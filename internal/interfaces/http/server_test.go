package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"local.io/go-astro-re/internal/domain"
	httpiface "local.io/go-astro-re/internal/interfaces/http"
)

type serviceStub struct {
	report domain.EvaluationReport
}

func (s serviceStub) Evaluate(context.Context, domain.AstrologyInput, int) (domain.EvaluationReport, error) {
	return s.report, nil
}

func (s serviceStub) GetEvaluation(context.Context, int64) (domain.EvaluationReport, error) {
	return s.report, nil
}

func (s serviceStub) ListActiveRules() []domain.RuleMetadata {
	return []domain.RuleMetadata{{ID: "r1", Name: "Rule 1", Status: domain.RuleStatusActive}}
}

func TestEvaluateEndpoint(t *testing.T) {
	server := httpiface.NewServer(serviceStub{
		report: domain.EvaluationReport{EvaluationID: 1, Summary: domain.ScoreSummary{NetScore: 3}},
	}, slog.Default())

	payload := map[string]any{
		"datetime": "2025-01-01T12:00:00Z",
		"timezone": "Asia/Kolkata",
		"location": map[string]any{
			"name": "Delhi", "latitude": 28.61, "longitude": 77.20, "country_code": "IN",
		},
		"calculation_profile": "default",
		"worker_count":        3,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/v1/evaluations", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	server.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestActiveRulesEndpoint(t *testing.T) {
	server := httpiface.NewServer(serviceStub{
		report: domain.EvaluationReport{CreatedAt: time.Now().UTC()},
	}, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/v1/rules/active", nil)
	rr := httptest.NewRecorder()
	server.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
