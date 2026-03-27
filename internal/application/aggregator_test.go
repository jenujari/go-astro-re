package application_test

import (
	"testing"

	"local.io/go-astro-re/internal/application"
	"local.io/go-astro-re/internal/domain"
)

func TestAggregatorAggregate(t *testing.T) {
	aggregator := application.NewAggregator()
	summary := aggregator.Aggregate([]domain.RuleResult{
		{Category: "strength", WeightedScore: 3, Status: domain.RuleExecutionMatched},
		{Category: "strength", WeightedScore: -2, Status: domain.RuleExecutionMatched},
		{Category: "affliction", WeightedScore: -4, Status: domain.RuleExecutionMatched},
	})

	if summary.PositiveTotal != 3 {
		t.Fatalf("expected positive total 3, got %.2f", summary.PositiveTotal)
	}
	if summary.NegativeTotal != 6 {
		t.Fatalf("expected negative total 6, got %.2f", summary.NegativeTotal)
	}
	if summary.NetScore != -3 {
		t.Fatalf("expected net score -3, got %.2f", summary.NetScore)
	}
}
