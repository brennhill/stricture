// annotations_defaults.go — Default values and identifier derivation for lineage annotations.
package lineage

import (
	"strings"
	"unicode"
)

// staticDefaults are applied in order when the key is blank.
var staticDefaults = []struct{ key, value string }{
	{"transform_type", defaultTransformType},
	{"break_policy", defaultBreakPolicy},
	{"confidence", defaultConfidence},
	{"data_classification", defaultDataClassification},
	{"introduced_at", defaultIntroducedAt},
	{"note", defaultNote},
}

// setDefault fills key with value when the current value is blank.
func setDefault(fields map[string]string, key, value string) {
	if isBlank(fields, key) {
		fields[key] = value
	}
}

func applyAnnotationDefaults(fields map[string]string) {
	setDefault(fields, "annotation_schema_version", defaultAnnotationSchemaVersion)
	deriveMissingFieldKeys(fields)
	if !isBlank(fields, "source_version") {
		setDefault(fields, "min_supported_source_version", strings.TrimSpace(fields["source_version"]))
	}
	for _, d := range staticDefaults {
		setDefault(fields, d.key, d.value)
	}
	setDefault(fields, "merge_strategy", defaultMergeStrategy(fields["sources"]))

	systemSlug := slugifySystemID(fields["source_system"])
	setDefault(fields, "owner", "team."+systemSlug)
	setDefault(fields, "escalation", "slack:#"+systemSlug+"-oncall")
	if !isBlank(fields, "field_id") {
		setDefault(fields, "contract_test_id", "ci://contracts/"+systemSlug+"/"+strings.TrimSpace(fields["field_id"]))
	}
	if !isBlank(fields, "source_system") {
		setDefault(fields, "flow", "from @"+strings.TrimSpace(fields["source_system"])+" mapped @self")
	}
}

// deriveMissingFieldKeys derives field_id from field, or field from field_id.
func deriveMissingFieldKeys(fields map[string]string) {
	field := strings.TrimSpace(fields["field"])
	fieldID := strings.TrimSpace(fields["field_id"])
	if fieldID == "" && field != "" {
		fields["field_id"] = deriveFieldID(field)
	}
	if field == "" && fieldID != "" {
		fields["field"] = deriveFieldPath(fieldID)
	}
}

func defaultMergeStrategy(sources string) string {
	if hasMultipleSources(sources) {
		return "priority"
	}
	return "single_source"
}

func hasMultipleSources(raw string) bool {
	count := 0
	for _, part := range strings.Split(raw, ",") {
		if strings.TrimSpace(part) == "" {
			continue
		}
		count++
		if count > 1 {
			return true
		}
	}
	return false
}

func toSnakeCase(value string) string {
	var b strings.Builder
	lastUnderscore := false
	for i, r := range value {
		switch {
		case unicode.IsUpper(r):
			if i > 0 && !lastUnderscore {
				b.WriteRune('_')
			}
			b.WriteRune(unicode.ToLower(r))
			lastUnderscore = false
		case unicode.IsLower(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
			lastUnderscore = false
		default:
			if !lastUnderscore {
				b.WriteRune('_')
				lastUnderscore = true
			}
		}
	}
	return strings.Trim(b.String(), "_")
}

func deriveFieldID(field string) string {
	var b strings.Builder
	b.Grow(len(field))
	lastUnderscore := false
	for _, r := range field {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(unicode.ToLower(r))
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			b.WriteRune('_')
			lastUnderscore = true
		}
	}
	return clampFieldID(strings.Trim(b.String(), "_"))
}

// clampFieldID enforces the field_id shape: leading letter, 3..64 chars.
func clampFieldID(out string) string {
	if out == "" {
		out = "field_unknown"
	}
	if out[0] < 'a' || out[0] > 'z' {
		out = "f_" + out
	}
	if len(out) < 3 {
		out += "_id"
	}
	if len(out) > 64 {
		out = strings.TrimRight(out[:64], "_")
		if len(out) < 3 {
			out = "field_unknown"
		}
	}
	return out
}

func deriveFieldPath(fieldID string) string {
	return strings.ReplaceAll(strings.TrimSpace(fieldID), "_", ".")
}

func slugifySystemID(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "unknown"
	}

	var b strings.Builder
	lastDash := false
	prevAlphaNum := false
	prevLowerOrDigit := false

	for _, r := range trimmed {
		switch {
		case unicode.IsLetter(r) || unicode.IsNumber(r):
			isUpper := unicode.IsUpper(r)
			if isUpper && prevLowerOrDigit && !lastDash {
				b.WriteRune('-')
			}
			b.WriteRune(unicode.ToLower(r))
			lastDash = false
			prevAlphaNum = true
			prevLowerOrDigit = unicode.IsLower(r) || unicode.IsNumber(r)
		default:
			if prevAlphaNum && !lastDash {
				b.WriteRune('-')
				lastDash = true
			}
			prevAlphaNum = false
			prevLowerOrDigit = false
		}
	}

	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = "unknown"
	}
	if !unicode.IsLetter(rune(out[0])) {
		out = "sys-" + out
	}
	return out
}
