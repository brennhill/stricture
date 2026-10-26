// register.go — Registers every CTR rule with a rule registry.
package ctr

import "github.com/stricture/stricture/internal/model"

// Register adds all CTR rules to r.
func Register(r *model.RuleRegistry) {
	r.Register(&RequestShape{})
	r.Register(&ResponseShape{})
	r.Register(&StatusCodeHandling{})
	r.Register(&SharedTypeSync{})
	r.Register(&JSONTagMatch{})
	r.Register(&DualTest{})
	r.Register(&StrictnessParity{})
	r.Register(&ManifestConformance{})
}
