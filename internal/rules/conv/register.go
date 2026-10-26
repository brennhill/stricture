// register.go — Registers every CONV rule with a rule registry.
package conv

import "github.com/stricture/stricture/internal/model"

// Register adds all CONV rules to r.
func Register(r *model.RuleRegistry) {
	r.Register(&FileNaming{})
	r.Register(&FileHeader{})
	r.Register(&ErrorFormat{})
	r.Register(&ExportNaming{})
	r.Register(&TestFileLocation{})
	r.Register(&RequiredExports{})
}
