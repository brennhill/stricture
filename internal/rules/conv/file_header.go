// file_header.go — CONV-file-header: Require file header comments.
package conv

import (
	"github.com/stricture/stricture/internal/model"
)

// FileHeader requires file header comments matching a pattern.
type FileHeader struct{}

func (r *FileHeader) ID() string             { return "CONV-file-header" }
func (r *FileHeader) Category() string       { return "conv" }
func (r *FileHeader) Description() string    { return "Require file header comments" }
func (r *FileHeader) Why() string            { return "File headers provide quick context about a file's purpose." }
func (r *FileHeader) DefaultSeverity() string { return "error" }
func (r *FileHeader) NeedsProjectContext() bool { return false }

func (r *FileHeader) Check(file *model.UnifiedFileModel, context *model.ProjectContext, config model.RuleConfig) []model.Violation {
	// TODO: Implement — currently returns nil (tests will fail, TDD red phase)
	return nil
}
