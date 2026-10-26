// register.go — Registers every TQ rule with a rule registry.
package tq

import "github.com/stricture/stricture/internal/model"

// Register adds all TQ rules to r.
func Register(r *model.RuleRegistry) {
	r.Register(&NoShallowAssertions{})
	r.Register(&ReturnTypeVerified{})
	r.Register(&SchemaConformance{})
	r.Register(&ErrorPathCoverage{})
	r.Register(&AssertionDepth{})
	r.Register(&BoundaryTested{})
	r.Register(&MockScope{})
	r.Register(&TestIsolation{})
	r.Register(&NegativeCases{})
	r.Register(&TestNaming{})
}
