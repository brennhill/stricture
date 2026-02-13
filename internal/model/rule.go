// rule.go — Rule interface, RuleConfig, and RuleRegistry.
package model

// Rule defines the interface all lint rules must implement.
type Rule interface {
	// ID returns the unique rule identifier (e.g., "CONV-file-naming").
	ID() string

	// Category returns the rule category (CONV, ARCH, TQ, CTR).
	Category() string

	// Description returns a human-readable description of the rule.
	Description() string

	// DefaultSeverity returns the default severity level (error, warn, off).
	DefaultSeverity() string

	// Check evaluates the rule against a file model and returns violations.
	// The context parameter may be nil for rules that don't need cross-file analysis.
	Check(file *UnifiedFileModel, context *ProjectContext, config RuleConfig) []Violation

	// NeedsProjectContext returns true if this rule requires project-level context.
	NeedsProjectContext() bool

	// Why returns a brief explanation of why this rule matters.
	Why() string
}

// RuleConfig holds configuration for a specific rule instance.
type RuleConfig struct {
	Severity string
	Options  map[string]interface{}
}

// RuleRegistry holds all registered rules.
type RuleRegistry struct {
	rules []Rule
}

// NewRuleRegistry creates a new rule registry.
func NewRuleRegistry() *RuleRegistry {
	return &RuleRegistry{}
}

// Register adds a rule to the registry.
func (r *RuleRegistry) Register(rule Rule) {
	r.rules = append(r.rules, rule)
}

// All returns all registered rules.
func (r *RuleRegistry) All() []Rule {
	return r.rules
}

// ByID returns a rule by its ID.
func (r *RuleRegistry) ByID(id string) (Rule, bool) {
	for _, rule := range r.rules {
		if rule.ID() == id {
			return rule, true
		}
	}
	return nil, false
}

// ByCategory returns all rules in a category.
func (r *RuleRegistry) ByCategory(category string) []Rule {
	var result []Rule
	for _, rule := range r.rules {
		if rule.Category() == category {
			result = append(result, rule)
		}
	}
	return result
}
