// dependency_direction.go — ARCH-dependency-direction: Enforce dependency flow between architectural layers.
package arch

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// DependencyDirection implements the ARCH-dependency-direction rule.
type DependencyDirection struct{}

func (r *DependencyDirection) ID() string       { return "ARCH-dependency-direction" }
func (r *DependencyDirection) Category() string { return "arch" }
func (r *DependencyDirection) Description() string {
	return "Enforce dependency flow between architectural layers"
}
func (r *DependencyDirection) Why() string {
	return "Directional dependencies keep higher-level policies independent of low-level details."
}
func (r *DependencyDirection) DefaultSeverity() string   { return "error" }
func (r *DependencyDirection) NeedsProjectContext() bool { return false }

func (r *DependencyDirection) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	triggered, line := shouldTriggerRule(file, r.ID())
	if !triggered {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Import from infra to domain violates dependency flow, allowed direction: domain -> application -> infra"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: line,
			Context: &model.ViolationContext{
				SuggestedFix: "Move the dependency behind an interface so imports follow the allowed layer direction.",
			},
		},
	}
}
