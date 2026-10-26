// annotations_sources.go — Parses source references, query strings, and key aliases for lineage annotations.
package lineage

import (
	"fmt"
	"net/url"
	"strings"
)

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
	core, queryRaw, _ := strings.Cut(ref, "?")

	kind, target, pathAndMeta, perr := splitSourceHead(ref, core, line)
	if perr != nil {
		return SourceRef{}, perr
	}

	path, scope, asOf, perr := splitPathScope(ref, pathAndMeta, line)
	if perr != nil {
		return SourceRef{}, perr
	}

	query, err := parseQuery(queryRaw)
	if err != nil {
		return SourceRef{}, parseErr(line, fmt.Sprintf("invalid source reference %q (%s)", ref, err.Error()))
	}
	queryMappedFrom, err := normalizeSourceQueryAliases(query)
	if err != nil {
		return SourceRef{}, parseErr(line, fmt.Sprintf("invalid source reference %q (%s)", ref, err.Error()))
	}

	asOf, perr = resolveAsOf(ref, scope, asOf, query, line)
	if perr != nil {
		return SourceRef{}, perr
	}

	contractRef := strings.TrimSpace(query["contract_ref"])
	if contractRef == "" {
		return SourceRef{}, parseErr(line, fmt.Sprintf("source %q must include contract_ref", ref))
	}

	providerID, upstreamSystem, perr := resolveProvider(ref, scope, query, line)
	if perr != nil {
		return SourceRef{}, perr
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
		MappedFrom:     queryMappedFrom,
		Raw:            ref,
	}, nil
}

// splitSourceHead splits "kind:target#pathAndMeta" and validates kind and emptiness.
func splitSourceHead(ref, core string, line int) (kind, target, pathAndMeta string, perr *ParseError) {
	kindSplit := strings.SplitN(core, ":", 2)
	if len(kindSplit) != 2 {
		return "", "", "", parseErr(line, fmt.Sprintf("invalid source reference %q (expected kind:target#path@scope!as_of)", ref))
	}

	kind = strings.ToLower(strings.TrimSpace(kindSplit[0]))
	if !validEnum(kind, "api", "input", "db", "event", "file", "cache") {
		return "", "", "", parseErr(line, fmt.Sprintf("unsupported source kind %q", kind))
	}

	targetPathSplit := strings.SplitN(kindSplit[1], "#", 2)
	if len(targetPathSplit) != 2 {
		return "", "", "", parseErr(line, fmt.Sprintf("invalid source reference %q (missing #path)", ref))
	}

	target = strings.TrimSpace(targetPathSplit[0])
	pathAndMeta = strings.TrimSpace(targetPathSplit[1])
	if target == "" || pathAndMeta == "" {
		return "", "", "", parseErr(line, fmt.Sprintf("invalid source reference %q (target/path cannot be empty)", ref))
	}
	return kind, target, pathAndMeta, nil
}

// splitPathScope splits "path@scope!as_of" and validates the scope enum.
func splitPathScope(ref, pathAndMeta string, line int) (path, scope, asOf string, perr *ParseError) {
	path = pathAndMeta
	scope = "internal"

	if at := strings.LastIndex(pathAndMeta, "@"); at >= 0 {
		path = strings.TrimSpace(pathAndMeta[:at])
		scopeAndAsOf := strings.TrimSpace(pathAndMeta[at+1:])
		if scopeAndAsOf == "" {
			return "", "", "", parseErr(line, fmt.Sprintf("invalid source reference %q (missing scope after @)", ref))
		}
		before, after, found := strings.Cut(scopeAndAsOf, "!")
		scope = strings.TrimSpace(before)
		if found {
			asOf = strings.TrimSpace(after)
		}
	}

	if !validEnum(scope, "internal", "cross_repo", "external") {
		return "", "", "", parseErr(line, fmt.Sprintf("unsupported source scope %q (use internal|cross_repo|external)", scope))
	}
	return path, scope, asOf, nil
}

// resolveAsOf merges the inline as_of with the query alias and enforces scope rules.
func resolveAsOf(ref, scope, asOf string, query map[string]string, line int) (string, *ParseError) {
	if aliasAsOf := strings.TrimSpace(query["as_of"]); aliasAsOf != "" {
		if asOf != "" && asOf != aliasAsOf {
			return "", parseErr(line, fmt.Sprintf("source %q has conflicting as_of values (%q vs %q)", ref, asOf, aliasAsOf))
		}
		asOf = aliasAsOf
	}
	if scope == "external" {
		if !validDate(asOf) {
			return "", parseErr(line, fmt.Sprintf("external source %q must include as_of date YYYY-MM-DD", ref))
		}
	} else if asOf != "" {
		return "", parseErr(line, fmt.Sprintf("source %q includes as_of date but scope is %q", ref, scope))
	}
	return asOf, nil
}

// resolveProvider validates provider_id and upstream_system against the scope.
func resolveProvider(ref, scope string, query map[string]string, line int) (string, string, *ParseError) {
	providerID := strings.TrimSpace(query["provider_id"])
	upstreamSystem := strings.TrimSpace(query["upstream_system"])
	if scope == "external" {
		if providerID == "" {
			return "", "", parseErr(line, fmt.Sprintf("external source %q must include provider_id", ref))
		}
		if !providerIDRe.MatchString(providerID) {
			return "", "", parseErr(line, fmt.Sprintf("external source %q has invalid provider_id", ref))
		}
	} else if providerID != "" {
		return "", "", parseErr(line, fmt.Sprintf("source %q includes provider_id but scope is %q", ref, scope))
	}
	if upstreamSystem != "" && !systemIDRe.MatchString(upstreamSystem) {
		return "", "", parseErr(line, fmt.Sprintf("source %q has invalid upstream_system", ref))
	}
	return providerID, upstreamSystem, nil
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

func normalizeAnnotationAliases(fields map[string]string) ([]string, error) {
	aliases := map[string][]string{
		"annotation_schema_version":    {"schema_version"},
		"field":                        {"field_path", "json_path", "property_path"},
		"source_system":                {"service_name"},
		"source_version":               {"service_version", "spec_version"},
		"min_supported_source_version": {"min_source_version", "min_supported_version"},
		"owner":                        {"owner_team"},
		"contract_test_id":             {"contract_test", "test_id"},
	}
	return normalizeAliases(fields, aliases)
}

func normalizeSourceQueryAliases(query map[string]string) ([]string, error) {
	aliases := map[string][]string{
		"contract_ref":    {"schema_ref", "spec_ref", "contract_uri", "schema_url"},
		"provider_id":     {"provider", "external_provider"},
		"upstream_system": {"upstream_service", "upstream_source_system"},
		"as_of":           {"asof", "snapshot_as_of"},
	}
	return normalizeAliases(query, aliases)
}

func normalizeAliases(fields map[string]string, aliases map[string][]string) ([]string, error) {
	mappedFrom := make([]string, 0)
	for canonical, aliasKeys := range aliases {
		canonicalValue := strings.TrimSpace(fields[canonical])
		for _, aliasKey := range aliasKeys {
			aliasValue := strings.TrimSpace(fields[aliasKey])
			if aliasValue == "" {
				continue
			}
			if canonicalValue == "" {
				fields[canonical] = aliasValue
				canonicalValue = aliasValue
				mappedFrom = append(mappedFrom, aliasKey)
				continue
			}
			if canonicalValue != aliasValue {
				return nil, fmt.Errorf("conflicting values for %q and alias %q", canonical, aliasKey)
			}
		}
	}
	return uniqueSortedStrings(mappedFrom), nil
}
