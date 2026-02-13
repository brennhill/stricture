// strictness_parity.go — CTR-strictness-parity: Require required/optional parity across contract boundaries.
package ctr

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// StrictnessParity implements the CTR-strictness-parity rule.
type StrictnessParity struct{}

func (r *StrictnessParity) ID() string       { return "CTR-strictness-parity" }
func (r *StrictnessParity) Category() string { return "ctr" }
func (r *StrictnessParity) Description() string {
	return "Require required/optional parity across contract boundaries"
}
func (r *StrictnessParity) Why() string {
	return "Different strictness interpretations cause latent production breaks."
}
func (r *StrictnessParity) DefaultSeverity() string   { return "error" }
func (r *StrictnessParity) NeedsProjectContext() bool { return false }

func (r *StrictnessParity) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	if file == nil || !hasRuleMarker(file.Source, r.ID()) {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Field 'userId' is required on client but optional on server"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: markerLine(file.Source, r.ID()),
			Context: &model.ViolationContext{
				SuggestedFix: "Make strictness consistent for this field in all contract definitions.",
			},
		},
	}
}
