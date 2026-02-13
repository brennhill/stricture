// diff.go - Drift classification between lineage artifacts.
package lineage

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Severity indicates drift impact.
type Severity string

const (
	SeverityInfo   Severity = "info"
	SeverityLow    Severity = "low"
	SeverityMedium Severity = "medium"
	SeverityHigh   Severity = "high"
)

// EnforcementMode controls whether drift above threshold blocks CI.
type EnforcementMode string

const (
	ModeBlock EnforcementMode = "block"
	ModeWarn  EnforcementMode = "warn"
)

// DriftChange is one classified difference between artifacts.
type DriftChange struct {
	Severity   Severity  `json:"severity"`
	ChangeType string    `json:"change_type"`
	FieldID    string    `json:"field_id"`
	Message    string    `json:"message"`
	Overridden bool      `json:"overridden,omitempty"`
	Override   *Override `json:"override,omitempty"`
}

// DiffSummary aggregates change counts by severity.
type DiffSummary struct {
	Total  int `json:"total"`
	High   int `json:"high"`
	Medium int `json:"medium"`
	Low    int `json:"low"`
	Info   int `json:"info"`
}

// DiffResult is the full drift report.
type DiffResult struct {
	Summary DiffSummary   `json:"summary"`
	Changes []DriftChange `json:"changes"`
}

// DiffArtifacts classifies drift from base -> head.
func DiffArtifacts(base Artifact, head Artifact) DiffResult {
	changes := make([]DriftChange, 0)

	baseByID := map[string]Annotation{}
	headByID := map[string]Annotation{}
	renameMap := map[string]string{}
	for _, f := range base.Fields {
		baseByID[f.FieldID] = f
	}
	for _, f := range head.Fields {
		headByID[f.FieldID] = f
		if f.RenamedFrom != "" {
			renameMap[f.RenamedFrom] = f.FieldID
		}
	}

	processedHead := map[string]bool{}

	for baseID, baseField := range baseByID {
		headField, exists := headByID[baseID]
		if exists {
			processedHead[baseID] = true
			changes = append(changes, compareField(baseField, headField)...)
			continue
		}

		if renamedID, renamed := renameMap[baseID]; renamed {
			headRenamed := headByID[renamedID]
			processedHead[renamedID] = true
			changes = append(changes, DriftChange{
				Severity:   SeverityMedium,
				ChangeType: "field_renamed",
				FieldID:    renamedID,
				Message:    fmt.Sprintf("Field ID renamed from %s to %s", baseID, renamedID),
			})
			changes = append(changes, compareField(baseField, headRenamed)...)
			continue
		}

		changes = append(changes, DriftChange{
			Severity:   SeverityHigh,
			ChangeType: "field_removed",
			FieldID:    baseID,
			Message:    "Field was removed from lineage artifact",
		})
	}

	for headID, headField := range headByID {
		if processedHead[headID] {
			continue
		}
		if _, existed := baseByID[headID]; existed {
			continue
		}
		if headField.RenamedFrom != "" {
			if _, existed := baseByID[headField.RenamedFrom]; existed {
				continue
			}
		}
		changes = append(changes, DriftChange{
			Severity:   SeverityMedium,
			ChangeType: "field_added",
			FieldID:    headID,
			Message:    "New field added to lineage artifact",
		})
	}

	applyOverrides(changes, head.Overrides)

	sort.Slice(changes, func(i, j int) bool {
		ri := severityRank(changes[i].Severity)
		rj := severityRank(changes[j].Severity)
		if ri != rj {
			return ri > rj
		}
		if changes[i].FieldID != changes[j].FieldID {
			return changes[i].FieldID < changes[j].FieldID
		}
		return changes[i].ChangeType < changes[j].ChangeType
	})

	summary := DiffSummary{Total: len(changes)}
	for _, c := range changes {
		switch c.Severity {
		case SeverityHigh:
			summary.High++
		case SeverityMedium:
			summary.Medium++
		case SeverityLow:
			summary.Low++
		case SeverityInfo:
			summary.Info++
		}
	}

	return DiffResult{Summary: summary, Changes: changes}
}

