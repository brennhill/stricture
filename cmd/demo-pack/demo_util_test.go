// demo_util_test.go — Pins deterministic parent inheritance for the demo pack.
package main

import "testing"

func TestEnsureParentNodesFromSubsystemsIsDeterministic(t *testing.T) {
	for i := 0; i < 50; i++ {
		nodes := map[string]*nodeMeta{
			"shop:b": {ID: "shop:b", Escalation: "pagerduty:b"},
			"shop:a": {ID: "shop:a", Escalation: "slack:a"},
		}
		ensureParentNodesFromSubsystems(nodes)
		if got := nodes["shop"].Escalation; got != "slack:a" {
			t.Fatalf("run %d: parent escalation = %q, want slack:a (first child by ID)", i, got)
		}
	}
}
