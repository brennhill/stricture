// register_test.go — Verifies Register adds every TQ rule exactly once.
package tq

import (
	"strings"
	"testing"

	"github.com/stricture/stricture/internal/model"
)

func TestRegisterAddsAllRules(t *testing.T) {
	reg := model.NewRuleRegistry()
	Register(reg)

	rules := reg.All()
	if len(rules) != 10 {
		t.Fatalf("Register added %d rules, want 10", len(rules))
	}
	seen := map[string]bool{}
	for _, r := range rules {
		if seen[r.ID()] {
			t.Errorf("rule %s registered twice", r.ID())
		}
		seen[r.ID()] = true
		if !strings.HasPrefix(r.ID(), "TQ-") {
			t.Errorf("rule %s is not a TQ rule", r.ID())
		}
	}
}
