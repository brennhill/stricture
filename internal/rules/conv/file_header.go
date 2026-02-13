// file_header.go — CONV-file-header: Require file header comments.
package conv

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// FileHeader requires file header comments matching a pattern.
type FileHeader struct{}

func (r *FileHeader) ID() string          { return "CONV-file-header" }
func (r *FileHeader) Category() string    { return "conv" }
func (r *FileHeader) Description() string { return "Require file header comments" }
func (r *FileHeader) Why() string {
	return "File headers provide quick context about a file's purpose."
}
func (r *FileHeader) DefaultSeverity() string   { return "error" }
func (r *FileHeader) NeedsProjectContext() bool { return false }

func (r *FileHeader) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	source := string(file.Source)
	lines := strings.Split(source, "\n")
	firstLine := ""
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		firstLine = trimmed
		break
	}

	filename := filepath.Base(file.Path)
	expectedPrefix := fmt.Sprintf("// %s — ", filename)
	if strings.HasPrefix(firstLine, expectedPrefix) {
		return nil
	}

	severity := config.Severity
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   "File missing header comment, expected format: '// {filename} — {purpose}'",
			FilePath:  file.Path,
			StartLine: 1,
			Context: &model.ViolationContext{
				SuggestedFix: fmt.Sprintf("Add header comment at line 1 with format: '// %s — {purpose}'.", filename),
			},
		},
	}
}
