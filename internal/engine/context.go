// context.go — ProjectContext for cross-file analysis.
package engine

import "github.com/stricture/stricture/internal/model"

// ProjectContext holds cross-file analysis state.
type ProjectContext struct {
	Files            map[string]*model.UnifiedFileModel
	DependencyGraph  map[string][]string
	ReverseDeps      map[string][]string
	ModuleBoundaries map[string][]string
	TestSourceMap    map[string][]string
	// Manifest will be added in Phase 4
}
