// file_naming.go — CONV-file-naming: Enforce file naming convention.
package conv

import (
	"github.com/stricture/stricture/internal/model"
)

// FileNaming enforces file naming conventions (kebab-case, snake_case, etc.).
type FileNaming struct{}

func (r *FileNaming) ID() string             { return "CONV-file-naming" }
func (r *FileNaming) Category() string       { return "conv" }
func (r *FileNaming) Description() string    { return "Enforce file naming convention" }
func (r *FileNaming) Why() string            { return "Consistent naming makes files predictable and searchable." }
func (r *FileNaming) DefaultSeverity() string { return "error" }
func (r *FileNaming) NeedsProjectContext() bool { return false }

func (r *FileNaming) Check(file *model.UnifiedFileModel, context *model.ProjectContext, config model.RuleConfig) []model.Violation {
	// TODO: Implement — currently returns nil (tests will fail, TDD red phase)
	return nil
}
