package rulesengine

import (
	"fmt"
	"time"

	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/jenujari/go-astro-re/facts"
)

type EvaluateCustomerInput struct {
	TenantID string
	Version  string
	Phase    string
	Customer *facts.Customer
}

type ExecutionMetrics struct {
	TenantID string        `json:"tenantId"`
	Version  string        `json:"version"`
	Phase    string        `json:"phase"`
	Duration time.Duration `json:"duration"`
}

type Service struct {
	provider   KnowledgeBaseProvider
	ctxBuilder DataContextBuilder
	executor   RuleExecutor
}

func NewService(provider KnowledgeBaseProvider, ctxBuilder DataContextBuilder, executor RuleExecutor) *Service {
	return &Service{
		provider:   provider,
		ctxBuilder: ctxBuilder,
		executor:   executor,
	}
}

func (s *Service) EvaluateCustomer(input EvaluateCustomerInput) (ExecutionMetrics, error) {
	if s == nil {
		return ExecutionMetrics{}, fmt.Errorf("service is nil")
	}
	if input.Customer == nil {
		return ExecutionMetrics{}, fmt.Errorf("customer is nil")
	}

	tenantID := normalizeTenant(input.TenantID)
	version := input.Version
	if version == "" {
		version = "0.0.1"
	}
	phase := input.Phase
	if phase == "" {
		phase = facts.DefaultPhase
	}

	kb, err := s.provider.KnowledgeBase(tenantID, version)
	if err != nil {
		return ExecutionMetrics{}, fmt.Errorf("resolve knowledge base: %w", err)
	}

	dataCtx, err := s.ctxBuilder.BuildDataContext(input.Customer, phase)
	if err != nil {
		return ExecutionMetrics{}, fmt.Errorf("build data context: %w", err)
	}

	start := time.Now()
	err = s.executor.Execute(dataCtx, kb)
	metrics := ExecutionMetrics{
		TenantID: tenantID,
		Version:  version,
		Phase:    phase,
		Duration: time.Since(start),
	}
	if err != nil {
		return metrics, fmt.Errorf("execute rules: %w", err)
	}

	return metrics, nil
}

type DefaultDataContextBuilder struct{}

func (DefaultDataContextBuilder) BuildDataContext(customer any, phase string) (ast.IDataContext, error) {
	customerFact, ok := customer.(*facts.Customer)
	if !ok || customerFact == nil {
		return nil, fmt.Errorf("customer fact must be *facts.Customer")
	}

	ctx := ast.NewDataContext()
	if err := ctx.Add("Customer", customerFact); err != nil {
		return nil, err
	}

	ruleCtx := &facts.RuleContext{Phase: phase}
	if err := ctx.Add("Context", ruleCtx); err != nil {
		return nil, err
	}

	return ctx, nil
}