func compareField(base Annotation, head Annotation) []DriftChange {
	changes := make([]DriftChange, 0)
	fieldID := head.FieldID
	if fieldID == "" {
		fieldID = base.FieldID
	}

	if base.Field != head.Field {
		changes = append(changes, DriftChange{
			Severity:   SeverityMedium,
			ChangeType: "field_path_changed",
			FieldID:    fieldID,
			Message:    fmt.Sprintf("Field path changed from %s to %s", base.Field, head.Field),
		})
	}

	if base.SourceSystem != head.SourceSystem {
		changes = append(changes, DriftChange{
			Severity:   SeverityHigh,
			ChangeType: "source_system_changed",
			FieldID:    fieldID,
			Message:    fmt.Sprintf("source_system changed from %s to %s", base.SourceSystem, head.SourceSystem),
		})
	}

	if base.SourceVersion != head.SourceVersion {
		changes = append(changes, DriftChange{
			Severity:   SeverityMedium,
			ChangeType: "source_version_changed",
			FieldID:    fieldID,
			Message:    fmt.Sprintf("source_version changed from %s to %s", base.SourceVersion, head.SourceVersion),
		})
	}

	if base.MinSupportedSourceVersion != head.MinSupportedSourceVersion {
		changes = append(changes, DriftChange{
			Severity:   SeverityHigh,
			ChangeType: "min_supported_source_version_changed",
			FieldID:    fieldID,
			Message:    fmt.Sprintf("min_supported_source_version changed from %s to %s", base.MinSupportedSourceVersion, head.MinSupportedSourceVersion),
		})
	}

	if base.TransformType != head.TransformType {
		changes = append(changes, DriftChange{
			Severity:   SeverityMedium,
			ChangeType: "transform_type_changed",
			FieldID:    fieldID,
			Message:    fmt.Sprintf("transform_type changed from %s to %s", base.TransformType, head.TransformType),
		})
	}

	if base.MergeStrategy != head.MergeStrategy {
		changes = append(changes, DriftChange{
			Severity:   SeverityMedium,
			ChangeType: "merge_strategy_changed",
			FieldID:    fieldID,
			Message:    fmt.Sprintf("merge_strategy changed from %s to %s", base.MergeStrategy, head.MergeStrategy),
		})
	}

	if base.BreakPolicy != head.BreakPolicy {
		changes = append(changes, DriftChange{
			Severity:   SeverityHigh,
			ChangeType: "break_policy_changed",
			FieldID:    fieldID,
			Message:    fmt.Sprintf("break_policy changed from %s to %s", base.BreakPolicy, head.BreakPolicy),
		})
	}

	if base.Confidence != head.Confidence {
		severity := SeverityLow
		if base.Confidence == "declared" && head.Confidence == "inferred" {
			severity = SeverityMedium
		}
		changes = append(changes, DriftChange{
			Severity:   severity,
			ChangeType: "confidence_changed",
			FieldID:    fieldID,
			Message:    fmt.Sprintf("confidence changed from %s to %s", base.Confidence, head.Confidence),
		})
	}

	if base.DataClassification != head.DataClassification {
		baseRank := classificationRank(base.DataClassification)
		headRank := classificationRank(head.DataClassification)
		if headRank < baseRank {
			changes = append(changes, DriftChange{
				Severity:   SeverityHigh,
				ChangeType: "classification_relaxed",
				FieldID:    fieldID,
				Message:    fmt.Sprintf("data_classification relaxed from %s to %s", base.DataClassification, head.DataClassification),
			})
		} else {
			changes = append(changes, DriftChange{
				Severity:   SeverityLow,
				ChangeType: "classification_tightened",
				FieldID:    fieldID,
				Message:    fmt.Sprintf("data_classification tightened from %s to %s", base.DataClassification, head.DataClassification),
			})
		}
	}

	if base.Owner != head.Owner {
		changes = append(changes, DriftChange{
			Severity:   SeverityLow,
			ChangeType: "owner_changed",
			FieldID:    fieldID,
			Message:    fmt.Sprintf("owner changed from %s to %s", base.Owner, head.Owner),
		})
	}

	if base.Escalation != head.Escalation {
		changes = append(changes, DriftChange{
			Severity:   SeverityLow,
			ChangeType: "escalation_changed",
			FieldID:    fieldID,
			Message:    "escalation reference changed",
		})
	}

	if base.ContractTestID != head.ContractTestID {
		changes = append(changes, DriftChange{
			Severity:   SeverityMedium,
			ChangeType: "contract_test_id_changed",
			FieldID:    fieldID,
			Message:    "contract_test_id changed",
		})
	}

	if base.SunsetAt != head.SunsetAt {
		changes = append(changes, DriftChange{
			Severity:   SeverityMedium,
			ChangeType: "sunset_changed",
			FieldID:    fieldID,
			Message:    fmt.Sprintf("sunset_at changed from %s to %s", base.SunsetAt, head.SunsetAt),
		})
	}

	if base.Flow != head.Flow {
		changes = append(changes, DriftChange{
			Severity:   SeverityLow,
			ChangeType: "flow_changed",
			FieldID:    fieldID,
			Message:    "flow changed",
		})
	}

	if base.Note != head.Note {
		changes = append(changes, DriftChange{
			Severity:   SeverityInfo,
			ChangeType: "note_changed",
			FieldID:    fieldID,
			Message:    "note changed",
		})
	}

	changes = append(changes, compareSources(fieldID, base.Sources, head.Sources)...)
	return changes
}

