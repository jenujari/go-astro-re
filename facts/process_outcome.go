package facts

// ProcessOutcome carries outcome state produced by rule phases.
type ProcessOutcome struct {
	Status              string
	Eligible            bool
	Message             string
	AfterEffectsApplied bool
}
