package rulesengine

import (
	"fmt"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/jenujari/go-astro-re/facts"
)

func parseEvaluateCustomerInput(event cloudevents.Event) (EvaluateCustomerInput, error) {
	var payload facts.RuleEvaluationRequest
	if err := event.DataAs(&payload); err != nil {
		return EvaluateCustomerInput{}, fmt.Errorf("decode request cloud event data: %w", err)
	}

	customer := payload.Customer
	return EvaluateCustomerInput{
		TenantID: strings.TrimSpace(payload.TenantID),
		Version:  strings.TrimSpace(payload.Version),
		Customer: &customer,
	}, nil
}

func (s *Service) EvaluateCustomerEvent(event cloudevents.Event) (cloudevents.Event, error) {
	input, err := parseEvaluateCustomerInput(event)
	if err != nil {
		return cloudevents.Event{}, err
	}

	result, err := s.EvaluateCustomerPhases(input)
	if err != nil {
		return cloudevents.Event{}, err
	}

	outcomePayload := toRuleEvaluationOutcome(result)
	return newRuleEvaluationCompletedEvent(event, outcomePayload)
}

func newRuleEvaluationCompletedEvent(requestEvent cloudevents.Event, payload facts.RuleEvaluationOutcome) (cloudevents.Event, error) {
	outEvent := cloudevents.NewEvent()
	outEvent.SetID(uuid.NewString())
	outEvent.SetSource(requestEvent.Source())
	if outEvent.Source() == "" {
		outEvent.SetSource("go-astro-re/rules-engine")
	}
	outEvent.SetType(facts.RuleEvaluationCompletedType)
	outEvent.SetDataSchema(facts.RuleEvaluationDataSchema)
	if requestID := strings.TrimSpace(requestEvent.ID()); requestID != "" {
		outEvent.SetExtension("causationid", requestID)
	}
	if err := outEvent.SetData(cloudevents.ApplicationJSON, payload); err != nil {
		return cloudevents.Event{}, fmt.Errorf("set cloud event outcome data: %w", err)
	}
	return outEvent, nil
}

func toRuleEvaluationOutcome(result EvaluateCustomerResult) facts.RuleEvaluationOutcome {
	phaseMetrics := make([]facts.RulePhaseMetric, 0, len(result.Metrics.Phases))
	for _, phase := range result.Metrics.Phases {
		phaseMetrics = append(phaseMetrics, facts.RulePhaseMetric{
			Phase:           phase.Phase,
			DurationMillis:  phase.Duration.Milliseconds(),
			Cycles:          phase.Cycles,
			FiredRulesNames: append([]string(nil), phase.FiredRules...),
		})
	}

	outcome := facts.RuleEvaluationOutcome{
		TenantID: result.Metrics.TenantID,
		Version:  result.Metrics.Version,
		Metrics: facts.RuleExecutionMetrics{
			TenantID:            result.Metrics.TenantID,
			Version:             result.Metrics.Version,
			TotalDurationMillis: result.Metrics.TotalDuration.Milliseconds(),
			Phases:              phaseMetrics,
		},
	}

	if result.Customer != nil {
		outcome.Customer = *result.Customer
	}
	if result.Outcome != nil {
		outcome.Outcome = *result.Outcome
	}

	return outcome
}
