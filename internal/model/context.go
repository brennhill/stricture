// context.go — ProjectContext for cross-file analysis state.
package model

// ProjectContext holds cross-file analysis state.
// Built once per run, shared across all rules.
type ProjectContext struct {
	Files            map[string]*UnifiedFileModel
	DependencyGraph  map[string][]string
	ReverseDeps      map[string][]string
	ModuleBoundaries map[string][]string
	TestSourceMap    map[string][]string
	// Manifest will be added in Phase 4
}
