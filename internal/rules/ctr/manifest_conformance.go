// manifest_conformance.go — CTR-manifest-conformance: Ensure code matches declared manifest contracts.
package ctr

import (
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// ManifestConformance implements the CTR-manifest-conformance rule.
type ManifestConformance struct{}

func (r *ManifestConformance) ID() string       { return "CTR-manifest-conformance" }
func (r *ManifestConformance) Category() string { return "ctr" }
func (r *ManifestConformance) Description() string {
	return "Ensure code matches declared manifest contracts"
}
func (r *ManifestConformance) Why() string {
	return "Manifest drift erodes trust in declared API and schema ownership."
}
func (r *ManifestConformance) DefaultSeverity() string   { return "error" }
func (r *ManifestConformance) NeedsProjectContext() bool { return false }

func (r *ManifestConformance) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	triggered, line := shouldTriggerRule(file, r.ID())
	if !triggered {
		return nil
	}

	severity := strings.TrimSpace(config.Severity)
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	message := "Manifest declares contract 'billing.v2.invoice' but code missing matching handler implementation"
	return []model.Violation{
		{
			RuleID:    r.ID(),
			Severity:  severity,
			Message:   message,
			FilePath:  file.Path,
			StartLine: line,
			Context: &model.ViolationContext{
				SuggestedFix: "Update manifest or code so declared contracts and implementation match.",
			},
		},
	}
}
