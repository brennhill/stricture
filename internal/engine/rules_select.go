// rules_select.go — Selects rules for a run and binds each to its effective configuration.
package engine

import (
	"fmt"
	"sort"
	"strings"

	"github.com/stricture/stricture/internal/config"
	"github.com/stricture/stricture/internal/model"
)

// RuleWithConfig pairs a rule with the configuration it should run under.
type RuleWithConfig struct {
	model.Rule
	Config model.RuleConfig
}

// SelectRules picks the rules to run (explicit ids, else configured rules, else all), optionally limited
// to one category, and binds each to its effective severity and options. Rules set to "off" are dropped.
func SelectRules(registry *model.RuleRegistry, cfg *config.Config, requestedRules []string, category string) ([]model.Rule, error) {
	requested, err := requestedRuleIDs(registry, requestedRules)
	if err != nil {
		return nil, err
	}
	targetCategory := strings.ToLower(strings.TrimSpace(category))

	selected := make([]model.Rule, 0)
	for _, r := range candidateRules(registry, cfg, requested) {
		if targetCategory != "" && strings.ToLower(r.Category()) != targetCategory {
			continue
		}
		ruleCfg := effectiveConfig(r, cfg)
		if strings.EqualFold(ruleCfg.Severity, "off") {
			continue
		}
		selected = append(selected, RuleWithConfig{Rule: r, Config: ruleCfg})
	}
	return selected, nil
}

// requestedRuleIDs validates explicitly requested ids against the registry.
func requestedRuleIDs(registry *model.RuleRegistry, requested []string) (map[string]bool, error) {
	ids := map[string]bool{}
	for _, raw := range requested {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		if _, ok := registry.ByID(id); !ok {
			return nil, fmt.Errorf("unknown rule %q", id)
		}
		ids[id] = true
	}
	return ids, nil
}

// candidateRules returns requested rules if any, else the configured rules, else every registered rule.
func candidateRules(registry *model.RuleRegistry, cfg *config.Config, requested map[string]bool) []model.Rule {
	switch {
	case len(requested) > 0:
		return rulesByID(registry, sortedKeys(requested))
	case cfg != nil && len(cfg.Rules) > 0:
		return rulesByID(registry, sortedKeys(cfg.Rules))
	default:
		return registry.All()
	}
}

// rulesByID resolves ids in order, skipping ids the registry does not know.
func rulesByID(registry *model.RuleRegistry, ids []string) []model.Rule {
	rules := make([]model.Rule, 0, len(ids))
	for _, id := range ids {
		if r, ok := registry.ByID(id); ok {
			rules = append(rules, r)
		}
	}
	return rules
}

// effectiveConfig applies a rule's default severity and any override from the config file.
func effectiveConfig(r model.Rule, cfg *config.Config) model.RuleConfig {
	ruleCfg := model.RuleConfig{
		Severity: r.DefaultSeverity(),
		Options:  map[string]interface{}{},
	}
	if cfg == nil {
		return ruleCfg
	}
	override, ok := cfg.Rules[r.ID()]
	if !ok {
		return ruleCfg
	}
	if strings.TrimSpace(override.Severity) != "" {
		ruleCfg.Severity = override.Severity
	}
	if override.Options != nil {
		ruleCfg.Options = override.Options
	}
	return ruleCfg
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
