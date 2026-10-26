// annotations_compact.go — Parses compact-form lineage annotations and infers their missing keys.
package lineage

import (
	"fmt"
	"strings"
	"time"
)

// compactKeyMap maps compact-form chunk keys to canonical field keys.
var compactKeyMap = map[string]string{
	"source_version": "source_version",
	"transform":      "transform_type",
	"merge":          "merge_strategy",
	"classification": "data_classification",
	"break_policy":   "break_policy",
	"scope":          "__scope",
	"provider":       "__provider_id",
}

func parseCompactPayload(payload string) (map[string]string, bool) {
	raw := strings.TrimSpace(payload)
	raw = strings.TrimSpace(strings.TrimPrefix(raw, ":"))
	if raw == "" || strings.Contains(raw, "=") {
		return nil, false
	}
	if strings.HasPrefix(strings.ToLower(raw), "from ") {
		raw = strings.TrimSpace(raw[5:])
	}
	if raw == "" {
		return nil, false
	}

	parts := strings.Split(raw, ",")
	sourceSystem := strings.TrimSpace(parts[0])
	if sourceSystem == "" {
		return nil, false
	}

	fields := map[string]string{
		"source_system": sourceSystem,
	}
	for _, part := range parts[1:] {
		applyCompactChunk(fields, strings.TrimSpace(part))
	}
	return fields, true
}

// applyCompactChunk records one "key value" chunk; malformed chunks are ignored.
func applyCompactChunk(fields map[string]string, chunk string) {
	if chunk == "" {
		return
	}
	key, value, ok := strings.Cut(chunk, " ")
	if !ok {
		return
	}
	key = strings.ToLower(strings.TrimSpace(key))
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	if key == "note" {
		fields["note"] = unquote(value)
		return
	}
	if canonical, known := compactKeyMap[key]; known {
		fields[canonical] = value
	}
}

func applyCompactDefaults(fields map[string]string, lines []string, lineIndex int) {
	if strings.TrimSpace(fields["source_system"]) == "" {
		return
	}

	if isBlank(fields, "field") && isBlank(fields, "field_id") {
		if inferred := inferFieldFromNextLine(lines, lineIndex); inferred != "" {
			fields["field"] = inferred
		}
	}
	if isBlank(fields, "field") && !isBlank(fields, "field_id") {
		fields["field"] = deriveFieldPath(strings.TrimSpace(fields["field_id"]))
	}
	if isBlank(fields, "field_id") && !isBlank(fields, "field") {
		fields["field_id"] = deriveFieldID(strings.TrimSpace(fields["field"]))
	}

	if isBlank(fields, "source_version") {
		fields["source_version"] = "v1"
	}
	if isBlank(fields, "min_supported_source_version") {
		fields["min_supported_source_version"] = fields["source_version"]
	}
	if isBlank(fields, "sources") && !isBlank(fields, "field") && !isBlank(fields, "field_id") {
		fields["sources"] = compactSourceRef(fields)
	}
}

func isBlank(fields map[string]string, key string) bool {
	return strings.TrimSpace(fields[key]) == ""
}

func compactSourceRef(fields map[string]string) string {
	system := strings.TrimSpace(fields["source_system"])
	field := strings.TrimSpace(fields["field"])
	fieldID := strings.TrimSpace(fields["field_id"])
	systemSlug := slugifySystemID(system)
	target := systemSlug + ".emitter"
	contractRef := fmt.Sprintf("internal://strict/%s/%s", systemSlug, fieldID)

	scope := strings.TrimSpace(fields["__scope"])
	if scope == "" {
		scope = "internal"
	}

	if scope == "external" {
		providerID := strings.TrimSpace(fields["__provider_id"])
		if providerID == "" {
			providerID = systemSlug
		}
		asOf := time.Now().UTC().Format("2006-01-02")
		return fmt.Sprintf("api:%s#%s@external!%s?provider_id=%s&contract_ref=%s", target, field, asOf, providerID, contractRef)
	}

	return fmt.Sprintf("api:%s#%s@%s?contract_ref=%s", target, field, scope, contractRef)
}

func inferFieldFromNextLine(lines []string, lineIndex int) string {
	if lineIndex < 0 || lineIndex >= len(lines)-1 {
		return ""
	}
	limit := lineIndex + 4
	if limit >= len(lines) {
		limit = len(lines) - 1
	}
	for i := lineIndex + 1; i <= limit; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || isCommentLine(trimmed) {
			continue
		}
		if token := extractFieldToken(trimmed); token != "" {
			return "response." + token
		}
	}
	return ""
}

func isCommentLine(trimmed string) bool {
	return strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*")
}

// fieldTokenKeywords are declaration keywords skipped when inferring a field name.
var fieldTokenKeywords = map[string]bool{
	"type": true, "struct": true, "func": true, "var": true, "const": true, "let": true,
	"return": true, "class": true, "interface": true, "public": true, "private": true,
	"protected": true, "static": true, "final": true, "package": true, "import": true, "new": true,
}

func extractFieldToken(line string) string {
	if split := strings.SplitN(line, "//", 2); len(split) > 0 {
		line = strings.TrimSpace(split[0])
	}
	if hash := strings.Index(line, "#"); hash >= 0 {
		line = strings.TrimSpace(line[:hash])
	}
	if line == "" {
		return ""
	}

	for _, ident := range identifierRe.FindAllString(line, -1) {
		if fieldTokenKeywords[strings.ToLower(ident)] {
			continue
		}
		return toSnakeCase(ident)
	}
	return ""
}
