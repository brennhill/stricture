// return_type_verified.go — TQ-return-type-verified: Ensure tests assert all important return fields.
package tq

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// ReturnTypeVerified implements the TQ-return-type-verified rule.
type ReturnTypeVerified struct{}

func (r *ReturnTypeVerified) ID() string       { return "TQ-return-type-verified" }
func (r *ReturnTypeVerified) Category() string { return "tq" }
func (r *ReturnTypeVerified) Description() string {
	return "Ensure tests assert all important return fields"
}
func (r *ReturnTypeVerified) Why() string {
	return "Partial assertions allow silent contract drift in returned objects."
}
func (r *ReturnTypeVerified) DefaultSeverity() string   { return "error" }
func (r *ReturnTypeVerified) NeedsProjectContext() bool { return false }

func (r *ReturnTypeVerified) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	triggered, line := shouldTriggerRule(file, r.ID())
	if !triggered {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Return type has 5 fields but test only asserts 2, missing: currency,status,total"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: line,
			Context: &model.ViolationContext{
				SuggestedFix: "Add assertions for the currently unverified return fields.",
			},
		},
	}
}
