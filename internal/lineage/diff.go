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
	Severity   Severity   `json:"severity"`
	ChangeType string     `json:"change_type"`
	FieldID    string     `json:"field_id"`
	Message    string     `json:"message"`
	Source     *DriftEdge `json:"source,omitempty"`
	Impact     *DriftEdge `json:"impact,omitempty"`
	TypeDelta  *TypeDelta `json:"type_delta,omitempty"`
	ModifiedBy []string   `json:"modified_by,omitempty"`
	Validation string     `json:"validation,omitempty"`
	Suggestion string     `json:"suggestion,omitempty"`
	Overridden bool       `json:"overridden,omitempty"`
	Override   *Override  `json:"override,omitempty"`
}

// DriftEdge describes the producer/consumer side of a drift change.
type DriftEdge struct {
	Service string `json:"service,omitempty"`
	API     string `json:"api,omitempty"`
}

// TypeDelta captures typed-shape metadata when available.
type TypeDelta struct {
	Change            string `json:"change,omitempty"` // expanded|contracted|changed|unknown
	BeforeContractRef string `json:"before_contract_ref,omitempty"`
	AfterContractRef  string `json:"after_contract_ref,omitempty"`
	BeforeLabel       string `json:"before_label,omitempty"`
	AfterLabel        string `json:"after_label,omitempty"`
}

func edgeOrNil(service string, api string) *DriftEdge {
	service = strings.TrimSpace(service)
	api = strings.TrimSpace(api)
	if service == "" && api == "" {
		return nil
	}
	return &DriftEdge{Service: service, API: api}
}