func compareSources(fieldID string, baseSources []SourceRef, headSources []SourceRef) []DriftChange {
	changes := make([]DriftChange, 0)

	baseByIdentity := map[string]SourceRef{}
	headByIdentity := map[string]SourceRef{}
	for _, src := range baseSources {
		baseByIdentity[sourceIdentity(src)] = src
	}
	for _, src := range headSources {
		headByIdentity[sourceIdentity(src)] = src
	}

	for id, src := range baseByIdentity {
		headSrc, exists := headByIdentity[id]
		if !exists {
			changes = append(changes, DriftChange{
				Severity:   SeverityHigh,
				ChangeType: "source_removed",
				FieldID:    fieldID,
				Message:    fmt.Sprintf("source removed: %s", src.Raw),
			})
			continue
		}

		if src.ContractRef != headSrc.ContractRef {
			changes = append(changes, DriftChange{
				Severity:   SeverityMedium,
				ChangeType: "source_contract_ref_changed",
				FieldID:    fieldID,
				Message:    fmt.Sprintf("contract_ref changed for source %s", id),
			})
		}
		if src.ProviderID != headSrc.ProviderID {
			changes = append(changes, DriftChange{
				Severity:   SeverityMedium,
				ChangeType: "source_provider_changed",
				FieldID:    fieldID,
				Message:    fmt.Sprintf("provider_id changed for source %s", id),
			})
		}
		if src.UpstreamSystem != headSrc.UpstreamSystem {
			changes = append(changes, DriftChange{
				Severity:   SeverityMedium,
				ChangeType: "source_upstream_system_changed",
				FieldID:    fieldID,
				Message:    fmt.Sprintf("upstream_system changed for source %s", id),
			})
		}
		if src.Scope == "external" && headSrc.Scope == "external" && src.AsOf != headSrc.AsOf {
			severity, changeType := classifyAsOfChange(src.AsOf, headSrc.AsOf)
			changes = append(changes, DriftChange{
				Severity:   severity,
				ChangeType: changeType,
				FieldID:    fieldID,
				Message:    fmt.Sprintf("external as_of changed for source %s from %s to %s", id, src.AsOf, headSrc.AsOf),
			})
		}
	}

	for id, src := range headByIdentity {
		if _, exists := baseByIdentity[id]; exists {
			continue
		}
		changes = append(changes, DriftChange{
			Severity:   SeverityMedium,
			ChangeType: "source_added",
			FieldID:    fieldID,
			Message:    fmt.Sprintf("source added: %s", src.Raw),
		})
	}

	return changes
}

