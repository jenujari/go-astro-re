package saturn_aspect_moon

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
		ID:          "saturn_aspect_on_moon",
		Name:        "Saturn Aspect On Moon",
		Version:     "1.0.0",
		Category:    "lunar_affliction",
		Status:      domain.RuleStatusActive,
		Tags:        []string{"saturn", "moon", "aspect"},
		Priority:    20,
		Description: "Applies a negative score when Saturn aspects the Moon.",
	}
}

func (rule) Evaluate(_ context.Context, ctx domain.AstrologyContext) domain.RuleResult {
	aspects, ok := ctx.DerivedFacts.Bool("saturn_aspects_moon")
	if !ok || !aspects {
		return domain.RuleResult{
			Status:               domain.RuleExecutionNotMatched,
			Matched:              false,
			Polarity:             domain.ScoreNeutral,
			Weight:               1,
			ConfidenceMultiplier: 1,
			Explanation:          "Saturn does not aspect the Moon in the mocked context.",
			FactsUsed:            []string{"saturn_aspects_moon=false"},
		}
	}

	return domain.RuleResult{
		Status:               domain.RuleExecutionMatched,
		Matched:              true,
		Polarity:             domain.ScoreNegative,
		RawScore:             -4,
		Weight:               1,
		ConfidenceMultiplier: 1,
		WeightedScore:        -4,
		Explanation:          "Saturn aspects the Moon, so the rule subtracts 4 points.",
		FactsUsed:            []string{"saturn_aspects_moon=true"},
	}
}
