// all_test.go — Verifies the built-in registry contains every rule exactly once.
package all

import (
	"strings"
	"testing"
)

func TestRegistryHasAllBuiltInRules(t *testing.T) {
	reg := Registry()
	counts := map[string]int{}
	seen := map[string]bool{}
	for _, r := range reg.All() {
		if seen[r.ID()] {
			t.Errorf("rule %s registered twice", r.ID())
		}
		seen[r.ID()] = true
		counts[strings.ToLower(r.Category())]++
	}
	want := map[string]int{"conv": 6, "arch": 6, "tq": 10, "ctr": 8}
	for category, n := range want {
		if counts[category] != n {
			t.Errorf("category %s has %d rules, want %d", category, counts[category], n)
		}
	}
}
