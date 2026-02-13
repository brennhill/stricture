// config.go — Rule configuration helpers for tests.
package testutil

import "github.com/stricture/stricture/internal/model"

// DefaultConfig returns a RuleConfig with default severity and no options.
func DefaultConfig() model.RuleConfig {
	return model.RuleConfig{
		Severity: "error",
		Options:  map[string]interface{}{},
	}
}

// ConfigWithOption returns a RuleConfig with a single option set.
func ConfigWithOption(key string, value interface{}) model.RuleConfig {
	return model.RuleConfig{
		Severity: "error",
		Options: map[string]interface{}{
			key: value,
		},
	}
}

// ConfigWithOptions returns a RuleConfig with multiple options.
func ConfigWithOptions(opts map[string]interface{}) model.RuleConfig {
	return model.RuleConfig{
		Severity: "error",
		Options:  opts,
	}
}

// WarnConfig returns a RuleConfig with warning severity.
func WarnConfig() model.RuleConfig {
	return model.RuleConfig{
		Severity: "warn",
		Options:  map[string]interface{}{},
	}
}
