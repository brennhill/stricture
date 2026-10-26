// annotations_payload.go — Parses and validates one lineage annotation payload into an Annotation.
package lineage

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// payloadRule is one ordered validation check over the normalized field map.
type payloadRule struct {
	ok      func(fields map[string]string) bool
	message string
}

func matchesRe(key string, re *regexp.Regexp) func(map[string]string) bool {
	return func(f map[string]string) bool { return re.MatchString(f[key]) }
}

func matchesReTrimmed(key string, re *regexp.Regexp) func(map[string]string) bool {
	return func(f map[string]string) bool { return re.MatchString(strings.TrimSpace(f[key])) }
}

func inEnum(key string, allowed ...string) func(map[string]string) bool {
	return func(f map[string]string) bool { return validEnum(f[key], allowed...) }
}

func notBlank(key string) func(map[string]string) bool {
	return func(f map[string]string) bool { return strings.TrimSpace(f[key]) != "" }
}

func validDateKey(key string) func(map[string]string) bool {
	return func(f map[string]string) bool { return validDate(f[key]) }
}

func optionalDateKey(key string) func(map[string]string) bool {
	return func(f map[string]string) bool { return f[key] == "" || validDate(f[key]) }
}

func optionalFieldID(key string) func(map[string]string) bool {
	return func(f map[string]string) bool { return f[key] == "" || fieldIDRe.MatchString(f[key]) }
}

func renamedDiffersFromFieldID(f map[string]string) bool {
	return f["renamed_from"] == "" || f["renamed_from"] != f["field_id"]
}

func sunsetNotBeforeIntroduced(f map[string]string) bool {
	if f["sunset_at"] == "" {
		return true
	}
	introduced, err := time.Parse("2006-01-02", f["introduced_at"])
	if err != nil {
		return false
	}
	sunset, err := time.Parse("2006-01-02", f["sunset_at"])
	if err != nil {
		return false
	}
	return !sunset.Before(introduced)
}

func schemaVersionIsDefault(f map[string]string) bool {
	return f["annotation_schema_version"] == defaultAnnotationSchemaVersion
}

// payloadRules are evaluated in order; the first failure is reported.
var payloadRules = []payloadRule{
	{schemaVersionIsDefault, "annotation_schema_version must be '1'"},
	{matchesRe("field_id", fieldIDRe), "field_id must match [a-z][a-z0-9_]{2,63}"},
	{optionalFieldID("renamed_from"), "renamed_from must match [a-z][a-z0-9_]{2,63}"},
	{renamedDiffersFromFieldID, "renamed_from cannot equal field_id"},
	{matchesRe("field", fieldPathRe), "field must match [A-Za-z0-9_-.[]]"},
	{matchesRe("source_system", systemIDRe), "source_system must match [A-Za-z][A-Za-z0-9_-]{0,63}(:[A-Za-z][A-Za-z0-9_-]{0,63})?"},
	{matchesRe("source_version", sourceVersionRe), "source_version must be non-empty and use [A-Za-z0-9._-]"},
	{matchesRe("min_supported_source_version", sourceVersionRe), "min_supported_source_version must be non-empty and use [A-Za-z0-9._-]"},
	{inEnum("transform_type", "passthrough", "normalize", "derive", "aggregate", "mask", "join"), "transform_type must be one of passthrough|normalize|derive|aggregate|mask|join"},
	{inEnum("merge_strategy", "single_source", "priority", "first_non_null", "union", "custom"), "merge_strategy must be one of single_source|priority|first_non_null|union|custom"},
	{inEnum("break_policy", "additive_only", "strict", "opaque"), "break_policy must be one of additive_only|strict|opaque"},
	{inEnum("confidence", "declared", "inferred"), "confidence must be one of declared|inferred"},
	{inEnum("data_classification", "public", "internal", "sensitive", "regulated"), "data_classification must be one of public|internal|sensitive|regulated"},
	{matchesRe("owner", ownerIDRe), "owner must be a normalized team ID like team.identity"},
	{matchesRe("escalation", escalationRefRe), "escalation must be a typed ref like slack:#channel or pagerduty:service"},
	{notBlank("contract_test_id"), "contract_test_id cannot be empty"},
	{validDateKey("introduced_at"), "introduced_at must use YYYY-MM-DD"},
	{optionalDateKey("sunset_at"), "sunset_at must use YYYY-MM-DD"},
	{sunsetNotBeforeIntroduced, "sunset_at must be >= introduced_at"},
	{matchesReTrimmed("flow", flowRe), "flow must look like 'from @X enriched @self' (verbs: enriched|normalized|derived|validated|mapped|merged)"},
	{notBlank("note"), "note cannot be empty"},
}

