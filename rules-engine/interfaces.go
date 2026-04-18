package rulesengine

import (
	"github.com/hyperjumptech/grule-rule-engine/ast"
	"github.com/jenujari/go-astro-re/facts"
)

// KnowledgeBaseProvider resolves cached knowledge base for tenant/version.
type KnowledgeBaseProvider interface {
	KnowledgeBase(tenantID string, version string) (*ast.KnowledgeBase, error)
}

// DataContextBuilder creates request-scoped data context.
type DataContextBuilder interface {
	BuildDataContext(customer *facts.Customer, outcome *facts.ProcessOutcome, ruleCtx facts.RuleContext) (ast.IDataContext, error)
}

// RuleExecutor executes rules with reusable engine.
type RuleExecutor interface {
	Execute(dataCtx ast.IDataContext, kb *ast.KnowledgeBase) (ExecutionTrace, error)
}
