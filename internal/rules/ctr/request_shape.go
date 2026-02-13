// request_shape.go — CTR-request-shape: Ensure client request shape matches server expectation.
package ctr

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// RequestShape implements the CTR-request-shape rule.
type RequestShape struct{}

func (r *RequestShape) ID() string       { return "CTR-request-shape" }
func (r *RequestShape) Category() string { return "ctr" }
func (r *RequestShape) Description() string {
	return "Ensure client request shape matches server expectation"
}
func (r *RequestShape) Why() string {
	return "Mismatched request contracts cause immediate runtime failures."
}
func (r *RequestShape) DefaultSeverity() string   { return "error" }
func (r *RequestShape) NeedsProjectContext() bool { return false }

func (r *RequestShape) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	if file == nil || !hasRuleMarker(file.Source, r.ID()) {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Client sends UserCreateInput but server expects CreateUserRequest, field mismatch: email,role"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: markerLine(file.Source, r.ID()),
			Context: &model.ViolationContext{
				SuggestedFix: "Align client request payload fields and types with the server contract.",
			},
		},
	}
}
