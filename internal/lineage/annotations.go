// annotations.go — Annotation types, comment scanning, and override parsing for data-lineage metadata.
package lineage

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	kvPairRe        = regexp.MustCompile(`([a-z_]+)=("([^"\\]|\\.)*"|'([^'\\]|\\.)*'|[^[:space:]]+)`)
	fieldIDRe       = regexp.MustCompile(`^[a-z][a-z0-9_]{2,63}$`)
	fieldPathRe     = regexp.MustCompile(`^[A-Za-z0-9_\-\[\]\.]+$`)
	systemIDRe      = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{0,63}(:[A-Za-z][A-Za-z0-9_-]{0,63})?$`)
	sourceVersionRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)
	ownerIDRe       = regexp.MustCompile(`^[a-z][a-z0-9_.-]{2,63}$`)
	escalationRefRe = regexp.MustCompile(`^[a-z]+:.+$`)
	providerIDRe    = regexp.MustCompile(`^[a-z][a-z0-9_-]{1,63}$`)
	changeTypeRe    = regexp.MustCompile(`^(\*|[a-z][a-z0-9_-]{1,63})$`)
	flowRe          = regexp.MustCompile(`(?i)^from @[A-Za-z][A-Za-z0-9_-]*(:[A-Za-z][A-Za-z0-9_-]*)?( (enriched|normalized|derived|validated|mapped|merged) @[A-Za-z][A-Za-z0-9_-]*(:[A-Za-z][A-Za-z0-9_-]*)?)*$`)
	identifierRe    = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)
)

const (
	defaultAnnotationSchemaVersion = "1"
	defaultTransformType           = "passthrough"
	defaultBreakPolicy             = "strict"
	defaultConfidence              = "declared"
	defaultDataClassification      = "internal"
	defaultIntroducedAt            = "1970-01-01"
	defaultNote                    = "defaulted_by=strict"
)

// Annotation describes lineage metadata for one API output field.
type Annotation struct {
	AnnotationSchemaVersion   string      `json:"annotation_schema_version"`
	FieldID                   string      `json:"field_id"`
	RenamedFrom               string      `json:"renamed_from,omitempty"`
	Field                     string      `json:"field"`
	SourceSystem              string      `json:"source_system"`
	SourceVersion             string      `json:"source_version"`
	MinSupportedSourceVersion string      `json:"min_supported_source_version"`
	TransformType             string      `json:"transform_type"`
	MergeStrategy             string      `json:"merge_strategy"`
	BreakPolicy               string      `json:"break_policy"`
	Confidence                string      `json:"confidence"`
	DataClassification        string      `json:"data_classification"`
	Owner                     string      `json:"owner"`
	Escalation                string      `json:"escalation"`
	ContractTestID            string      `json:"contract_test_id"`
	IntroducedAt              string      `json:"introduced_at"`
	SunsetAt                  string      `json:"sunset_at,omitempty"`
	Sources                   []SourceRef `json:"sources"`
	Flow                      string      `json:"flow"`
	Note                      string      `json:"note"`
	MappedFrom                []string    `json:"mapped_from,omitempty"`
	FilePath                  string      `json:"file_path,omitempty"`
	Line                      int         `json:"line"`
}

// SourceRef identifies one upstream source for a field.
//
// Grammar:
//
//	kind:target#path[@scope[!as_of]][?contract_ref=<ref>[&provider_id=<id>]]
//
// Examples:
//
//	api:identity.GetUser#response.id@cross_repo?contract_ref=git+https://...
//	api:spotify.GetTrack#response.track@external!2026-02-13?provider_id=spotify&contract_ref=https://developer.spotify.com/...
type SourceRef struct {
	Kind           string   `json:"kind"`
	Target         string   `json:"target"`
	Path           string   `json:"path"`
	Scope          string   `json:"scope"`
	AsOf           string   `json:"as_of,omitempty"`
	ProviderID     string   `json:"provider_id,omitempty"`
	UpstreamSystem string   `json:"upstream_system,omitempty"`
	ContractRef    string   `json:"contract_ref"`
	MappedFrom     []string `json:"mapped_from,omitempty"`
	Raw            string   `json:"raw"`
}

// Override is a comment-based temporary drift override.
type Override struct {
	FieldID    string `json:"field_id"`
	ChangeType string `json:"change_type"`
	Expires    string `json:"expires"`
	Reason     string `json:"reason"`
	Ticket     string `json:"ticket,omitempty"`
	FilePath   string `json:"file_path,omitempty"`
	Line       int    `json:"line"`
}

// ParseError reports invalid annotation syntax or semantics.
type ParseError struct {
	FilePath string `json:"file_path,omitempty"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
}

func (e ParseError) Error() string {
	if e.FilePath == "" {
		return fmt.Sprintf("line %d: %s", e.Line, e.Message)
	}
	return fmt.Sprintf("%s:%d: %s", e.FilePath, e.Line, e.Message)
}

