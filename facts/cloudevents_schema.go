package facts

const (
	RuleEvaluationRequestedType = "com.goastrore.rules.evaluation.requested.v1"
	RuleEvaluationCompletedType = "com.goastrore.rules.evaluation.completed.v1"
	RuleEvaluationDataSchema    = "https://schemas.go-astro-re.local/rules/evaluation/v1"
)

type RuleEvaluationRequest struct {
	TenantID string   `json:"tenantId,omitempty"`
	Version  string   `json:"version,omitempty"`
	Customer Customer `json:"customer"`
}

type RulePhaseMetric struct {
	Phase           string   `json:"phase"`
	DurationMillis  int64    `json:"durationMillis"`
	Cycles          uint64   `json:"cycles"`
	FiredRulesNames []string `json:"firedRulesNames"`
}

type RuleExecutionMetrics struct {
	TenantID            string            `json:"tenantId"`
	Version             string            `json:"version"`
	TotalDurationMillis int64             `json:"totalDurationMillis"`
	Phases              []RulePhaseMetric `json:"phases"`
}

type RuleEvaluationOutcome struct {
	TenantID string               `json:"tenantId"`
	Version  string               `json:"version"`
	Customer Customer             `json:"customer"`
	Outcome  ProcessOutcome       `json:"outcome"`
	Metrics  RuleExecutionMetrics `json:"metrics"`
}
