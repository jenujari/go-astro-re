package facts

const (
	PhaseTransform    = "TRANSFORM"
	PhaseOutcome      = "OUTCOME"
	PhaseAfterEffects = "AFTER_EFFECTS"
	DefaultPhase      = PhaseTransform
)

// RuleContext carries execution metadata consumed by phase-based rules.
type RuleContext struct {
	Phase    string
	TenantID string
	Version  string
}