// Parse extracts data-lineage annotations from source comments.
func Parse(source []byte) ([]Annotation, []ParseError) {
	annotations, _, errs := ParseWithOverrides(source)
	return annotations, errs
}

// ParseWithOverrides extracts both annotations and lineage override comments.
func ParseWithOverrides(source []byte) ([]Annotation, []Override, []ParseError) {
	lines := strings.Split(string(source), "\n")
	annotations := make([]Annotation, 0)
	overrides := make([]Override, 0)
	errors := make([]ParseError, 0)

	for i, line := range lines {
		lineNo := i + 1
		commentBody, ok := commentText(line)
		if !ok {
			continue
		}

		if payload, ok := annotationPayload(commentBody); ok {
			annotation, err := parsePayload(payload, lineNo, lines, i)
			if err != nil {
				errors = append(errors, *err)
				continue
			}
			annotations = append(annotations, annotation)
			continue
		}

		if payload, ok := overridePayload(commentBody); ok {
			override, err := parseOverridePayload(payload, lineNo)
			if err != nil {
				errors = append(errors, *err)
				continue
			}
			overrides = append(overrides, override)
		}
	}

	return annotations, overrides, errors
}

func commentText(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	switch {
	case strings.HasPrefix(trimmed, "//"):
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "//")), true
	case strings.HasPrefix(trimmed, "#"):
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "#")), true
	case strings.HasPrefix(trimmed, "/*") && strings.HasSuffix(trimmed, "*/"):
		inner := strings.TrimPrefix(trimmed, "/*")
		inner = strings.TrimSuffix(inner, "*/")
		return strings.TrimSpace(inner), true
	case strings.HasPrefix(trimmed, "*"):
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "*")), true
	default:
		return "", false
	}
}

func annotationPayload(comment string) (string, bool) {
	trimmed := strings.TrimSpace(comment)
	if strings.HasPrefix(trimmed, "strict-source") {
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "strict-source")), true
	}
	if strings.HasPrefix(trimmed, "strict:source") {
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "strict:source")), true
	}
	if strings.HasPrefix(trimmed, "stricture-source") {
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "stricture-source")), true
	}
	if strings.HasPrefix(trimmed, "stricture:source") {
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "stricture:source")), true
	}
	return "", false
}

func overridePayload(comment string) (string, bool) {
	trimmed := strings.TrimSpace(comment)
	if strings.HasPrefix(trimmed, "strict-lineage-override") {
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "strict-lineage-override")), true
	}
	if strings.HasPrefix(trimmed, "strict:lineage-override") {
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "strict:lineage-override")), true
	}
	if strings.HasPrefix(trimmed, "stricture-lineage-override") {
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "stricture-lineage-override")), true
	}
	if strings.HasPrefix(trimmed, "stricture:lineage-override") {
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "stricture:lineage-override")), true
	}
	return "", false
}

func parseOverridePayload(payload string, line int) (Override, *ParseError) {
	fields := map[string]string{}

	matches := kvPairRe.FindAllStringSubmatch(payload, -1)
	for _, match := range matches {
		key := strings.TrimSpace(match[1])
		value := strings.TrimSpace(match[2])
		fields[key] = unquote(value)
	}

	required := []string{"field_id", "change_type", "expires", "reason"}
	for _, key := range required {
		if strings.TrimSpace(fields[key]) == "" {
			return Override{}, parseErr(line, fmt.Sprintf("lineage override missing required key %q", key))
		}
	}

	fieldID := fields["field_id"]
	if !fieldIDRe.MatchString(fieldID) {
		return Override{}, parseErr(line, "lineage override field_id must match [a-z][a-z0-9_]{2,63}")
	}

	changeType := strings.TrimSpace(fields["change_type"])
	if !changeTypeRe.MatchString(changeType) {
		return Override{}, parseErr(line, "lineage override change_type must be '*' or snake_case token")
	}

	expires := strings.TrimSpace(fields["expires"])
	if !validDate(expires) {
		return Override{}, parseErr(line, "lineage override expires must use YYYY-MM-DD")
	}

	reason := strings.TrimSpace(fields["reason"])
	if reason == "" {
		return Override{}, parseErr(line, "lineage override reason cannot be empty")
	}

	return Override{
		FieldID:    fieldID,
		ChangeType: changeType,
		Expires:    expires,
		Reason:     reason,
		Ticket:     strings.TrimSpace(fields["ticket"]),
		Line:       line,
	}, nil
}

func validDate(value string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}

func validEnum(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func parseErr(line int, message string) *ParseError {
	return &ParseError{Line: line, Message: message}
}

func unquote(value string) string {
	value = strings.TrimSpace(value)
	if len(value) < 2 {
		return value
	}

	first := rune(value[0])
	last := rune(value[len(value)-1])
	if first == last && (first == '"' || first == '\'') {
		return value[1 : len(value)-1]
	}
	return value
}

func uniqueSortedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		v := strings.TrimSpace(value)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	if len(out) == 0 {
		return nil
	}
	sort.Strings(out)
	return out
}
