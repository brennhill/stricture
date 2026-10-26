// rules_select_test.go — Tests rule selection by id, config, category, and severity.
package engine

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stricture/stricture/internal/config"
	"github.com/stricture/stricture/internal/model"
)

type categorized struct {
	fakeRule
	category string
}

func (c categorized) Category() string { return c.category }

func selectRegistry() *model.RuleRegistry {
	reg := model.NewRuleRegistry()
	reg.Register(categorized{fakeRule{id: "A-1"}, "alpha"})
	reg.Register(categorized{fakeRule{id: "A-2"}, "alpha"})
	reg.Register(categorized{fakeRule{id: "B-1"}, "beta"})
	return reg
}

func ids(rules []model.Rule) []string {
	out := make([]string, 0, len(rules))
	for _, r := range rules {
		out = append(out, r.ID())
	}
	return out
}

func TestSelectRules(t *testing.T) {
	reg := selectRegistry()
	cfgWith := func(rules map[string]model.RuleConfig) *config.Config {
		c := config.Default()
		c.Rules = rules
		return c
	}
	cases := []struct {
		name      string
		cfg       *config.Config
		requested []string
		category  string
		want      []string
	}{
		{"no config selects every rule", nil, nil, "", []string{"A-1", "A-2", "B-1"}},
		{"config limits to configured rules, sorted", cfgWith(map[string]model.RuleConfig{"B-1": {}, "A-2": {}}), nil, "", []string{"A-2", "B-1"}},
		{"config entries unknown to the registry are ignored", cfgWith(map[string]model.RuleConfig{"A-1": {}, "GONE": {}}), nil, "", []string{"A-1"}},
		{"explicit rules override config", cfgWith(map[string]model.RuleConfig{"A-1": {}}), []string{"B-1"}, "", []string{"B-1"}},
		{"category filter is case-insensitive", nil, nil, " Alpha ", []string{"A-1", "A-2"}},
		{"rules set to off are dropped", cfgWith(map[string]model.RuleConfig{"A-1": {Severity: "off"}, "A-2": {}}), nil, "", []string{"A-2"}},
		{"blank requested ids are ignored", nil, []string{" ", "A-1"}, "", []string{"A-1"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SelectRules(reg, tc.cfg, tc.requested, tc.category)
			if err != nil {
				t.Fatalf("SelectRules error: %v", err)
			}
			if !reflect.DeepEqual(ids(got), tc.want) {
				t.Errorf("SelectRules = %v, want %v", ids(got), tc.want)
			}
		})
	}
}

func TestSelectRulesUnknownRule(t *testing.T) {
	_, err := SelectRules(selectRegistry(), nil, []string{"NOPE"}, "")
	if err == nil || !strings.Contains(err.Error(), `unknown rule "NOPE"`) {
		t.Fatalf("error = %v, want unknown rule", err)
	}
}

func TestSelectRulesBindsEffectiveConfig(t *testing.T) {
	cfg := config.Default()
	cfg.Rules = map[string]model.RuleConfig{
		"A-1": {Severity: "warn", Options: map[string]interface{}{"max": 5}},
		"A-2": {},
	}
	got, err := SelectRules(selectRegistry(), cfg, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	bound := got[0].(RuleWithConfig)
	if bound.Config.Severity != "warn" || bound.Config.Options["max"] != 5 {
		t.Errorf("A-1 config = %+v, want warn with max=5", bound.Config)
	}
	defaulted := got[1].(RuleWithConfig)
	if defaulted.Config.Severity != "error" || defaulted.Config.Options == nil {
		t.Errorf("A-2 config = %+v, want default severity and non-nil options", defaulted.Config)
	}
}