func edgeCopy(edge *DriftEdge) *DriftEdge {
	if edge == nil {
		return nil
	}
	copy := *edge
	return &copy
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

type plainLanguageGuidance struct {
	Impact   string
	NextStep string
}

var plainLanguageGuidanceByChangeType = map[string]plainLanguageGuidance{
	"field_removed": {
		Impact:   "Consumers expecting this field can crash or silently mis-handle payloads.",
		NextStep: "Restore the field or ship a compatibility adapter and coordinated rollout.",
	},
	"source_version_changed": {
		Impact:   "Version bumps can change producer semantics even when type shape is unchanged.",
		NextStep: "Run compatibility tests and communicate rollout plan to consumers.",
	},
	"merge_strategy_changed": {
		Impact:   "Different merge behavior can change output values without any schema/type diff.",
		NextStep: "Add merge semantic regression tests with representative multi-source fixtures.",
	},
	"source_contract_ref_changed": {
		Impact:   "Contract reference changes can introduce enum/type drift across service boundaries.",
		NextStep: "Run contract tests for producer and consumers before promoting the new reference.",
	},
	"source_removed": {
		Impact:   "Removing a source can change or null out producer output values.",
		NextStep: "Confirm fallback behavior and downstream assumptions with integration tests.",
	},
	"source_added": {
		Impact:   "Adding a source can alter precedence and value composition.",
		NextStep: "Validate precedence rules and update consumer expectations.",
	},
	"external_as_of_rollback": {
		Impact:   "Older external snapshots can reintroduce stale or incompatible values.",
		NextStep: "Refresh provider snapshot and verify time-sensitive invariants.",
	},
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

	for i := range changes {
		enrichPlainLanguage(&changes[i])
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

	producerService := strings.TrimSpace(head.SourceSystem)
	if producerService == "" {
		producerService = strings.TrimSpace(base.SourceSystem)
	}
	producerField := strings.TrimSpace(head.Field)
	if producerField == "" {
		producerField = strings.TrimSpace(base.Field)
	}
	if producerField == "" {
		producerField = fieldID
	}
	producerEdge := edgeOrNil(producerService, producerField)
	impactEdge := edgeOrNil(producerService, fieldID)

	appendProducerChange := func(severity Severity, changeType string, message string) {
		changes = append(changes, DriftChange{
			Severity:   severity,
			ChangeType: changeType,
			FieldID:    fieldID,
			Message:    message,
			Source:     edgeCopy(producerEdge),
			Impact:     edgeCopy(impactEdge),
		})
	}

	if base.Field != head.Field {
		appendProducerChange(SeverityMedium, "field_path_changed", fmt.Sprintf("Producer %s changed field path from %s to %s for %s.", producerServiceLabel(producerService), base.Field, head.Field, fieldID))
	}

	if base.SourceSystem != head.SourceSystem {
		appendProducerChange(SeverityHigh, "source_system_changed", fmt.Sprintf("Producer system changed from %s to %s for %s.", base.SourceSystem, head.SourceSystem, fieldID))
	}

	if base.SourceVersion != head.SourceVersion {
		appendProducerChange(SeverityMedium, "source_version_changed", fmt.Sprintf("Producer %s changed source version for %s from %s to %s. Consumers pinned to previous semantics should verify compatibility.", producerServiceLabel(producerService), fieldID, base.SourceVersion, head.SourceVersion))
	}

	if base.MinSupportedSourceVersion != head.MinSupportedSourceVersion {
		appendProducerChange(SeverityHigh, "min_supported_source_version_changed", fmt.Sprintf("Producer %s changed min supported source version for %s from %s to %s.", producerServiceLabel(producerService), fieldID, base.MinSupportedSourceVersion, head.MinSupportedSourceVersion))
	}

	if base.TransformType != head.TransformType {
		appendProducerChange(SeverityMedium, "transform_type_changed", fmt.Sprintf("Producer %s changed transform type for %s from %s to %s.", producerServiceLabel(producerService), fieldID, base.TransformType, head.TransformType))
	}

	if base.MergeStrategy != head.MergeStrategy {
		changes = append(changes, DriftChange{
			Severity:   SeverityMedium,
			ChangeType: "merge_strategy_changed",
			FieldID:    fieldID,
			Message:    fmt.Sprintf("Producer %s changed merge strategy for %s from %s (%s) to %s (%s). This changes how upstream values are combined and can change downstream behavior even when schema/type stays the same.", producerServiceLabel(producerService), fieldID, base.MergeStrategy, mergeStrategyMeaning(base.MergeStrategy), head.MergeStrategy, mergeStrategyMeaning(head.MergeStrategy)),
			Source:     edgeCopy(producerEdge),
			Impact:     edgeCopy(impactEdge),
			Validation: "merge semantics regression check for precedence/fallback behavior",
			Suggestion: "add fixture tests for representative source combinations and verify downstream expectations before rollout",
		})
	}

	if base.BreakPolicy != head.BreakPolicy {
		appendProducerChange(SeverityHigh, "break_policy_changed", fmt.Sprintf("Producer %s changed break policy for %s from %s to %s.", producerServiceLabel(producerService), fieldID, base.BreakPolicy, head.BreakPolicy))
	}

	if base.Confidence != head.Confidence {
		severity := SeverityLow
		if base.Confidence == "declared" && head.Confidence == "inferred" {
			severity = SeverityMedium
		}
		appendProducerChange(severity, "confidence_changed", fmt.Sprintf("Producer %s changed confidence for %s from %s to %s.", producerServiceLabel(producerService), fieldID, base.Confidence, head.Confidence))
	}

	if base.DataClassification != head.DataClassification {
		baseRank := classificationRank(base.DataClassification)
		headRank := classificationRank(head.DataClassification)
		if headRank < baseRank {
			appendProducerChange(SeverityHigh, "classification_relaxed", fmt.Sprintf("Producer %s relaxed data classification for %s from %s to %s.", producerServiceLabel(producerService), fieldID, base.DataClassification, head.DataClassification))
		} else {
			appendProducerChange(SeverityLow, "classification_tightened", fmt.Sprintf("Producer %s tightened data classification for %s from %s to %s.", producerServiceLabel(producerService), fieldID, base.DataClassification, head.DataClassification))
		}
	}

	if base.Owner != head.Owner {
		appendProducerChange(SeverityLow, "owner_changed", fmt.Sprintf("Owner changed from %s to %s for %s.", base.Owner, head.Owner, fieldID))
	}

	if base.Escalation != head.Escalation {
		appendProducerChange(SeverityLow, "escalation_changed", fmt.Sprintf("Escalation contact changed for %s.", fieldID))
	}

	if base.ContractTestID != head.ContractTestID {
		appendProducerChange(SeverityMedium, "contract_test_id_changed", fmt.Sprintf("Contract test reference changed for %s.", fieldID))
	}

	if base.SunsetAt != head.SunsetAt {
		appendProducerChange(SeverityMedium, "sunset_changed", fmt.Sprintf("Sunset date changed for %s from %s to %s.", fieldID, base.SunsetAt, head.SunsetAt))
	}

	if base.Flow != head.Flow {
		appendProducerChange(SeverityLow, "flow_changed", fmt.Sprintf("Flow narrative changed for %s.", fieldID))
	}

	if base.Note != head.Note {
		appendProducerChange(SeverityInfo, "note_changed", fmt.Sprintf("Annotation note changed for %s.", fieldID))
	}

	changes = append(changes, compareSources(fieldID, head.SourceSystem, head.Flow, base.Sources, head.Sources)...)
	return changes
}

func compareSources(fieldID string, impactedService string, flow string, baseSources []SourceRef, headSources []SourceRef) []DriftChange {
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
		sourceService := sourceServiceForRef(src)
		sourceEdge := edgeOrNil(sourceService, src.Target)
		impactEdge := edgeOrNil(impactedService, fieldID)
		if !exists {
			changes = append(changes, DriftChange{
				Severity:   SeverityHigh,
				ChangeType: "source_removed",
				FieldID:    fieldID,
				Message:    fmt.Sprintf("Source %s no longer contributes to %s in producer %s. This can change downstream values.", sourceServiceLabel(sourceService), fieldID, producerServiceLabel(impactedService)),
				Source:     edgeCopy(sourceEdge),
				Impact:     edgeCopy(impactEdge),
			})
			continue
		}

		if src.ContractRef != headSrc.ContractRef {
			modifiers := flowModifiers(flow, sourceService, impactedService)
			fromRef := contractRefLabel(src.ContractRef)
			toRef := contractRefLabel(headSrc.ContractRef)
			validation := "enum/type compatibility check against the consumer contract"
			suggestion := "add unit tests + contract assertions for allowed enums/types, then roll out producer and consumers together"
			changes = append(changes, DriftChange{
				Severity:   SeverityMedium,
				ChangeType: "source_contract_ref_changed",
				FieldID:    fieldID,
				Message:    contractRefNarrative(sourceService, impactedService, fieldID, src.Target, modifiers, validation, suggestion, fromRef, toRef),
				Source: &DriftEdge{
					Service: sourceService,
					API:     src.Target,
				},
				Impact: &DriftEdge{
					Service: impactedService,
					API:     fieldID,
				},
				TypeDelta: &TypeDelta{
					Change:            "changed",
					BeforeContractRef: src.ContractRef,
					AfterContractRef:  headSrc.ContractRef,
					BeforeLabel:       fromRef,
					AfterLabel:        toRef,
				},
				ModifiedBy: modifiers,
				Validation: validation,
				Suggestion: suggestion,
			})
		}
		if src.ProviderID != headSrc.ProviderID {
			changes = append(changes, DriftChange{
				Severity:   SeverityMedium,
				ChangeType: "source_provider_changed",
				FieldID:    fieldID,
				Message:    fmt.Sprintf("Provider changed for source %s from %s to %s in %s.", id, src.ProviderID, headSrc.ProviderID, fieldID),
				Source:     edgeCopy(sourceEdge),
				Impact:     edgeCopy(impactEdge),
			})
		}
		if src.UpstreamSystem != headSrc.UpstreamSystem {
			changes = append(changes, DriftChange{
				Severity:   SeverityMedium,
				ChangeType: "source_upstream_system_changed",
				FieldID:    fieldID,
				Message:    fmt.Sprintf("Upstream system changed for source %s from %s to %s in %s.", id, src.UpstreamSystem, headSrc.UpstreamSystem, fieldID),
				Source:     edgeCopy(sourceEdge),
				Impact:     edgeCopy(impactEdge),
			})
		}
		if src.Scope == "external" && headSrc.Scope == "external" && src.AsOf != headSrc.AsOf {
			severity, changeType := classifyAsOfChange(src.AsOf, headSrc.AsOf)
			changes = append(changes, DriftChange{
				Severity:   severity,
				ChangeType: changeType,
				FieldID:    fieldID,
				Message:    fmt.Sprintf("External snapshot date changed for source %s from %s to %s in %s.", id, src.AsOf, headSrc.AsOf, fieldID),
				Source:     edgeCopy(sourceEdge),
				Impact:     edgeCopy(impactEdge),
			})
		}
	}

	for id, src := range headByIdentity {
		if _, exists := baseByIdentity[id]; exists {
			continue
		}
		sourceService := sourceServiceForRef(src)
		changes = append(changes, DriftChange{
			Severity:   SeverityMedium,
			ChangeType: "source_added",
			FieldID:    fieldID,
			Message:    fmt.Sprintf("Source %s was added for %s in producer %s.", sourceServiceLabel(sourceService), fieldID, producerServiceLabel(impactedService)),
			Source:     edgeOrNil(sourceService, src.Target),
			Impact:     edgeOrNil(impactedService, fieldID),
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

func contractRefLabel(ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "unknown"
	}
	if rev := contractRefRevision(ref); rev != "" {
		return rev
	}
	if len(ref) > 72 {
		return ref[:69] + "..."
	}
	return ref
}

func contractRefRevision(ref string) string {
	idx := strings.LastIndex(strings.TrimSpace(ref), "@")
	if idx < 0 || idx == len(ref)-1 {
		return ""
	}
	rev := strings.TrimSpace(ref[idx+1:])
	if rev == "" {
		return ""
	}
	for _, ch := range rev {
		if ch == '?' || ch == '&' || ch == '/' {
			return ""
		}
	}
	return rev
}

func sourceServiceForRef(src SourceRef) string {
	if up := strings.TrimSpace(src.UpstreamSystem); up != "" {
		return up
	}
	target := strings.TrimSpace(src.Target)
	if target == "" {
		return "unknown_source"
	}
	if dot := strings.Index(target, "."); dot > 0 {
		return target[:dot]
	}
	return target
}

func flowModifiers(flow string, sourceService string, impactedService string) []string {
	parts := strings.Fields(flow)
	if len(parts) == 0 {
		return nil
	}
	nodes := make([]string, 0, len(parts))
	for _, part := range parts {
		if !strings.HasPrefix(part, "@") {
			continue
		}
		node := strings.TrimPrefix(part, "@")
		if strings.EqualFold(node, "self") {
			node = impactedService
		}
		node = strings.TrimSpace(node)
		if node == "" {
			continue
		}
		nodes = append(nodes, node)
	}
	if len(nodes) <= 2 {
		return nil
	}
	mods := make([]string, 0, len(nodes))
	seen := map[string]bool{}
	for _, node := range nodes[1 : len(nodes)-1] {
		if node == sourceService || node == impactedService || seen[node] {
			continue
		}
		seen[node] = true
		mods = append(mods, node)
	}
	return mods
}

func contractRefNarrative(
	sourceService string,
	impactedService string,
	fieldID string,
	sourceAPI string,
	modifiers []string,
	validation string,
	suggestion string,
	fromLabel string,
	toLabel string,
) string {
	modText := "none"
	if len(modifiers) > 0 {
		modText = strings.Join(modifiers, ", ")
	}
	return fmt.Sprintf(
		"Producer %s changed contract reference for %s in %s from %s to %s. Stricture flags potential impact in %s. Field modifiers en-route: %s. Validation that can fail: %s. Suggestion: %s.",
		sourceService,
		fieldID,
		sourceAPI,
		fromLabel,
		toLabel,
		impactedService,
		modText,
		validation,
		suggestion,
	)
}

func enrichPlainLanguage(change *DriftChange) {
	template, ok := plainLanguageGuidanceByChangeType[change.ChangeType]
	if !ok {
		return
	}
	if change.Validation == "" {
		change.Validation = template.Impact
	}
	if change.Suggestion == "" {
		change.Suggestion = template.NextStep
	}
}

func producerServiceLabel(service string) string {
	service = strings.TrimSpace(service)
	if service == "" {
		return "unknown producer"
	}
	return service
}

func sourceServiceLabel(service string) string {
	service = strings.TrimSpace(service)
	if service == "" {
		return "unknown source"
	}
	return service
}

func mergeStrategyMeaning(strategy string) string {
	switch strings.TrimSpace(strategy) {
	case "single_source":
		return "one source is selected; other sources are ignored"
	case "priority":
		return "sources are evaluated in priority order"
	case "first_non_null":
		return "first non-null source value wins"
	case "union":
		return "values from multiple sources are combined"
	case "custom":
		return "custom merge logic applies service-specific rules"
	default:
		return "strategy semantics are custom/unknown"
	}
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
	result := make([]Override, 0, len(overrides))
	for _, ov := range overrides {
		expires, err := time.Parse("2006-01-02", ov.Expires)
		if err != nil {
			continue
		}
		if expires.Format("2006-01-02") < nowDate {
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
