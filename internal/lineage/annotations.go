// annotations.go - Parser for Stricture data-lineage annotations.
package lineage

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	kvPairRe        = regexp.MustCompile(`([a-z_]+)=("([^"\\]|\\.)*"|'([^'\\]|\\.)*'|[^[:space:]]+)`)
	fieldIDRe       = regexp.MustCompile(`^[a-z][a-z0-9_]{2,63}$`)
	fieldPathRe     = regexp.MustCompile(`^[A-Za-z0-9_\-\[\]\.]+$`)
	sourceSystemRe  = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{0,63}$`)
	sourceVersionRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)
	ownerIDRe       = regexp.MustCompile(`^[a-z][a-z0-9_.-]{2,63}$`)
	escalationRefRe = regexp.MustCompile(`^[a-z]+:.+$`)
	providerIDRe    = regexp.MustCompile(`^[a-z][a-z0-9_-]{1,63}$`)
	changeTypeRe    = regexp.MustCompile(`^(\*|[a-z][a-z0-9_-]{1,63})$`)
	flowRe          = regexp.MustCompile(`(?i)^from @[A-Za-z][A-Za-z0-9_-]*( (enriched|normalized|derived|validated|mapped|merged) @[A-Za-z][A-Za-z0-9_-]*)*$`)
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
	Kind           string `json:"kind"`
	Target         string `json:"target"`
	Path           string `json:"path"`
	Scope          string `json:"scope"`
	AsOf           string `json:"as_of,omitempty"`
	ProviderID     string `json:"provider_id,omitempty"`
	UpstreamSystem string `json:"upstream_system,omitempty"`
	ContractRef    string `json:"contract_ref"`
	Raw            string `json:"raw"`
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
			annotation, err := parsePayload(payload, lineNo)
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
	if strings.HasPrefix(trimmed, "stricture-lineage-override") {
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "stricture-lineage-override")), true
	}
	if strings.HasPrefix(trimmed, "stricture:lineage-override") {
		return strings.TrimSpace(strings.TrimPrefix(trimmed, "stricture:lineage-override")), true
	}
	return "", false
}

func parsePayload(payload string, line int) (Annotation, *ParseError) {
	fields := map[string]string{}

	matches := kvPairRe.FindAllStringSubmatch(payload, -1)
	for _, match := range matches {
		key := strings.TrimSpace(match[1])
		value := strings.TrimSpace(match[2])
		fields[key] = unquote(value)
	}

	required := []string{
		"annotation_schema_version",
		"field_id",
		"field",
		"source_system",
		"source_version",
		"min_supported_source_version",
		"transform_type",
		"merge_strategy",
		"break_policy",
		"confidence",
		"data_classification",
		"owner",
		"escalation",
		"contract_test_id",
		"introduced_at",
		"sources",
		"flow",
		"note",
	}
	for _, key := range required {
		if strings.TrimSpace(fields[key]) == "" {
			return Annotation{}, parseErr(line, fmt.Sprintf("missing required key %q", key))
		}
	}

	if fields["annotation_schema_version"] != "1" {
		return Annotation{}, parseErr(line, "annotation_schema_version must be '1'")
	}

	fieldID := fields["field_id"]
	if !fieldIDRe.MatchString(fieldID) {
		return Annotation{}, parseErr(line, "field_id must match [a-z][a-z0-9_]{2,63}")
	}

	renamedFrom := fields["renamed_from"]
	if renamedFrom != "" {
		if !fieldIDRe.MatchString(renamedFrom) {
			return Annotation{}, parseErr(line, "renamed_from must match [a-z][a-z0-9_]{2,63}")
		}
		if renamedFrom == fieldID {
			return Annotation{}, parseErr(line, "renamed_from cannot equal field_id")
		}
	}

	fieldPath := fields["field"]
	if !fieldPathRe.MatchString(fieldPath) {
		return Annotation{}, parseErr(line, "field must match [A-Za-z0-9_-.[]]")
	}

	if !sourceSystemRe.MatchString(fields["source_system"]) {
		return Annotation{}, parseErr(line, "source_system must match [A-Za-z][A-Za-z0-9_-]{0,63}")
	}

	if !sourceVersionRe.MatchString(fields["source_version"]) {
		return Annotation{}, parseErr(line, "source_version must be non-empty and use [A-Za-z0-9._-]")
	}
	if !sourceVersionRe.MatchString(fields["min_supported_source_version"]) {
		return Annotation{}, parseErr(line, "min_supported_source_version must be non-empty and use [A-Za-z0-9._-]")
	}

	transformType := fields["transform_type"]
	if !validEnum(transformType, "passthrough", "normalize", "derive", "aggregate", "mask", "join") {
		return Annotation{}, parseErr(line, "transform_type must be one of passthrough|normalize|derive|aggregate|mask|join")
	}

	mergeStrategy := fields["merge_strategy"]
	if !validEnum(mergeStrategy, "single_source", "priority", "first_non_null", "union", "custom") {
		return Annotation{}, parseErr(line, "merge_strategy must be one of single_source|priority|first_non_null|union|custom")
	}

	breakPolicy := fields["break_policy"]
	if !validEnum(breakPolicy, "additive_only", "strict", "opaque") {
		return Annotation{}, parseErr(line, "break_policy must be one of additive_only|strict|opaque")
	}

	confidence := fields["confidence"]
	if !validEnum(confidence, "declared", "inferred") {
		return Annotation{}, parseErr(line, "confidence must be one of declared|inferred")
	}

	classification := fields["data_classification"]
	if !validEnum(classification, "public", "internal", "sensitive", "regulated") {
		return Annotation{}, parseErr(line, "data_classification must be one of public|internal|sensitive|regulated")
	}

	if !ownerIDRe.MatchString(fields["owner"]) {
		return Annotation{}, parseErr(line, "owner must be a normalized team ID like team.identity")
	}

	if !escalationRefRe.MatchString(fields["escalation"]) {
		return Annotation{}, parseErr(line, "escalation must be a typed ref like slack:#channel or pagerduty:service")
	}

	if strings.TrimSpace(fields["contract_test_id"]) == "" {
		return Annotation{}, parseErr(line, "contract_test_id cannot be empty")
	}

	if !validDate(fields["introduced_at"]) {
		return Annotation{}, parseErr(line, "introduced_at must use YYYY-MM-DD")
	}

	sunset := fields["sunset_at"]
	if sunset != "" && !validDate(sunset) {
		return Annotation{}, parseErr(line, "sunset_at must use YYYY-MM-DD")
	}
	if sunset != "" {
		introducedT, err := time.Parse("2006-01-02", fields["introduced_at"])
		if err != nil {
			return Annotation{}, parseErr(line, "introduced_at must use YYYY-MM-DD")
		}
		sunsetT, err := time.Parse("2006-01-02", sunset)
		if err != nil {
			return Annotation{}, parseErr(line, "sunset_at must use YYYY-MM-DD")
		}
		if sunsetT.Before(introducedT) {
			return Annotation{}, parseErr(line, "sunset_at must be >= introduced_at")
		}
	}

	if !flowRe.MatchString(strings.TrimSpace(fields["flow"])) {
		return Annotation{}, parseErr(line, "flow must look like 'from @X enriched @self' (verbs: enriched|normalized|derived|validated|mapped|merged)")
	}

	if strings.TrimSpace(fields["note"]) == "" {
		return Annotation{}, parseErr(line, "note cannot be empty")
	}

	sourceRefs, err := parseSources(fields["sources"], line)
	if err != nil {
		return Annotation{}, err
	}

	if len(sourceRefs) > 1 && mergeStrategy == "single_source" {
		return Annotation{}, parseErr(line, "merge_strategy=single_source is invalid when multiple sources are declared")
	}
	if len(sourceRefs) == 1 && mergeStrategy != "single_source" {
		return Annotation{}, parseErr(line, "merge_strategy must be single_source when only one source is declared")
	}

	return Annotation{
		AnnotationSchemaVersion:   fields["annotation_schema_version"],
		FieldID:                   fieldID,
		RenamedFrom:               renamedFrom,
		Field:                     fieldPath,
		SourceSystem:              fields["source_system"],
		SourceVersion:             fields["source_version"],
		MinSupportedSourceVersion: fields["min_supported_source_version"],
		TransformType:             transformType,
		MergeStrategy:             mergeStrategy,
		BreakPolicy:               breakPolicy,
		Confidence:                confidence,
		DataClassification:        classification,
		Owner:                     fields["owner"],
		Escalation:                fields["escalation"],
		ContractTestID:            fields["contract_test_id"],
		IntroducedAt:              fields["introduced_at"],
		SunsetAt:                  sunset,
		Sources:                   sourceRefs,
		Flow:                      fields["flow"],
		Note:                      fields["note"],
		Line:                      line,
	}, nil
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

func parseSources(raw string, line int) ([]SourceRef, *ParseError) {
	parts := strings.Split(raw, ",")
	sources := make([]SourceRef, 0, len(parts))
	for _, part := range parts {
		ref := strings.TrimSpace(part)
		if ref == "" {
			continue
		}

		parsed, err := parseSourceRef(ref, line)
		if err != nil {
			return nil, err
		}
		sources = append(sources, parsed)
	}

	if len(sources) == 0 {
		return nil, parseErr(line, "sources must contain at least one source reference")
	}
	return sources, nil
}

func parseSourceRef(ref string, line int) (SourceRef, *ParseError) {
	core := ref
	queryRaw := ""
	if before, after, ok := strings.Cut(ref, "?"); ok {
		core = before
		queryRaw = after
	}

	kindSplit := strings.SplitN(core, ":", 2)
	if len(kindSplit) != 2 {
		return SourceRef{}, parseErr(line, fmt.Sprintf("invalid source reference %q (expected kind:target#path@scope!as_of)", ref))
	}

	kind := strings.ToLower(strings.TrimSpace(kindSplit[0]))
	if !validEnum(kind, "api", "input", "db", "event", "file", "cache") {
		return SourceRef{}, parseErr(line, fmt.Sprintf("unsupported source kind %q", kind))
	}

	targetPathSplit := strings.SplitN(kindSplit[1], "#", 2)
	if len(targetPathSplit) != 2 {
		return SourceRef{}, parseErr(line, fmt.Sprintf("invalid source reference %q (missing #path)", ref))
	}

	target := strings.TrimSpace(targetPathSplit[0])
	pathAndMeta := strings.TrimSpace(targetPathSplit[1])
	if target == "" || pathAndMeta == "" {
		return SourceRef{}, parseErr(line, fmt.Sprintf("invalid source reference %q (target/path cannot be empty)", ref))
	}

	path := pathAndMeta
	scope := "internal"
	asOf := ""

	if at := strings.LastIndex(pathAndMeta, "@"); at >= 0 {
		path = strings.TrimSpace(pathAndMeta[:at])
		scopeAndAsOf := strings.TrimSpace(pathAndMeta[at+1:])
		if scopeAndAsOf == "" {
			return SourceRef{}, parseErr(line, fmt.Sprintf("invalid source reference %q (missing scope after @)", ref))
		}

		if bang := strings.Index(scopeAndAsOf, "!"); bang >= 0 {
			scope = strings.TrimSpace(scopeAndAsOf[:bang])
			asOf = strings.TrimSpace(scopeAndAsOf[bang+1:])
		} else {
			scope = strings.TrimSpace(scopeAndAsOf)
		}
	}

	if !validEnum(scope, "internal", "cross_repo", "external") {
		return SourceRef{}, parseErr(line, fmt.Sprintf("unsupported source scope %q (use internal|cross_repo|external)", scope))
	}

	if scope == "external" {
		if !validDate(asOf) {
			return SourceRef{}, parseErr(line, fmt.Sprintf("external source %q must include as_of date YYYY-MM-DD", ref))
		}
	} else if asOf != "" {
		return SourceRef{}, parseErr(line, fmt.Sprintf("source %q includes as_of date but scope is %q", ref, scope))
	}

	query, err := parseQuery(queryRaw)
	if err != nil {
		return SourceRef{}, parseErr(line, fmt.Sprintf("invalid source reference %q (%s)", ref, err.Error()))
	}

	contractRef := strings.TrimSpace(query["contract_ref"])
	if contractRef == "" {
		return SourceRef{}, parseErr(line, fmt.Sprintf("source %q must include contract_ref", ref))
	}

	providerID := strings.TrimSpace(query["provider_id"])
	upstreamSystem := strings.TrimSpace(query["upstream_system"])
	if scope == "external" {
		if providerID == "" {
			return SourceRef{}, parseErr(line, fmt.Sprintf("external source %q must include provider_id", ref))
		}
		if !providerIDRe.MatchString(providerID) {
			return SourceRef{}, parseErr(line, fmt.Sprintf("external source %q has invalid provider_id", ref))
		}
	} else if providerID != "" {
		return SourceRef{}, parseErr(line, fmt.Sprintf("source %q includes provider_id but scope is %q", ref, scope))
	}
	if upstreamSystem != "" && !sourceSystemRe.MatchString(upstreamSystem) {
		return SourceRef{}, parseErr(line, fmt.Sprintf("source %q has invalid upstream_system", ref))
	}

	return SourceRef{
		Kind:           kind,
		Target:         target,
		Path:           path,
		Scope:          scope,
		AsOf:           asOf,
		ProviderID:     providerID,
		UpstreamSystem: upstreamSystem,
		ContractRef:    contractRef,
		Raw:            ref,
	}, nil
}

func parseQuery(raw string) (map[string]string, error) {
	result := map[string]string{}
	if strings.TrimSpace(raw) == "" {
		return result, nil
	}

	pairs := strings.Split(raw, "&")
	for _, pair := range pairs {
		if strings.TrimSpace(pair) == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("query pair %q must be key=value", pair)
		}
		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])
		if k == "" || v == "" {
			return nil, fmt.Errorf("query pair %q must not have empty key/value", pair)
		}
		decoded, err := url.QueryUnescape(v)
		if err != nil {
			return nil, fmt.Errorf("query value for %q is not valid encoding", k)
		}
		result[k] = decoded
	}
	return result, nil
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
