// engine_test.go — Baseline tests for engine package invariants.
package engine

import "testing"

func TestProjectContextZeroValue(t *testing.T) {
	var ctx ProjectContext

	if ctx.Files != nil {
		t.Fatalf("Files map must default to nil")
	}
	if ctx.DependencyGraph != nil {
		t.Fatalf("DependencyGraph map must default to nil")
	}
	if ctx.ReverseDeps != nil {
		t.Fatalf("ReverseDeps map must default to nil")
	}
	if ctx.ModuleBoundaries != nil {
		t.Fatalf("ModuleBoundaries map must default to nil")
	}
	if ctx.TestSourceMap != nil {
		t.Fatalf("TestSourceMap map must default to nil")
	}
}
