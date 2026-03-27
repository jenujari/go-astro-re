package engine_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"local.io/go-astro-re/internal/domain"
	"local.io/go-astro-re/internal/engine"
)

type registryStub struct {
	rules []domain.Rule
}

func (r registryStub) Register(domain.Rule)                      {}
func (r registryStub) ActiveRules() []domain.Rule                { return r.rules }
func (r registryStub) ListActiveMetadata() []domain.RuleMetadata { return nil }

type testRule struct {
	metadata domain.RuleMetadata
	eval     func(context.Context, domain.AstrologyContext) domain.RuleResult
}

func (r testRule) Metadata() domain.RuleMetadata { return r.metadata }
func (r testRule) Evaluate(ctx context.Context, astro domain.AstrologyContext) domain.RuleResult {
	return r.eval(ctx, astro)
}

func TestRuleEvaluatorDeterministicAndPanicSafe(t *testing.T) {
	evaluator := engine.NewRuleEvaluator(registryStub{rules: []domain.Rule{
		testRule{
			metadata: domain.RuleMetadata{ID: "b", Name: "B", Version: "1", Category: "z", Status: domain.RuleStatusActive},
			eval: func(context.Context, domain.AstrologyContext) domain.RuleResult {
				return domain.RuleResult{Status: domain.RuleExecutionMatched, Matched: true, WeightedScore: 2, Weight: 1, ConfidenceMultiplier: 1}
			},
		},
		testRule{
			metadata: domain.RuleMetadata{ID: "a", Name: "A", Version: "1", Category: "a", Status: domain.RuleStatusActive},
			eval: func(context.Context, domain.AstrologyContext) domain.RuleResult {
				panic("boom")
			},
		},
	}}, 2)

	results, failures := evaluator.Evaluate(context.Background(), domain.AstrologyContext{})
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].RuleID != "a" {
		t.Fatalf("expected sorted results, got first rule %s", results[0].RuleID)
	}
	if len(failures) != 1 {
		t.Fatalf("expected 1 partial failure, got %d", len(failures))
	}
}

func TestRuleEvaluatorUsesBoundedWorkers(t *testing.T) {
	var concurrent int32
	var maxConcurrent int32
	rules := make([]domain.Rule, 0, 6)
	for i := 0; i < 6; i++ {
		rules = append(rules, testRule{
			metadata: domain.RuleMetadata{ID: string(rune('a' + i)), Name: "rule", Version: "1", Category: "x", Status: domain.RuleStatusActive},
			eval: func(context.Context, domain.AstrologyContext) domain.RuleResult {
				current := atomic.AddInt32(&concurrent, 1)
				for {
					old := atomic.LoadInt32(&maxConcurrent)
					if current <= old || atomic.CompareAndSwapInt32(&maxConcurrent, old, current) {
						break
					}
				}
				time.Sleep(10 * time.Millisecond)
				atomic.AddInt32(&concurrent, -1)
				return domain.RuleResult{Status: domain.RuleExecutionMatched, Matched: true, Weight: 1, ConfidenceMultiplier: 1}
			},
		})
	}

	evaluator := engine.NewRuleEvaluator(registryStub{rules: rules}, 2)
	_, _ = evaluator.Evaluate(context.Background(), domain.AstrologyContext{})
	if atomic.LoadInt32(&maxConcurrent) > 2 {
		t.Fatalf("expected max concurrency <= 2, got %d", maxConcurrent)
	}
}
