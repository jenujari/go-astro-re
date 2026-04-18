package rulesengine

import (
	"context"
	"sync"

	"github.com/hyperjumptech/grule-rule-engine/ast"
)

// ExecutionTrace captures per-run engine activity.
type ExecutionTrace struct {
	FiredRules []string
	CycleCount uint64
}

type ruleTraceListener struct {
	mu        sync.Mutex
	fired     []string
	maxCycles uint64
}

func newRuleTraceListener() *ruleTraceListener {
	return &ruleTraceListener{fired: make([]string, 0)}
}

func (l *ruleTraceListener) EvaluateRuleEntry(_ context.Context, _ uint64, _ *ast.RuleEntry, _ bool) {
}

func (l *ruleTraceListener) ExecuteRuleEntry(_ context.Context, _ uint64, entry *ast.RuleEntry) {
	if entry == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fired = append(l.fired, entry.RuleName)
}

func (l *ruleTraceListener) BeginCycle(_ context.Context, cycle uint64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if cycle > l.maxCycles {
		l.maxCycles = cycle
	}
}

func (l *ruleTraceListener) Trace() ExecutionTrace {
	l.mu.Lock()
	defer l.mu.Unlock()

	cloned := make([]string, len(l.fired))
	copy(cloned, l.fired)
	return ExecutionTrace{FiredRules: cloned, CycleCount: l.maxCycles}
}
