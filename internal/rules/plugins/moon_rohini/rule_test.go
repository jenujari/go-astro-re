package moon_rohini_test

import (
	"context"
	"testing"

	"local.io/go-astro-re/internal/domain"
	moonrohini "local.io/go-astro-re/internal/rules/plugins/moon_rohini"
)

func TestRuleEvaluate(t *testing.T) {
	rule := moonrohiniRule()
	result := rule.Evaluate(context.Background(), domain.AstrologyContext{
		DerivedFacts: domain.DerivedFacts{Items: map[string]domain.DerivedFact{
			"moon_nakshatra": {Key: "moon_nakshatra", Value: "Rohini", Source: "test"},
		}},
	})
	if !result.Matched || result.WeightedScore != 3 {
		t.Fatalf("expected matched rule with +3, got %+v", result)
	}
}

func moonrohiniRule() interface {
	Evaluate(context.Context, domain.AstrologyContext) domain.RuleResult
} {
	return struct{ domain.Rule }{Rule: moonrohiniExpose{}}
}

type moonrohiniExpose struct{}

func (moonrohiniExpose) Metadata() domain.RuleMetadata { return moonrohini.RuleMetadataForTest() }
func (moonrohiniExpose) Evaluate(ctx context.Context, a domain.AstrologyContext) domain.RuleResult {
	return moonrohini.RuleForTest().Evaluate(ctx, a)
}
