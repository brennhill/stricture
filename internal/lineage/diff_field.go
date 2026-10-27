// diff_field.go — Field-level drift comparison between base and head lineage annotations.
package lineage

import (
	"fmt"
	"strings"
)

// fieldCompare carries the shared context for comparing one field across artifacts.
type fieldCompare struct {
	base, head      Annotation
	fieldID         string
	producerService string
	producerEdge    *DriftEdge
	impactEdge      *DriftEdge
}

// change builds a producer-scoped drift change with fresh edge copies.
func (c fieldCompare) change(severity Severity, changeType string, message string) DriftChange {
	return DriftChange{
		Severity:   severity,
		ChangeType: changeType,
		FieldID:    c.fieldID,
		Message:    message,
		Source:     edgeCopy(c.producerEdge),
		Impact:     edgeCopy(c.impactEdge),
	}
}

func (c fieldCompare) producerLabel() string {
	return producerServiceLabel(c.producerService)
}

type attrCheck func(c fieldCompare) []DriftChange

// attrDiff builds a check that fires when get(base) != get(head).
func attrDiff(get func(Annotation) string, severity func(b, h string) Severity, changeType string, message func(c fieldCompare, b, h string) string) attrCheck {
	return func(c fieldCompare) []DriftChange {
		b, h := get(c.base), get(c.head)
		if b == h {
			return nil
		}
		return []DriftChange{c.change(severity(b, h), changeType, message(c, b, h))}
	}
}

func fixedSeverity(s Severity) func(b, h string) Severity {
	return func(string, string) Severity { return s }
}

func fieldPathOf(a Annotation) string     { return a.Field }
func sourceSystemOf(a Annotation) string  { return a.SourceSystem }
func sourceVersionOf(a Annotation) string { return a.SourceVersion }
func minSupportedOf(a Annotation) string  { return a.MinSupportedSourceVersion }
func transformTypeOf(a Annotation) string { return a.TransformType }
func breakPolicyOf(a Annotation) string   { return a.BreakPolicy }
func confidenceOf(a Annotation) string    { return a.Confidence }
func ownerOf(a Annotation) string         { return a.Owner }
func escalationOf(a Annotation) string    { return a.Escalation }
func contractTestOf(a Annotation) string  { return a.ContractTestID }
func sunsetOf(a Annotation) string        { return a.SunsetAt }
func flowOf(a Annotation) string          { return a.Flow }
func noteOf(a Annotation) string          { return a.Note }

func confidenceSeverity(b, h string) Severity {
	if b == "declared" && h == "inferred" {
		return SeverityMedium
	}
	return SeverityLow
}

func checkMergeStrategy(c fieldCompare) []DriftChange {
	if c.base.MergeStrategy == c.head.MergeStrategy {
		return nil
	}
	change := c.change(SeverityMedium, "merge_strategy_changed", fmt.Sprintf("Producer %s changed merge strategy for %s from %s (%s) to %s (%s). This changes how upstream values are combined and can change downstream behavior even when schema/type stays the same.", c.producerLabel(), c.fieldID, c.base.MergeStrategy, mergeStrategyMeaning(c.base.MergeStrategy), c.head.MergeStrategy, mergeStrategyMeaning(c.head.MergeStrategy)))
	change.Validation = "merge semantics regression check for precedence/fallback behavior"
	change.Suggestion = "add fixture tests for representative source combinations and verify downstream expectations before rollout"
	return []DriftChange{change}
}

func checkClassification(c fieldCompare) []DriftChange {
	if c.base.DataClassification == c.head.DataClassification {
		return nil
	}
	if classificationRank(c.head.DataClassification) < classificationRank(c.base.DataClassification) {
		return []DriftChange{c.change(SeverityHigh, "classification_relaxed", fmt.Sprintf("Producer %s relaxed data classification for %s from %s to %s.", c.producerLabel(), c.fieldID, c.base.DataClassification, c.head.DataClassification))}
	}
	return []DriftChange{c.change(SeverityLow, "classification_tightened", fmt.Sprintf("Producer %s tightened data classification for %s from %s to %s.", c.producerLabel(), c.fieldID, c.base.DataClassification, c.head.DataClassification))}
}

