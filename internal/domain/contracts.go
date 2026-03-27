package domain

import "context"

type Rule interface {
	Metadata() RuleMetadata
	Evaluate(context.Context, AstrologyContext) RuleResult
}

type RuleRegistry interface {
	Register(Rule)
	ActiveRules() []Rule
	ListActiveMetadata() []RuleMetadata
}

type AstrologyContextBuilder interface {
	Build(context.Context, AstrologyInput) (AstrologyContext, error)
}

type Aggregator interface {
	Aggregate([]RuleResult) ScoreSummary
}

type Repository interface {
	SaveEvaluation(context.Context, EvaluationReport) (requestID int64, evaluationID int64, err error)
	GetEvaluation(context.Context, int64) (EvaluationReport, error)
}
