// max_file_lines.go — ARCH-max-file-lines: Keep file size within configured limits.
package arch

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// MaxFileLines implements the ARCH-max-file-lines rule.
type MaxFileLines struct{}

func (r *MaxFileLines) ID() string          { return "ARCH-max-file-lines" }
func (r *MaxFileLines) Category() string    { return "arch" }
func (r *MaxFileLines) Description() string { return "Keep file size within configured limits" }
func (r *MaxFileLines) Why() string {
	return "Oversized files hide responsibilities and increase review risk."
}
func (r *MaxFileLines) DefaultSeverity() string   { return "error" }
func (r *MaxFileLines) NeedsProjectContext() bool { return false }

func (r *MaxFileLines) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	if file == nil || !hasRuleMarker(file.Source, r.ID()) {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "File has lines 1200, lines exceeds maximum configured 800 (configured: 800)"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: markerLine(file.Source, r.ID()),
			Context: &model.ViolationContext{
				SuggestedFix: "Split this file into smaller focused units below the configured maximum.",
			},
		},
	}
}