// producerFromTo formats the common "Producer X changed <what> for F from A to B." message.
func producerFromTo(what string) func(c fieldCompare, b, h string) string {
	return func(c fieldCompare, b, h string) string {
		return fmt.Sprintf("Producer %s changed %s for %s from %s to %s.", c.producerLabel(), what, c.fieldID, b, h)
	}
}

func staticMessage(format string) func(c fieldCompare, b, h string) string {
	return func(c fieldCompare, _, _ string) string { return fmt.Sprintf(format, c.fieldID) }
}

// fieldChecks run in order; the first entry that fires is reported first.
var fieldChecks = []attrCheck{
	attrDiff(fieldPathOf, fixedSeverity(SeverityMedium), "field_path_changed", func(c fieldCompare, b, h string) string {
		return fmt.Sprintf("Producer %s changed field path from %s to %s for %s.", c.producerLabel(), b, h, c.fieldID)
	}),
	attrDiff(sourceSystemOf, fixedSeverity(SeverityHigh), "source_system_changed", func(c fieldCompare, b, h string) string {
		return fmt.Sprintf("Producer system changed from %s to %s for %s.", b, h, c.fieldID)
	}),
	attrDiff(sourceVersionOf, fixedSeverity(SeverityMedium), "source_version_changed", func(c fieldCompare, b, h string) string {
		return fmt.Sprintf("Producer %s changed source version for %s from %s to %s. Consumers pinned to previous semantics should verify compatibility.", c.producerLabel(), c.fieldID, b, h)
	}),
	attrDiff(minSupportedOf, fixedSeverity(SeverityHigh), "min_supported_source_version_changed", producerFromTo("min supported source version")),
	attrDiff(transformTypeOf, fixedSeverity(SeverityMedium), "transform_type_changed", producerFromTo("transform type")),
	checkMergeStrategy,
	attrDiff(breakPolicyOf, fixedSeverity(SeverityHigh), "break_policy_changed", producerFromTo("break policy")),
	attrDiff(confidenceOf, confidenceSeverity, "confidence_changed", producerFromTo("confidence")),
	checkClassification,
	attrDiff(ownerOf, fixedSeverity(SeverityLow), "owner_changed", func(c fieldCompare, b, h string) string {
		return fmt.Sprintf("Owner changed from %s to %s for %s.", b, h, c.fieldID)
	}),
	attrDiff(escalationOf, fixedSeverity(SeverityLow), "escalation_changed", staticMessage("Escalation contact changed for %s.")),
	attrDiff(contractTestOf, fixedSeverity(SeverityMedium), "contract_test_id_changed", staticMessage("Contract test reference changed for %s.")),
	attrDiff(sunsetOf, fixedSeverity(SeverityMedium), "sunset_changed", func(c fieldCompare, b, h string) string {
		return fmt.Sprintf("Sunset date changed for %s from %s to %s.", c.fieldID, b, h)
	}),
	attrDiff(flowOf, fixedSeverity(SeverityLow), "flow_changed", staticMessage("Flow narrative changed for %s.")),
	attrDiff(noteOf, fixedSeverity(SeverityInfo), "note_changed", staticMessage("Annotation note changed for %s.")),
}

func newFieldCompare(base Annotation, head Annotation) fieldCompare {
	fieldID := firstNonEmpty(head.FieldID, base.FieldID)
	producerService := firstNonEmpty(trimmed(head.SourceSystem), trimmed(base.SourceSystem))
	producerField := firstNonEmpty(trimmed(head.Field), trimmed(base.Field), fieldID)
	return fieldCompare{
		base:            base,
		head:            head,
		fieldID:         fieldID,
		producerService: producerService,
		producerEdge:    edgeOrNil(producerService, producerField),
		impactEdge:      edgeOrNil(producerService, fieldID),
	}
}

func compareField(base Annotation, head Annotation) []DriftChange {
	c := newFieldCompare(base, head)
	changes := make([]DriftChange, 0, len(fieldChecks))
	for _, check := range fieldChecks {
		changes = append(changes, check(c)...)
	}
	changes = append(changes, compareSources(c.fieldID, head.SourceSystem, head.Flow, base.Sources, head.Sources)...)
	return changes
}

func trimmed(s string) string { return strings.TrimSpace(s) }

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
