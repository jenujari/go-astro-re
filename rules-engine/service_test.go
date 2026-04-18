package rulesengine

import (
	"errors"
	"testing"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/jenujari/go-astro-re/facts"
	"github.com/jenujari/go-astro-re/rules-engine/mocks"
)

func TestServiceEvaluateCustomerSuccess(t *testing.T) {
	provider := mocks.NewKnowledgeBaseProvider(t)
	ctxBuilder := mocks.NewDataContextBuilder(t)
	executor := mocks.NewRuleExecutor(t)
	service := NewService(provider, ctxBuilder, executor)

	customer := &facts.Customer{Age: 20}
	kb := &ast.KnowledgeBase{}
	ctx := ast.NewDataContext()

	provider.EXPECT().KnowledgeBase(DefaultTenantID, "0.0.1").Return(kb, nil)
	ctxBuilder.EXPECT().BuildDataContext(customer, facts.DefaultPhase).Return(ctx, nil)
	executor.EXPECT().Execute(ctx, kb).Return(nil)

	metrics, err := service.EvaluateCustomer(EvaluateCustomerInput{Customer: customer})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metrics.TenantID != DefaultTenantID {
		t.Fatalf("expected default tenant, got %s", metrics.TenantID)
	}
	if metrics.Version != "0.0.1" {
		t.Fatalf("expected default version, got %s", metrics.Version)
	}
	if metrics.Phase != facts.DefaultPhase {
		t.Fatalf("expected default phase, got %s", metrics.Phase)
	}
}

func TestServiceEvaluateCustomerExecutionError(t *testing.T) {
	provider := mocks.NewKnowledgeBaseProvider(t)
	ctxBuilder := mocks.NewDataContextBuilder(t)
	executor := mocks.NewRuleExecutor(t)
	service := NewService(provider, ctxBuilder, executor)

	customer := &facts.Customer{Age: 20}
	kb := &ast.KnowledgeBase{}
	ctx := ast.NewDataContext()
	wantErr := errors.New("boom")

	provider.EXPECT().KnowledgeBase("tenantA", "2.0.0").Return(kb, nil)
	ctxBuilder.EXPECT().BuildDataContext(customer, "PHASE2").Return(ctx, nil)
	executor.EXPECT().Execute(ctx, kb).Return(wantErr)

	metrics, err := service.EvaluateCustomer(EvaluateCustomerInput{
		TenantID: "tenantA",
		Version:  "2.0.0",
		Phase:    "PHASE2",
		Customer: customer,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if metrics.TenantID != "tenantA" || metrics.Version != "2.0.0" || metrics.Phase != "PHASE2" {
		t.Fatalf("unexpected metrics: %+v", metrics)
	}
}