func classifyAsOfChange(baseDate string, headDate string) (Severity, string) {
	baseT, baseErr := time.Parse("2006-01-02", baseDate)
	headT, headErr := time.Parse("2006-01-02", headDate)
	if baseErr != nil || headErr != nil {
		return SeverityMedium, "external_as_of_changed"
	}
	if headT.Before(baseT) {
		return SeverityHigh, "external_as_of_rollback"
	}
	return SeverityLow, "external_as_of_advanced"
}

func classificationRank(classification string) int {
	switch strings.ToLower(classification) {
	case "public":
		return 1
	case "internal":
		return 2
	case "sensitive":
		return 3
	case "regulated":
		return 4
	default:
		return 0
	}
}

func severityRank(severity Severity) int {
	switch severity {
	case SeverityHigh:
		return 4
	case SeverityMedium:
		return 3
	case SeverityLow:
		return 2
	case SeverityInfo:
		return 1
	default:
		return 0
	}
}

// ParseSeverity parses a CLI threshold string.
func ParseSeverity(raw string) (Severity, error) {
	value := Severity(strings.ToLower(strings.TrimSpace(raw)))
	switch value {
	case SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo:
		return value, nil
	case "none":
		return "none", nil
	default:
		return "", fmt.Errorf("invalid severity %q (valid: high|medium|low|info|none)", raw)
	}
}

// ParseEnforcementMode parses CLI drift enforcement mode.
func ParseEnforcementMode(raw string) (EnforcementMode, error) {
	value := EnforcementMode(strings.ToLower(strings.TrimSpace(raw)))
	switch value {
	case ModeBlock, ModeWarn:
		return value, nil
	default:
		return "", fmt.Errorf("invalid mode %q (valid: block|warn)", raw)
	}
}

// ShouldFailAtThreshold reports whether any change meets/exceeds threshold.
func ShouldFailAtThreshold(result DiffResult, threshold Severity) bool {
	if threshold == "none" {
		return false
	}
	thresholdRank := severityRank(threshold)
	for _, change := range result.Changes {
		if change.Overridden {
			continue
		}
		if severityRank(change.Severity) >= thresholdRank {
			return true
		}
	}
	return false
}

// ShouldFailAtThresholdWithMode reports whether diff should return non-zero.
func ShouldFailAtThresholdWithMode(result DiffResult, threshold Severity, mode EnforcementMode) bool {
	if mode == ModeWarn {
		return false
	}
	return ShouldFailAtThreshold(result, threshold)
}

func applyOverrides(changes []DriftChange, overrides []Override) {
	if len(changes) == 0 || len(overrides) == 0 {
		return
	}

	active := activeOverrides(overrides)
	if len(active) == 0 {
		return
	}

	for i := range changes {
		match := matchOverride(changes[i], active)
		if match == nil {
			continue
		}
		changes[i].Overridden = true
		overrideCopy := *match
		changes[i].Override = &overrideCopy
	}
}

func activeOverrides(overrides []Override) []Override {
	nowDate := time.Now().UTC().Format("2006-01-02")
	now, _ := time.Parse("2006-01-02", nowDate)
	result := make([]Override, 0, len(overrides))
	for _, ov := range overrides {
		expires, err := time.Parse("2006-01-02", ov.Expires)
		if err != nil {
			continue
		}
		if expires.Before(now) {
			continue
		}
		result = append(result, ov)
	}
	return result
}

func matchOverride(change DriftChange, overrides []Override) *Override {
	for i := range overrides {
		ov := &overrides[i]
		if ov.FieldID != change.FieldID {
			continue
		}
		if ov.ChangeType == "*" || ov.ChangeType == change.ChangeType {
			return ov
		}
	}
	return nil
}
