package facts

const (
	DefaultPhase = "PHASE1"
)

// RuleContext carries execution metadata consumed by phase-based rules.
type RuleContext struct {
	Phase string
}
