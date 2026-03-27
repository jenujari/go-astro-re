package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"local.io/go-astro-re/internal/domain"
	"local.io/go-astro-re/internal/engine"
	"local.io/go-astro-re/internal/observability"
)

var ErrPersistenceDisabled = errors.New("persistence disabled")

type EvaluationService struct {
	builder    domain.AstrologyContextBuilder
	registry   domain.RuleRegistry
	aggregator domain.Aggregator
	repo       domain.Repository
	logger     *slog.Logger
	metrics    observability.Metrics
	persist    bool
}

func NewEvaluationService(
	builder domain.AstrologyContextBuilder,
	registry domain.RuleRegistry,
	aggregator domain.Aggregator,
	repo domain.Repository,
	logger *slog.Logger,
	metrics observability.Metrics,
	persist bool,
) EvaluationService {
	return EvaluationService{
		builder:    builder,
		registry:   registry,
		aggregator: aggregator,
		repo:       repo,
		logger:     logger,
		metrics:    metrics,
		persist:    persist,
	}
}

func (s EvaluationService) Evaluate(ctx context.Context, input domain.AstrologyInput, workerCount int) (domain.EvaluationReport, error) {
	if input.DateTime.IsZero() {
		return domain.EvaluationReport{}, fmt.Errorf("datetime is required")
	}
	if input.Timezone == "" {
		return domain.EvaluationReport{}, fmt.Errorf("timezone is required")
	}
	if input.CalculationProfile == "" {
		input.CalculationProfile = "default"
	}

	started := time.Now()
	contextData, err := s.builder.Build(ctx, input)
	if err != nil {
		return domain.EvaluationReport{}, fmt.Errorf("build context: %w", err)
	}

	ruleResults, partialFailures := engine.NewRuleEvaluator(s.registry, workerCount).Evaluate(ctx, contextData)
	summary := s.aggregator.Aggregate(ruleResults)

	report := domain.EvaluationReport{
		Input:           input,
		Context:         contextData,
		Summary:         summary,
		RuleResults:     ruleResults,
		PartialFailures: partialFailures,
		CreatedAt:       time.Now().UTC(),
	}

	if s.persist && s.repo != nil {
		requestID, evaluationID, saveErr := s.repo.SaveEvaluation(ctx, report)
		if saveErr != nil {
			return domain.EvaluationReport{}, fmt.Errorf("persist evaluation: %w", saveErr)
		}
		report.RequestID = requestID
		report.EvaluationID = evaluationID
	}

	s.logger.Info("evaluation completed",
		"duration_ms", time.Since(started).Milliseconds(),
		"net_score", report.Summary.NetScore,
		"partial_failures", len(report.PartialFailures),
	)
	s.metrics.IncCounter("evaluations_total", map[string]string{"profile": input.CalculationProfile})
	s.metrics.ObserveDuration("evaluation_duration", time.Since(started), map[string]string{"profile": input.CalculationProfile})

	return report, nil
}

func (s EvaluationService) GetEvaluation(ctx context.Context, evaluationID int64) (domain.EvaluationReport, error) {
	if !s.persist || s.repo == nil {
		return domain.EvaluationReport{}, ErrPersistenceDisabled
	}
	return s.repo.GetEvaluation(ctx, evaluationID)
}

func (s EvaluationService) ListActiveRules() []domain.RuleMetadata {
	return s.registry.ListActiveMetadata()
}
