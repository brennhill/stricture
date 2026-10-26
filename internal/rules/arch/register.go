// register.go — Registers every ARCH rule with a rule registry.
package arch

import "github.com/stricture/stricture/internal/model"

// Register adds all ARCH rules to r.
func Register(r *model.RuleRegistry) {
	r.Register(&DependencyDirection{})
	r.Register(&ImportBoundary{})
	r.Register(&NoCircularDeps{})
	r.Register(&MaxFileLines{})
	r.Register(&LayerViolation{})
	r.Register(&ModuleBoundary{})
}
