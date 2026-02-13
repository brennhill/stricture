// negative_cases.go — TQ-negative-cases: Require at least one negative-path test when positives exist.
package tq

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// NegativeCases implements the TQ-negative-cases rule.
type NegativeCases struct{}

func (r *NegativeCases) ID() string       { return "TQ-negative-cases" }
func (r *NegativeCases) Category() string { return "tq" }
func (r *NegativeCases) Description() string {
	return "Require at least one negative-path test when positives exist"
}
func (r *NegativeCases) Why() string {
	return "Negative tests enforce defensive behavior and error handling contracts."
}
func (r *NegativeCases) DefaultSeverity() string   { return "error" }
func (r *NegativeCases) NeedsProjectContext() bool { return false }

func (r *NegativeCases) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	if file == nil || !hasRuleMarker(file.Source, r.ID()) {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Function CreateInvoice has 3 positive tests but 0 negative tests (expected at least 1)"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: markerLine(file.Source, r.ID()),
			Context: &model.ViolationContext{
				SuggestedFix: "Add at least one failing input test that validates error behavior.",
			},
		},
	}
}
