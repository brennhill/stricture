// boundary_tested.go — TQ-boundary-tested: Require boundary value coverage in tests.
package tq

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// BoundaryTested implements the TQ-boundary-tested rule.
type BoundaryTested struct{}

func (r *BoundaryTested) ID() string          { return "TQ-boundary-tested" }
func (r *BoundaryTested) Category() string    { return "tq" }
func (r *BoundaryTested) Description() string { return "Require boundary value coverage in tests" }
func (r *BoundaryTested) Why() string {
	return "Boundary cases catch edge bugs that happy-path tests miss."
}
func (r *BoundaryTested) DefaultSeverity() string   { return "error" }
func (r *BoundaryTested) NeedsProjectContext() bool { return false }

func (r *BoundaryTested) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	triggered, line := shouldTriggerRule(file, r.ID())
	if !triggered {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Function accepts integer but tests only cover 1,10, missing boundary: 0 and max int"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: line,
			Context: &model.ViolationContext{
				SuggestedFix: "Include min, max, empty, and invalid boundary values in tests.",
			},
		},
	}
}
