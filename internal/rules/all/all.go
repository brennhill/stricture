// all.go — Builds the registry of every built-in rule; the only place that wires rule packages together.
package all

import (
	"github.com/stricture/stricture/internal/model"
	"github.com/stricture/stricture/internal/rules/arch"
	"github.com/stricture/stricture/internal/rules/conv"
	"github.com/stricture/stricture/internal/rules/ctr"
	"github.com/stricture/stricture/internal/rules/tq"
)

// Registry returns a RuleRegistry containing all built-in rules.
func Registry() *model.RuleRegistry {
	r := model.NewRuleRegistry()
	conv.Register(r)
	arch.Register(r)
	tq.Register(r)
	ctr.Register(r)
	return r
}
