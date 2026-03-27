package moon_rohini

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
		ID:          "moon_in_rohini",
		Name:        "Moon In Rohini",
		Version:     "1.0.0",
		Category:    "lunar_strength",
		Status:      domain.RuleStatusActive,
		Tags:        []string{"moon", "nakshatra", "starter"},
		Priority:    10,
		Description: "Awards points when Moon is placed in Rohini nakshatra.",
	}
}

func (rule) Evaluate(_ context.Context, ctx domain.AstrologyContext) domain.RuleResult {
	nakshatra, ok := ctx.DerivedFacts.String("moon_nakshatra")
	if !ok || nakshatra != "Rohini" {
		return domain.RuleResult{
			Status:               domain.RuleExecutionNotMatched,
			Matched:              false,
			Polarity:             domain.ScoreNeutral,
			Weight:               1,
			ConfidenceMultiplier: 1,
			Explanation:          "Moon is not in Rohini.",
			FactsUsed:            []string{"moon_nakshatra"},
		}
	}

	return domain.RuleResult{
		Status:               domain.RuleExecutionMatched,
		Matched:              true,
		Polarity:             domain.ScorePositive,
		RawScore:             3,
		Weight:               1,
		ConfidenceMultiplier: 1,
		WeightedScore:        3,
		Explanation:          "Moon is in Rohini, so the rule adds +3.",
		FactsUsed:            []string{"moon_nakshatra=Rohini"},
	}
}
