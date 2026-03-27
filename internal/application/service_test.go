package application_test

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"local.io/go-astro-re/internal/application"
	"local.io/go-astro-re/internal/domain"
	"local.io/go-astro-re/internal/observability"
)

type builderStub struct {
	ctx domain.AstrologyContext
}

func (b builderStub) Build(context.Context, domain.AstrologyInput) (domain.AstrologyContext, error) {
	return b.ctx, nil
}

type registryStub struct {
	rules []domain.Rule
}

func (r registryStub) Register(domain.Rule)                      {}
func (r registryStub) ActiveRules() []domain.Rule                { return r.rules }
func (r registryStub) ListActiveMetadata() []domain.RuleMetadata { return nil }

type ruleStub struct {
	metadata domain.RuleMetadata
	result   domain.RuleResult
}

func (r ruleStub) Metadata() domain.RuleMetadata { return r.metadata }
func (r ruleStub) Evaluate(context.Context, domain.AstrologyContext) domain.RuleResult {
	return r.result
}

type repoStub struct {
	mu     sync.Mutex
	called bool
}

func (r *repoStub) SaveEvaluation(context.Context, domain.EvaluationReport) (int64, int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.called = true
	return 10, 20, nil
}

func (r *repoStub) GetEvaluation(context.Context, int64) (domain.EvaluationReport, error) {
	return domain.EvaluationReport{}, nil
}

func TestServiceEvaluate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := &repoStub{}
	service := application.NewEvaluationService(
		builderStub{ctx: domain.AstrologyContext{
			Input: domain.AstrologyInput{DateTime: time.Now().UTC(), Timezone: "UTC", CalculationProfile: "default"},
		}},
		registryStub{rules: []domain.Rule{
			ruleStub{
				metadata: domain.RuleMetadata{ID: "r1", Name: "Rule 1", Version: "1", Category: "strength", Status: domain.RuleStatusActive},
				result:   domain.RuleResult{Status: domain.RuleExecutionMatched, Matched: true, WeightedScore: 3, Weight: 1, ConfidenceMultiplier: 1},
			},
		}},
		application.NewAggregator(),
		repo,
		logger,
		observability.NoopMetrics{},
		true,
	)

	report, err := service.Evaluate(context.Background(), domain.AstrologyInput{
		DateTime:           time.Now().UTC(),
		Timezone:           "UTC",
		CalculationProfile: "default",
	}, 2)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if report.EvaluationID != 20 {
		t.Fatalf("expected evaluation id 20, got %d", report.EvaluationID)
	}
}

func TestServiceSkipsPersistence(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := &repoStub{}
	service := application.NewEvaluationService(
		builderStub{ctx: domain.AstrologyContext{
			Input: domain.AstrologyInput{DateTime: time.Now().UTC(), Timezone: "UTC", CalculationProfile: "default"},
		}},
		registryStub{},
		application.NewAggregator(),
		repo,
		logger,
		observability.NoopMetrics{},
		false,
	)

	_, err := service.Evaluate(context.Background(), domain.AstrologyInput{
		DateTime:           time.Now().UTC(),
		Timezone:           "UTC",
		CalculationProfile: "default",
	}, 1)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if repo.called {
		t.Fatal("expected repository not to be called")
	}
}
