package rules

import (
	"sort"
	"sync"

	"local.io/go-astro-re/internal/domain"
)

type Registry struct {
	mu    sync.RWMutex
	rules map[string]domain.Rule
}

var defaultRegistry = &Registry{rules: map[string]domain.Rule{}}

func DefaultRegistry() *Registry {
	return defaultRegistry
}

func (r *Registry) Register(rule domain.Rule) {
	metadata := rule.Metadata()
	if metadata.ID == "" {
		panic("rule id is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.rules[metadata.ID]; exists {
		panic("rule already registered: " + metadata.ID)
	}
	r.rules[metadata.ID] = rule
}

func (r *Registry) ActiveRules() []domain.Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]domain.Rule, 0, len(r.rules))
	for _, rule := range r.rules {
		if rule.Metadata().Status == domain.RuleStatusActive {
			items = append(items, rule)
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		mi := items[i].Metadata()
		mj := items[j].Metadata()
		if mi.Priority == mj.Priority {
			return mi.ID < mj.ID
		}
		return mi.Priority < mj.Priority
	})

	return items
}

func (r *Registry) ListActiveMetadata() []domain.RuleMetadata {
	activeRules := r.ActiveRules()
	metadata := make([]domain.RuleMetadata, 0, len(activeRules))
	for _, rule := range activeRules {
		metadata = append(metadata, rule.Metadata())
	}
	return metadata
}