func parsePayload(payload string, line int, lines []string, lineIndex int) (Annotation, *ParseError) {
	fields, compactMode := collectPayloadFields(payload)
	mappedFrom, aliasErr := normalizeAnnotationAliases(fields)
	if aliasErr != nil {
		return Annotation{}, parseErr(line, aliasErr.Error())
	}

	if compactMode {
		applyCompactDefaults(fields, lines, lineIndex)
	}
	applyAnnotationDefaults(fields)

	if perr := validateRequiredKeys(fields, line, compactMode); perr != nil {
		return Annotation{}, perr
	}
	for _, rule := range payloadRules {
		if !rule.ok(fields) {
			return Annotation{}, parseErr(line, rule.message)
		}
	}

	sourceRefs, perr := parseSources(fields["sources"], line)
	if perr != nil {
		return Annotation{}, perr
	}
	for _, src := range sourceRefs {
		mappedFrom = append(mappedFrom, src.MappedFrom...)
	}
	if perr := validateMergeStrategy(fields["merge_strategy"], len(sourceRefs), line); perr != nil {
		return Annotation{}, perr
	}

	return buildAnnotation(fields, sourceRefs, uniqueSortedStrings(mappedFrom), line), nil
}

// collectPayloadFields extracts key=value pairs, or compact-form fields when none exist.
func collectPayloadFields(payload string) (map[string]string, bool) {
	fields := map[string]string{}
	matches := kvPairRe.FindAllStringSubmatch(payload, -1)
	if len(matches) == 0 {
		compactFields, ok := parseCompactPayload(payload)
		for key, value := range compactFields {
			fields[key] = value
		}
		return fields, ok
	}
	for _, match := range matches {
		fields[strings.TrimSpace(match[1])] = unquote(strings.TrimSpace(match[2]))
	}
	return fields, false
}

func validateRequiredKeys(fields map[string]string, line int, compactMode bool) *ParseError {
	if strings.TrimSpace(fields["field_id"]) == "" && strings.TrimSpace(fields["field"]) == "" {
		return parseErr(line, "missing required key \"field\" or \"field_id\"")
	}
	required := []string{"source_system", "source_version", "sources"}
	if compactMode {
		required = []string{"source_system", "field"}
	}
	for _, key := range required {
		if strings.TrimSpace(fields[key]) == "" {
			return parseErr(line, fmt.Sprintf("missing required key %q", key))
		}
	}
	return nil
}

func validateMergeStrategy(mergeStrategy string, sourceCount, line int) *ParseError {
	if sourceCount > 1 && mergeStrategy == "single_source" {
		return parseErr(line, "merge_strategy=single_source is invalid when multiple sources are declared")
	}
	if sourceCount == 1 && mergeStrategy != "single_source" {
		return parseErr(line, "merge_strategy must be single_source when only one source is declared")
	}
	return nil
}

func buildAnnotation(fields map[string]string, sources []SourceRef, mappedFrom []string, line int) Annotation {
	return Annotation{
		AnnotationSchemaVersion:   fields["annotation_schema_version"],
		FieldID:                   fields["field_id"],
		RenamedFrom:               fields["renamed_from"],
		Field:                     fields["field"],
		SourceSystem:              fields["source_system"],
		SourceVersion:             fields["source_version"],
		MinSupportedSourceVersion: fields["min_supported_source_version"],
		TransformType:             fields["transform_type"],
		MergeStrategy:             fields["merge_strategy"],
		BreakPolicy:               fields["break_policy"],
		Confidence:                fields["confidence"],
		DataClassification:        fields["data_classification"],
		Owner:                     fields["owner"],
		Escalation:                fields["escalation"],
		ContractTestID:            fields["contract_test_id"],
		IntroducedAt:              fields["introduced_at"],
		SunsetAt:                  fields["sunset_at"],
		Sources:                   sources,
		Flow:                      fields["flow"],
		Note:                      fields["note"],
		MappedFrom:                mappedFrom,
		Line:                      line,
	}
}
