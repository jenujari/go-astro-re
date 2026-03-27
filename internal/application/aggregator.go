package application

import "local.io/go-astro-re/internal/domain"

type Aggregator struct{}

func NewAggregator() Aggregator {
	return Aggregator{}
}

func (Aggregator) Aggregate(results []domain.RuleResult) domain.ScoreSummary {
	summary := domain.ScoreSummary{
		CategoryTotals: make(map[string]domain.CategoryScore),
	}

	for _, result := range results {
		if result.Status == domain.RuleExecutionError {
			continue
		}

		category := summary.CategoryTotals[result.Category]
		score := result.WeightedScore
		if score > 0 {
			summary.PositiveTotal += score
			category.Positive += score
		}
		if score < 0 {
			negative := -score
			summary.NegativeTotal += negative
			category.Negative += negative
		}

		category.Net = category.Positive - category.Negative
		summary.CategoryTotals[result.Category] = category
	}

	summary.NetScore = summary.PositiveTotal - summary.NegativeTotal
	return summary
}
