// no_circular_deps.go — ARCH-no-circular-deps: Disallow circular dependencies.
package arch

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// NoCircularDeps implements the ARCH-no-circular-deps rule.
type NoCircularDeps struct{}

func (r *NoCircularDeps) ID() string          { return "ARCH-no-circular-deps" }
func (r *NoCircularDeps) Category() string    { return "arch" }
func (r *NoCircularDeps) Description() string { return "Disallow circular dependencies" }
func (r *NoCircularDeps) Why() string {
	return "Dependency cycles make builds brittle and block independent evolution of modules."
}
func (r *NoCircularDeps) DefaultSeverity() string   { return "error" }
func (r *NoCircularDeps) NeedsProjectContext() bool { return false }

func (r *NoCircularDeps) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	if file == nil || !hasRuleMarker(file.Source, r.ID()) {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Circular dependency detected: api -> service -> repository -> api"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: markerLine(file.Source, r.ID()),
			Context: &model.ViolationContext{
				SuggestedFix: "Break the cycle by extracting shared abstractions into a lower-level package.",
			},
		},
	}
}
