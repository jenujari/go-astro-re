package domain

import "time"

type Location struct {
	Name        string  `json:"name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	CountryCode string  `json:"country_code"`
}

type AstrologyInput struct {
	DateTime           time.Time         `json:"datetime"`
	Timezone           string            `json:"timezone"`
	Location           Location          `json:"location"`
	CalculationProfile string            `json:"calculation_profile"`
	Metadata           map[string]string `json:"metadata,omitempty"`
}

type EvaluationRequest struct {
	Input AstrologyInput `json:"input"`
}

type PlanetPosition struct {
	Planet        string `json:"planet"`
	Sign          string `json:"sign"`
	House         int    `json:"house"`
	Nakshatra     string `json:"nakshatra"`
	IsExalted     bool   `json:"is_exalted"`
	IsDebilitated bool   `json:"is_debilitated"`
	IsOwnSign     bool   `json:"is_own_sign"`
	Retrograde    bool   `json:"retrograde"`
}

type DerivedFact struct {
	Key          string   `json:"key"`
	Value        string   `json:"value,omitempty"`
	NumericValue *float64 `json:"numeric_value,omitempty"`
	BoolValue    *bool    `json:"bool_value,omitempty"`
	Source       string   `json:"source"`
}

type DerivedFacts struct {
	Items map[string]DerivedFact `json:"items"`
}

func (f DerivedFacts) String(key string) (string, bool) {
	item, ok := f.Items[key]
	if !ok || item.Value == "" {
		return "", false
	}
	return item.Value, true
}

func (f DerivedFacts) Number(key string) (float64, bool) {
	item, ok := f.Items[key]
	if !ok || item.NumericValue == nil {
		return 0, false
	}
	return *item.NumericValue, true
}

func (f DerivedFacts) Bool(key string) (bool, bool) {
	item, ok := f.Items[key]
	if !ok || item.BoolValue == nil {
		return false, false
	}
	return *item.BoolValue, true
}

type AstrologyContext struct {
	Input           AstrologyInput   `json:"input"`
	PlanetPositions []PlanetPosition `json:"planet_positions"`
	DerivedFacts    DerivedFacts     `json:"derived_facts"`
	BuilderVersion  string           `json:"builder_version"`
}

type RuleStatus string

const (
	RuleStatusActive     RuleStatus = "active"
	RuleStatusInactive   RuleStatus = "inactive"
	RuleStatusDeprecated RuleStatus = "deprecated"
)

type RuleMetadata struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Category    string     `json:"category"`
	Status      RuleStatus `json:"status"`
	Tags        []string   `json:"tags"`
	Priority    int        `json:"priority"`
	IsModifier  bool       `json:"is_modifier"`
	Description string     `json:"description"`
}

type RuleExecutionStatus string

const (
	RuleExecutionMatched    RuleExecutionStatus = "matched"
	RuleExecutionNotMatched RuleExecutionStatus = "not_matched"
	RuleExecutionError      RuleExecutionStatus = "error"
)

type ScorePolarity string

const (
	ScorePositive ScorePolarity = "positive"
	ScoreNegative ScorePolarity = "negative"
	ScoreNeutral  ScorePolarity = "neutral"
)

type RuleResult struct {
	RuleID               string              `json:"rule_id"`
	RuleName             string              `json:"rule_name"`
	RuleVersion          string              `json:"rule_version"`
	Category             string              `json:"category"`
	Status               RuleExecutionStatus `json:"status"`
	Matched              bool                `json:"matched"`
	Polarity             ScorePolarity       `json:"polarity"`
	RawScore             float64             `json:"raw_score"`
	Weight               float64             `json:"weight"`
	ConfidenceMultiplier float64             `json:"confidence_multiplier"`
	WeightedScore        float64             `json:"weighted_score"`
	Explanation          string              `json:"explanation"`
	FactsUsed            []string            `json:"facts_used"`
	ErrorText            string              `json:"error_text,omitempty"`
	DurationMillis       int64               `json:"duration_ms"`
}

type CategoryScore struct {
	Positive float64 `json:"positive"`
	Negative float64 `json:"negative"`
	Net      float64 `json:"net"`
}

type ScoreSummary struct {
	PositiveTotal  float64                  `json:"positive_total"`
	NegativeTotal  float64                  `json:"negative_total"`
	NetScore       float64                  `json:"net_score"`
	CategoryTotals map[string]CategoryScore `json:"category_totals"`
}

type EvaluationReport struct {
	RequestID       int64            `json:"request_id"`
	EvaluationID    int64            `json:"evaluation_id"`
	Input           AstrologyInput   `json:"input"`
	Context         AstrologyContext `json:"context"`
	Summary         ScoreSummary     `json:"summary"`
	RuleResults     []RuleResult     `json:"rule_results"`
	PartialFailures []string         `json:"partial_failures"`
	CreatedAt       time.Time        `json:"created_at"`
}
