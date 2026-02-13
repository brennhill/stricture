// assertion_depth.go — TQ-assertion-depth: Require nested assertions for composite structures.
package tq

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// AssertionDepth implements the TQ-assertion-depth rule.
type AssertionDepth struct{}

func (r *AssertionDepth) ID() string       { return "TQ-assertion-depth" }
func (r *AssertionDepth) Category() string { return "tq" }
func (r *AssertionDepth) Description() string {
	return "Require nested assertions for composite structures"
}
func (r *AssertionDepth) Why() string {
	return "Checking only parent containers misses nested contract regressions."
}
func (r *AssertionDepth) DefaultSeverity() string   { return "error" }
func (r *AssertionDepth) NeedsProjectContext() bool { return false }

func (r *AssertionDepth) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	if file == nil || !hasRuleMarker(file.Source, r.ID()) {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Nested object user.profile.address is not asserted, only parent user.profile is checked"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: markerLine(file.Source, r.ID()),
			Context: &model.ViolationContext{
				SuggestedFix: "Add assertions for nested fields that affect behavior.",
			},
		},
	}
}
