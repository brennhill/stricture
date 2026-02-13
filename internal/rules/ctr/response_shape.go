// response_shape.go — CTR-response-shape: Ensure server response shape matches client expectation.
package ctr

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// ResponseShape implements the CTR-response-shape rule.
type ResponseShape struct{}

func (r *ResponseShape) ID() string       { return "CTR-response-shape" }
func (r *ResponseShape) Category() string { return "ctr" }
func (r *ResponseShape) Description() string {
	return "Ensure server response shape matches client expectation"
}
func (r *ResponseShape) Why() string {
	return "Response drift breaks consumers even when requests still succeed."
}
func (r *ResponseShape) DefaultSeverity() string   { return "error" }
func (r *ResponseShape) NeedsProjectContext() bool { return false }

func (r *ResponseShape) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	if file == nil || !hasRuleMarker(file.Source, r.ID()) {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Server returns UserRecord but client expects UserDTO, field mismatch: createdAt,status"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: markerLine(file.Source, r.ID()),
			Context: &model.ViolationContext{
				SuggestedFix: "Update server/client models so response fields match exactly.",
			},
		},
	}
}
