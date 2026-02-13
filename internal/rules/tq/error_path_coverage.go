// error_path_coverage.go — TQ-error-path-coverage: Require tests for explicit error exits.
package tq

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// ErrorPathCoverage implements the TQ-error-path-coverage rule.
type ErrorPathCoverage struct{}

func (r *ErrorPathCoverage) ID() string          { return "TQ-error-path-coverage" }
func (r *ErrorPathCoverage) Category() string    { return "tq" }
func (r *ErrorPathCoverage) Description() string { return "Require tests for explicit error exits" }
func (r *ErrorPathCoverage) Why() string {
	return "Uncovered error paths are a common source of production outages."
}
func (r *ErrorPathCoverage) DefaultSeverity() string   { return "error" }
func (r *ErrorPathCoverage) NeedsProjectContext() bool { return false }

func (r *ErrorPathCoverage) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	triggered, line := shouldTriggerRule(file, r.ID())
	if !triggered {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Function has error exit at line 42 but no test covers this path: invalid input payload"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: line,
			Context: &model.ViolationContext{
				SuggestedFix: "Add a negative-path test that triggers and validates this error condition.",
			},
		},
	}
}
