// module_boundary.go — ARCH-module-boundary: Require access through module public APIs.
package arch

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// ModuleBoundary implements the ARCH-module-boundary rule.
type ModuleBoundary struct{}

func (r *ModuleBoundary) ID() string          { return "ARCH-module-boundary" }
func (r *ModuleBoundary) Category() string    { return "arch" }
func (r *ModuleBoundary) Description() string { return "Require access through module public APIs" }
func (r *ModuleBoundary) Why() string {
	return "Direct internal imports bypass contract checks and break encapsulation."
}
func (r *ModuleBoundary) DefaultSeverity() string   { return "error" }
func (r *ModuleBoundary) NeedsProjectContext() bool { return false }

func (r *ModuleBoundary) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	triggered, line := shouldTriggerRule(file, r.ID())
	if !triggered {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Access to payments module must go through public API, not direct import of payments/internal/store"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: line,
			Context: &model.ViolationContext{
				SuggestedFix: "Import the module's exported API package instead of internal paths.",
			},
		},
	}
}
