// mock_scope.go — TQ-mock-scope: Ensure mocks are scoped and cleaned up per test.
package tq

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// MockScope implements the TQ-mock-scope rule.
type MockScope struct{}

func (r *MockScope) ID() string          { return "TQ-mock-scope" }
func (r *MockScope) Category() string    { return "tq" }
func (r *MockScope) Description() string { return "Ensure mocks are scoped and cleaned up per test" }
func (r *MockScope) Why() string {
	return "Leaky mocks create flaky suites and hidden coupling across tests."
}
func (r *MockScope) DefaultSeverity() string   { return "error" }
func (r *MockScope) NeedsProjectContext() bool { return false }

func (r *MockScope) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	triggered, line := shouldTriggerRule(file, r.ID())
	if !triggered {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Mock paymentGateway created at package scope is not cleaned up, causing test pollution"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: line,
			Context: &model.ViolationContext{
				SuggestedFix: "Limit mock scope to each test and reset in cleanup hooks.",
			},
		},
	}
}
