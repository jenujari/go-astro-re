package engine

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"local.io/go-astro-re/internal/domain"
)

type RuleEvaluator struct {
	registry domain.RuleRegistry
	workers  int
}

func NewRuleEvaluator(registry domain.RuleRegistry, workers int) RuleEvaluator {
	if workers <= 0 {
		workers = 1
	}
	return RuleEvaluator{registry: registry, workers: workers}
}

func (e RuleEvaluator) Evaluate(ctx context.Context, astrologyContext domain.AstrologyContext) ([]domain.RuleResult, []string) {
	activeRules := e.registry.ActiveRules()
	if len(activeRules) == 0 {
		return nil, nil
	}

	type job struct {
		index int
		rule  domain.Rule
	}
	type result struct {
		index int
		item  domain.RuleResult
	}

	jobCh := make(chan job)
	resultCh := make(chan result, len(activeRules))

	workerCount := e.workers
	if workerCount > len(activeRules) {
		workerCount = len(activeRules)
	}

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobCh {
				resultCh <- result{index: j.index, item: evaluateSafely(ctx, j.rule, astrologyContext)}
			}
		}()
	}

	go func() {
		defer close(jobCh)
		for i, rule := range activeRules {
			select {
			case <-ctx.Done():
				return
			case jobCh <- job{index: i, rule: rule}:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	ordered := make([]domain.RuleResult, len(activeRules))
	for item := range resultCh {
		ordered[item.index] = item.item
	}

	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Category == ordered[j].Category {
			return ordered[i].RuleID < ordered[j].RuleID
		}
		return ordered[i].Category < ordered[j].Category
	})

	partialFailures := make([]string, 0)
	for _, result := range ordered {
		if result.Status == domain.RuleExecutionError {
			partialFailures = append(partialFailures, fmt.Sprintf("%s: %s", result.RuleID, result.ErrorText))
		}
	}
	return ordered, partialFailures
}

func evaluateSafely(ctx context.Context, rule domain.Rule, astrologyContext domain.AstrologyContext) (out domain.RuleResult) {
	metadata := rule.Metadata()
	started := time.Now()

	defer func() {
		out.RuleID = metadata.ID
		out.RuleName = metadata.Name
		out.RuleVersion = metadata.Version
		out.Category = metadata.Category
		if out.Weight == 0 {
			out.Weight = 1
		}
		if out.ConfidenceMultiplier == 0 {
			out.ConfidenceMultiplier = 1
		}
		if out.WeightedScore == 0 {
			out.WeightedScore = out.RawScore * out.Weight * out.ConfidenceMultiplier
		}
		out.DurationMillis = time.Since(started).Milliseconds()

		if recovered := recover(); recovered != nil {
			out = domain.RuleResult{
				RuleID:               metadata.ID,
				RuleName:             metadata.Name,
				RuleVersion:          metadata.Version,
				Category:             metadata.Category,
				Status:               domain.RuleExecutionError,
				Matched:              false,
				Polarity:             domain.ScoreNeutral,
				Weight:               1,
				ConfidenceMultiplier: 1,
				Explanation:          "rule execution panicked and was isolated",
				FactsUsed:            []string{"panic_recovered"},
				ErrorText:            fmt.Sprintf("panic: %v", recovered),
				DurationMillis:       time.Since(started).Milliseconds(),
			}
		}
	}()

	if err := ctx.Err(); err != nil {
		return domain.RuleResult{
			RuleID:               metadata.ID,
			RuleName:             metadata.Name,
			RuleVersion:          metadata.Version,
			Category:             metadata.Category,
			Status:               domain.RuleExecutionError,
			Matched:              false,
			Polarity:             domain.ScoreNeutral,
			Weight:               1,
			ConfidenceMultiplier: 1,
			Explanation:          "rule evaluation canceled",
			FactsUsed:            []string{"context_canceled"},
			ErrorText:            err.Error(),
		}
	}

	out = rule.Evaluate(ctx, astrologyContext)
	return out
}
