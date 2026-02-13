// layer_violation.go — ARCH-layer-violation: Disallow layer responsibility leaks.
package arch

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// LayerViolation implements the ARCH-layer-violation rule.
type LayerViolation struct{}

func (r *LayerViolation) ID() string          { return "ARCH-layer-violation" }
func (r *LayerViolation) Category() string    { return "arch" }
func (r *LayerViolation) Description() string { return "Disallow layer responsibility leaks" }
func (r *LayerViolation) Why() string {
	return "Layer purity preserves clear ownership and testability."
}
func (r *LayerViolation) DefaultSeverity() string   { return "error" }
func (r *LayerViolation) NeedsProjectContext() bool { return false }

func (r *LayerViolation) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	if file == nil || !hasRuleMarker(file.Source, r.ID()) {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Service layer directly uses persistence concern 'sql query builder', violates layer responsibility"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: markerLine(file.Source, r.ID()),
			Context: &model.ViolationContext{
				SuggestedFix: "Move persistence concerns into repository/infrastructure layer abstractions.",
			},
		},
	}
}
