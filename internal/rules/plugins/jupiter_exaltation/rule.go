package jupiter_exaltation

import (
	"context"

	"local.io/go-astro-re/internal/domain"
	"local.io/go-astro-re/internal/rules"
)

type rule struct{}

func init() {
	rules.DefaultRegistry().Register(rule{})
}

func (rule) Metadata() domain.RuleMetadata {
	return domain.RuleMetadata{
		ID:          "jupiter_exalted",
		Name:        "Jupiter Exalted",
		Version:     "1.0.0",
		Category:    "planetary_strength",
		Status:      domain.RuleStatusActive,
		Tags:        []string{"jupiter", "exaltation"},
		Priority:    40,
		Description: "Awards points when Jupiter is exalted.",
	}
}

func (rule) Evaluate(_ context.Context, ctx domain.AstrologyContext) domain.RuleResult {
	exalted, ok := ctx.DerivedFacts.Bool("jupiter_exalted")
	if !ok || !exalted {
		return domain.RuleResult{
			Status:               domain.RuleExecutionNotMatched,
			Matched:              false,
			Polarity:             domain.ScoreNeutral,
			Weight:               1,
			ConfidenceMultiplier: 1,
			Explanation:          "Jupiter is not exalted.",
			FactsUsed:            []string{"jupiter_exalted=false"},
		}
	}

	return domain.RuleResult{
		Status:               domain.RuleExecutionMatched,
		Matched:              true,
		Polarity:             domain.ScorePositive,
		RawScore:             5,
		Weight:               1,
		ConfidenceMultiplier: 1,
		WeightedScore:        5,
		Explanation:          "Jupiter is exalted, so the rule adds +5.",
		FactsUsed:            []string{"jupiter_exalted=true"},
	}
}
