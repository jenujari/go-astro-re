package facts

// ProcessOutcome carries outcome state produced by rule phases.
type ProcessOutcome struct {
	Status              string `json:"status"`
	Eligible            bool   `json:"eligible"`
	Message             string `json:"message"`
	AfterEffectsApplied bool   `json:"afterEffectsApplied"`
}
