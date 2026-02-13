// test_naming.go — TQ-test-naming: Enforce behavior-oriented test naming.
package tq

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// TestNaming implements the TQ-test-naming rule.
type TestNaming struct{}

func (r *TestNaming) ID() string          { return "TQ-test-naming" }
func (r *TestNaming) Category() string    { return "tq" }
func (r *TestNaming) Description() string { return "Enforce behavior-oriented test naming" }
func (r *TestNaming) Why() string {
	return "Clear behavior-focused names improve maintainability and diagnosis."
}
func (r *TestNaming) DefaultSeverity() string   { return "error" }
func (r *TestNaming) NeedsProjectContext() bool { return false }

func (r *TestNaming) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	triggered, line := shouldTriggerRule(file, r.ID())
	if !triggered {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Test name 'TestHandlerImpl' does not match pattern 'TestSubject_WhenCondition_ThenOutcome', should describe behavior not implementation"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: line,
			Context: &model.ViolationContext{
				SuggestedFix: "Rename tests to describe observable behavior and expected outcome.",
			},
		},
	}
}
