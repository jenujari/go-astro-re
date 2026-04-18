package rulesengine_test

import (
	"errors"
	"testing"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/jenujari/go-astro-re/facts"
	rulesengine "github.com/jenujari/go-astro-re/rules-engine"
	"github.com/jenujari/go-astro-re/rules-engine/mocks"
	"github.com/stretchr/testify/mock"
)

func TestServiceEvaluateCustomerPhasesSuccess(t *testing.T) {
	provider := mocks.NewKnowledgeBaseProvider(t)
	ctxBuilder := mocks.NewDataContextBuilder(t)
	executor := mocks.NewRuleExecutor(t)
	service := rulesengine.NewService(provider, ctxBuilder, executor, "0.0.1")

	customer := &facts.Customer{Age: 20}
	kb := &ast.KnowledgeBase{}
	ctx := ast.NewDataContext()

	provider.EXPECT().KnowledgeBase("default", "0.0.1").Return(kb, nil)
	ctxBuilder.EXPECT().BuildDataContext(customer, mock.AnythingOfType("*facts.ProcessOutcome"), facts.RuleContext{Phase: facts.PhaseTransform, TenantID: "default", Version: "0.0.1"}).Return(ctx, nil)
	ctxBuilder.EXPECT().BuildDataContext(customer, mock.AnythingOfType("*facts.ProcessOutcome"), facts.RuleContext{Phase: facts.PhaseOutcome, TenantID: "default", Version: "0.0.1"}).Return(ctx, nil)
	ctxBuilder.EXPECT().BuildDataContext(customer, mock.AnythingOfType("*facts.ProcessOutcome"), facts.RuleContext{Phase: facts.PhaseAfterEffects, TenantID: "default", Version: "0.0.1"}).Return(ctx, nil)
	executor.EXPECT().Execute(ctx, kb).Return(rulesengine.ExecutionTrace{FiredRules: []string{"R1"}, CycleCount: 1}, nil).Times(3)

	result, err := service.EvaluateCustomerPhases(rulesengine.EvaluateCustomerInput{Customer: customer})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Metrics.TenantID != "default" {
		t.Fatalf("expected default tenant, got %s", result.Metrics.TenantID)
	}
	if result.Metrics.Version != "0.0.1" {
		t.Fatalf("expected default version, got %s", result.Metrics.Version)
	}
	if len(result.Metrics.Phases) != 3 {
		t.Fatalf("expected 3 phase metrics, got %d", len(result.Metrics.Phases))
	}
}

func TestServiceEvaluateCustomerPhasesExecutionError(t *testing.T) {
	provider := mocks.NewKnowledgeBaseProvider(t)
	ctxBuilder := mocks.NewDataContextBuilder(t)
	executor := mocks.NewRuleExecutor(t)
	service := rulesengine.NewService(provider, ctxBuilder, executor, "0.0.1")

	customer := &facts.Customer{Age: 20}
	kb := &ast.KnowledgeBase{}
	ctx := ast.NewDataContext()
	wantErr := errors.New("boom")

	provider.EXPECT().KnowledgeBase("tenantA", "2.0.0").Return(kb, nil)
	ctxBuilder.EXPECT().BuildDataContext(customer, mock.AnythingOfType("*facts.ProcessOutcome"), facts.RuleContext{Phase: facts.PhaseTransform, TenantID: "tenantA", Version: "2.0.0"}).Return(ctx, nil)
	executor.EXPECT().Execute(ctx, kb).Return(rulesengine.ExecutionTrace{FiredRules: []string{"R1"}, CycleCount: 1}, wantErr)

	result, err := service.EvaluateCustomerPhases(rulesengine.EvaluateCustomerInput{
		TenantID: "tenantA",
		Version:  "2.0.0",
		Customer: customer,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if result.Metrics.TenantID != "tenantA" || result.Metrics.Version != "2.0.0" {
		t.Fatalf("unexpected metrics: %+v", result.Metrics)
	}
	if len(result.Metrics.Phases) != 1 {
		t.Fatalf("expected partial phase metrics, got %d", len(result.Metrics.Phases))
	}
}
