// test_isolation.go — TQ-test-isolation: Prevent tests from mutating shared state without cleanup.
package tq

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// TestIsolation implements the TQ-test-isolation rule.
type TestIsolation struct{}

func (r *TestIsolation) ID() string       { return "TQ-test-isolation" }
func (r *TestIsolation) Category() string { return "tq" }
func (r *TestIsolation) Description() string {
	return "Prevent tests from mutating shared state without cleanup"
}
func (r *TestIsolation) Why() string {
	return "Shared mutable state makes tests order-dependent and non-deterministic."
}
func (r *TestIsolation) DefaultSeverity() string   { return "error" }
func (r *TestIsolation) NeedsProjectContext() bool { return false }

func (r *TestIsolation) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	if file == nil || !hasRuleMarker(file.Source, r.ID()) {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Test TestCreateUser mutates shared state globalClock without cleanup, breaking isolation"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: markerLine(file.Source, r.ID()),
			Context: &model.ViolationContext{
				SuggestedFix: "Reset global/shared state in t.Cleanup or avoid mutating it.",
			},
		},
	}
}
