// archtest_test.go — Proves the architecture checks catch violations (synthetic graphs).
package archtest

import (
	"strings"
	"testing"
)

var testRules = []LayerRule{
	{Pattern: "internal/model", Allowed: nil},
	{Pattern: "internal/rules/...", Allowed: []string{"internal/model"}},
	{Pattern: "internal/cli", Allowed: []string{"internal/..."}},
}

func contains(msgs []string, sub string) bool {
	for _, m := range msgs {
		if strings.Contains(m, sub) {
			return true
		}
	}
	return false
}

func TestMatch(t *testing.T) {
	cases := []struct {
		pkg, pat string
		want     bool
	}{
		{"internal/rules/tq", "internal/rules/...", true},
		{"internal/rules", "internal/rules/...", true},
		{"internal/rulesx", "internal/rules/...", false},
		{"internal/model", "internal/model", true},
		{"internal/model/sub", "internal/model", false},
	}
	for _, c := range cases {
		if got := Match(c.pkg, c.pat); got != c.want {
			t.Errorf("Match(%q,%q)=%v want %v", c.pkg, c.pat, got, c.want)
		}
	}
}

func TestCheckLayering(t *testing.T) {
	g := Graph{
		"internal/model":    {"internal/rules/tq"},
		"internal/rules/tq": {"internal/model", "internal/cli"},
		"internal/cli":      {"internal/rules/tq"},
		"internal/mystery":  nil,
	}
	got := CheckLayering(g, testRules)
	if !contains(got, "internal/model imports internal/rules/tq") {
		t.Errorf("model -> rules not flagged: %v", got)
	}
	if !contains(got, "internal/rules/tq imports internal/cli") {
		t.Errorf("rules -> cli not flagged: %v", got)
	}
	if !contains(got, "internal/mystery is not classified") {
		t.Errorf("unclassified package not flagged: %v", got)
	}
	if contains(got, "internal/cli imports") {
		t.Errorf("cli import wrongly flagged: %v", got)
	}
}

func TestCheckRuleImports(t *testing.T) {
	rulePkgs := []string{"internal/rules/tq", "internal/rules/arch"}
	g := Graph{
		"internal/rules/all": {"internal/rules/tq"},
		"internal/cli":       {"internal/rules/arch"},
		"cmd/strict":         {"internal/rules/tq"},
	}
	got := CheckRuleImports(g, rulePkgs, []string{"internal/rules/all"}, []string{"cmd/strict", "cmd/gone"})
	if !contains(got, "internal/cli imports rule package internal/rules/arch") {
		t.Errorf("direct rule import not flagged: %v", got)
	}
	if contains(got, "cmd/strict imports") {
		t.Errorf("exempt importer flagged: %v", got)
	}
	if !contains(got, "exemption cmd/gone is stale") {
		t.Errorf("stale exemption not flagged: %v", got)
	}
}

func TestCheckLineCeilings(t *testing.T) {
	lines := map[string]int{"a.go": 900, "b.go": 100, "c.go": 1000, "d.go": 700, "e.go": 850}
	ceilings := map[string]int{"c.go": 950, "d.go": 900, "e.go": 900, "missing.go": 900}
	got := CheckLineCeilings(lines, 800, ceilings)
	for _, want := range []string{
		"a.go has 900 lines, over the 800 limit",
		"c.go grew to 1000",
		"d.go is 700 lines and fits",
		"e.go shrank to 850",
		"missing.go matches no file",
	} {
		if !contains(got, want) {
			t.Errorf("missing finding %q in %v", want, got)
		}
	}
	if contains(got, "b.go") {
		t.Errorf("compliant file flagged: %v", got)
	}
}
