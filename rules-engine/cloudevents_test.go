package rulesengine_test

import (
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/jenujari/go-astro-re/facts"
	rulesengine "github.com/jenujari/go-astro-re/rules-engine"
	"github.com/jenujari/go-astro-re/rules-engine/mocks"
	"github.com/stretchr/testify/mock"
)

func TestEvaluateCustomerEvent(t *testing.T) {
	provider := mocks.NewKnowledgeBaseProvider(t)
	ctxBuilder := mocks.NewDataContextBuilder(t)
	executor := mocks.NewRuleExecutor(t)
	service := rulesengine.NewService(provider, ctxBuilder, executor, "0.0.1")

	kb := &ast.KnowledgeBase{}
	ctx := ast.NewDataContext()
	trace := rulesengine.ExecutionTrace{FiredRules: []string{"AdultRule"}, CycleCount: 1}

	provider.EXPECT().KnowledgeBase("default", "0.0.1").Return(kb, nil)
	ctxBuilder.EXPECT().BuildDataContext(mock.AnythingOfType("*facts.Customer"), mock.AnythingOfType("*facts.ProcessOutcome"), facts.RuleContext{Phase: facts.PhaseTransform, TenantID: "default", Version: "0.0.1"}).Return(ctx, nil)
	ctxBuilder.EXPECT().BuildDataContext(mock.AnythingOfType("*facts.Customer"), mock.AnythingOfType("*facts.ProcessOutcome"), facts.RuleContext{Phase: facts.PhaseOutcome, TenantID: "default", Version: "0.0.1"}).Return(ctx, nil)
	ctxBuilder.EXPECT().BuildDataContext(mock.AnythingOfType("*facts.Customer"), mock.AnythingOfType("*facts.ProcessOutcome"), facts.RuleContext{Phase: facts.PhaseAfterEffects, TenantID: "default", Version: "0.0.1"}).Return(ctx, nil)
	executor.EXPECT().Execute(ctx, kb).Return(trace, nil).Times(3)

	requestEvent := cloudevents.NewEvent()
	requestEvent.SetID("req-1")
	requestEvent.SetSource("go-astro-re/test")
	requestEvent.SetType(facts.RuleEvaluationRequestedType)
	if err := requestEvent.SetData(cloudevents.ApplicationJSON, facts.RuleEvaluationRequest{
		Customer: facts.Customer{Age: 20},
	}); err != nil {
		t.Fatalf("set request event data: %v", err)
	}

	resultEvent, err := service.EvaluateCustomerEvent(requestEvent)
	if err != nil {
		t.Fatalf("evaluate customer event: %v", err)
	}

	if resultEvent.Type() != facts.RuleEvaluationCompletedType {
		t.Fatalf("unexpected event type: %s", resultEvent.Type())
	}

	var payload facts.RuleEvaluationOutcome
	if err := resultEvent.DataAs(&payload); err != nil {
		t.Fatalf("decode result payload: %v", err)
	}
	if payload.TenantID != "default" || payload.Version != "0.0.1" {
		t.Fatalf("unexpected payload scope: tenant=%s version=%s", payload.TenantID, payload.Version)
	}
	if len(payload.Metrics.Phases) != 3 {
		t.Fatalf("expected 3 phase metrics, got %d", len(payload.Metrics.Phases))
	}
}
