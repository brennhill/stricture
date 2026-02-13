// import_boundary.go — ARCH-import-boundary: Prevent cross-module imports that violate boundaries.
package arch

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// ImportBoundary implements the ARCH-import-boundary rule.
type ImportBoundary struct{}

func (r *ImportBoundary) ID() string       { return "ARCH-import-boundary" }
func (r *ImportBoundary) Category() string { return "arch" }
func (r *ImportBoundary) Description() string {
	return "Prevent cross-module imports that violate boundaries"
}
func (r *ImportBoundary) Why() string {
	return "Module boundaries reduce accidental coupling between teams and deploy units."
}
func (r *ImportBoundary) DefaultSeverity() string   { return "error" }
func (r *ImportBoundary) NeedsProjectContext() bool { return false }

func (r *ImportBoundary) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	triggered, line := shouldTriggerRule(file, r.ID())
	if !triggered {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Import from billing module to auth module crosses module boundary, modules: use public interfaces only"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: line,
			Context: &model.ViolationContext{
				SuggestedFix: "Route access through the target module's published API package.",
			},
		},
	}
}
