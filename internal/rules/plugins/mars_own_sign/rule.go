package mars_own_sign

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
		ID:          "mars_in_own_sign",
		Name:        "Mars In Own Sign",
		Version:     "1.0.0",
		Category:    "planetary_strength",
		Status:      domain.RuleStatusActive,
		Tags:        []string{"mars", "own-sign"},
		Priority:    30,
		Description: "Awards points when Mars is in its own sign.",
	}
}

func (rule) Evaluate(_ context.Context, ctx domain.AstrologyContext) domain.RuleResult {
	ownSign, ok := ctx.DerivedFacts.Bool("mars_own_sign")
	if !ok || !ownSign {
		return domain.RuleResult{
			Status:               domain.RuleExecutionNotMatched,
			Matched:              false,
			Polarity:             domain.ScoreNeutral,
			Weight:               1,
			ConfidenceMultiplier: 1,
			Explanation:          "Mars is not in its own sign.",
			FactsUsed:            []string{"mars_own_sign=false"},
		}
	}

	return domain.RuleResult{
		Status:               domain.RuleExecutionMatched,
		Matched:              true,
		Polarity:             domain.ScorePositive,
		RawScore:             2,
		Weight:               1,
		ConfidenceMultiplier: 1,
		WeightedScore:        2,
		Explanation:          "Mars is in its own sign, so the rule adds +2.",
		FactsUsed:            []string{"mars_own_sign=true"},
	}
}
