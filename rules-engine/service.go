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
	Customer *facts.Customer
}

type PhaseMetrics struct {
	Phase      string        `json:"phase"`
	Duration   time.Duration `json:"duration"`
	Cycles     uint64        `json:"cycles"`
	FiredRules []string      `json:"firedRules"`
}

type ExecutionMetrics struct {
	TenantID      string         `json:"tenantId"`
	Version       string         `json:"version"`
	TotalDuration time.Duration  `json:"totalDuration"`
	Phases        []PhaseMetrics `json:"phases"`
}

type EvaluateCustomerResult struct {
	Customer *facts.Customer       `json:"customer"`
	Outcome  *facts.ProcessOutcome `json:"outcome"`
	Metrics  ExecutionMetrics      `json:"metrics"`
}

type Service struct {
	provider       KnowledgeBaseProvider
	ctxBuilder     DataContextBuilder
	executor       RuleExecutor
	defaultVersion string
}

func NewService(provider KnowledgeBaseProvider, ctxBuilder DataContextBuilder, executor RuleExecutor, defaultVersion string) *Service {
	if defaultVersion == "" {
		defaultVersion = "0.0.1"
	}
	return &Service{
		provider:       provider,
		ctxBuilder:     ctxBuilder,
		executor:       executor,
		defaultVersion: defaultVersion,
	}
}

func (s *Service) EvaluateCustomerPhases(input EvaluateCustomerInput) (EvaluateCustomerResult, error) {
	if s == nil {
		return EvaluateCustomerResult{}, fmt.Errorf("service is nil")
	}
	if input.Customer == nil {
		return EvaluateCustomerResult{}, fmt.Errorf("customer is nil")
	}

	tenantID := normalizeTenant(input.TenantID)
	version := input.Version
	if version == "" {
		version = s.defaultVersion
	}

	kb, err := s.provider.KnowledgeBase(tenantID, version)
	if err != nil {
		return EvaluateCustomerResult{}, fmt.Errorf("resolve knowledge base: %w", err)
	}

	outcome := &facts.ProcessOutcome{Status: "PENDING"}
	phaseOrder := []string{facts.PhaseTransform, facts.PhaseOutcome, facts.PhaseAfterEffects}
	phaseMetrics := make([]PhaseMetrics, 0, len(phaseOrder))

	totalStart := time.Now()
	for _, phase := range phaseOrder {
		ruleCtx := facts.RuleContext{Phase: phase, TenantID: tenantID, Version: version}
		dataCtx, err := s.ctxBuilder.BuildDataContext(input.Customer, outcome, ruleCtx)
		if err != nil {
			return EvaluateCustomerResult{}, fmt.Errorf("build data context for phase %s: %w", phase, err)
		}

		start := time.Now()
		trace, err := s.executor.Execute(dataCtx, kb)
		duration := time.Since(start)

		phaseMetrics = append(phaseMetrics, PhaseMetrics{
			Phase:      phase,
			Duration:   duration,
			Cycles:     trace.CycleCount,
			FiredRules: trace.FiredRules,
		})

		if err != nil {
			return EvaluateCustomerResult{
				Customer: input.Customer,
				Outcome:  outcome,
				Metrics: ExecutionMetrics{
					TenantID:      tenantID,
					Version:       version,
					TotalDuration: time.Since(totalStart),
					Phases:        phaseMetrics,
				},
			}, fmt.Errorf("execute phase %s: %w", phase, err)
		}
	}

	return EvaluateCustomerResult{
		Customer: input.Customer,
		Outcome:  outcome,
		Metrics: ExecutionMetrics{
			TenantID:      tenantID,
			Version:       version,
			TotalDuration: time.Since(totalStart),
			Phases:        phaseMetrics,
		},
	}, nil
}

type DefaultDataContextBuilder struct{}

func (DefaultDataContextBuilder) BuildDataContext(customer *facts.Customer, outcome *facts.ProcessOutcome, ruleCtx facts.RuleContext) (ast.IDataContext, error) {
	if customer == nil {
		return nil, fmt.Errorf("customer fact is nil")
	}
	if outcome == nil {
		return nil, fmt.Errorf("outcome fact is nil")
	}

	ctx := ast.NewDataContext()
	if err := ctx.Add("Customer", customer); err != nil {
		return nil, err
	}
	if err := ctx.Add("Outcome", outcome); err != nil {
		return nil, err
	}
	if err := ctx.Add("Context", &ruleCtx); err != nil {
		return nil, err
	}

	return ctx, nil
}
