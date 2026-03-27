package sun_debilitation

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
		ID:          "sun_debilitated",
		Name:        "Sun Debilitated",
		Version:     "1.0.0",
		Category:    "planetary_affliction",
		Status:      domain.RuleStatusActive,
		Tags:        []string{"sun", "debilitation"},
		Priority:    50,
		Description: "Applies a negative score when Sun is debilitated.",
	}
}

func (rule) Evaluate(_ context.Context, ctx domain.AstrologyContext) domain.RuleResult {
	debilitated, ok := ctx.DerivedFacts.Bool("sun_debilitated")
	if !ok || !debilitated {
		return domain.RuleResult{
			Status:               domain.RuleExecutionNotMatched,
			Matched:              false,
			Polarity:             domain.ScoreNeutral,
			Weight:               1,
			ConfidenceMultiplier: 1,
			Explanation:          "Sun is not debilitated.",
			FactsUsed:            []string{"sun_debilitated=false"},
		}
	}

	return domain.RuleResult{
		Status:               domain.RuleExecutionMatched,
		Matched:              true,
		Polarity:             domain.ScoreNegative,
		RawScore:             -3,
		Weight:               1,
		ConfidenceMultiplier: 1,
		WeightedScore:        -3,
		Explanation:          "Sun is debilitated, so the rule subtracts 3 points.",
		FactsUsed:            []string{"sun_debilitated=true"},
	}
}
