// no_shallow_assertions.go — TQ-no-shallow-assertions: Reject assertions that only check existence.
package tq

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// NoShallowAssertions implements the TQ-no-shallow-assertions rule.
type NoShallowAssertions struct{}

func (r *NoShallowAssertions) ID() string       { return "TQ-no-shallow-assertions" }
func (r *NoShallowAssertions) Category() string { return "tq" }
func (r *NoShallowAssertions) Description() string {
	return "Reject assertions that only check existence"
}
func (r *NoShallowAssertions) Why() string {
	return "Shallow assertions hide regressions because they never verify actual values."
}
func (r *NoShallowAssertions) DefaultSeverity() string   { return "error" }
func (r *NoShallowAssertions) NeedsProjectContext() bool { return false }

func (r *NoShallowAssertions) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	if file == nil || !hasRuleMarker(file.Source, r.ID()) {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Shallow assertion assert.NotNil only checks existence, replace with assert.Equal to verify value"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: markerLine(file.Source, r.ID()),
			Context: &model.ViolationContext{
				SuggestedFix: "Assert concrete expected values, not only that an object exists.",
			},
		},
	}
}
