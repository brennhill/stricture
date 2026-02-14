// alias_unmarshal.go - JSON alias/synonym support for lineage models.
package lineage

import (
	"encoding/json"
	"fmt"
	"strings"
)

// UnmarshalJSON supports canonical keys plus profile aliases.
func (a *Annotation) UnmarshalJSON(data []byte) error {
	type annotationAlias Annotation
	var base annotationAlias
	if err := json.Unmarshal(data, &base); err != nil {
		return err
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	mappedFrom := make([]string, 0)
	var err error
	var mapped []string

	base.AnnotationSchemaVersion, mapped, err = resolveStringAlias(raw, base.AnnotationSchemaVersion, "annotation_schema_version", "schema_version")
	if err != nil {
		return err
	}
	mappedFrom = append(mappedFrom, mapped...)

	base.Field, mapped, err = resolveStringAlias(raw, base.Field, "field", "field_path", "json_path", "property_path", "openapi_field_path", "asyncapi_field_path")
	if err != nil {
		return err
	}
	mappedFrom = append(mappedFrom, mapped...)

	base.SourceSystem, mapped, err = resolveStringAlias(raw, base.SourceSystem, "source_system", "service_name", "service.name", "openlineage_job_name", "openlineage.job.name")
	if err != nil {
		return err
	}
	mappedFrom = append(mappedFrom, mapped...)

	base.SourceVersion, mapped, err = resolveStringAlias(raw, base.SourceVersion, "source_version", "service_version", "service.version", "spec_version", "openlineage_job_version", "openlineage.job.version")
	if err != nil {
		return err
	}
	mappedFrom = append(mappedFrom, mapped...)

	base.MinSupportedSourceVersion, mapped, err = resolveStringAlias(raw, base.MinSupportedSourceVersion, "min_supported_source_version", "min_source_version", "min_supported_version")
	if err != nil {
		return err
	}
	mappedFrom = append(mappedFrom, mapped...)

	base.Owner, mapped, err = resolveStringAlias(raw, base.Owner, "owner", "owner_team", "owner.team")
	if err != nil {
		return err
	}
	mappedFrom = append(mappedFrom, mapped...)

	base.ContractTestID, mapped, err = resolveStringAlias(raw, base.ContractTestID, "contract_test_id", "contract_test", "test_id")
	if err != nil {
		return err
	}
	mappedFrom = append(mappedFrom, mapped...)

	if len(base.Sources) == 0 {
		if rawRefs, ok := raw["source_refs"]; ok {
			var refs []SourceRef
			if err := json.Unmarshal(rawRefs, &refs); err != nil {
				return fmt.Errorf("parse source_refs: %w", err)
			}
			base.Sources = refs
			mappedFrom = append(mappedFrom, "source_refs")
		}
	}

	for _, src := range base.Sources {
		mappedFrom = append(mappedFrom, src.MappedFrom...)
	}
	base.MappedFrom = uniqueSortedStrings(append(base.MappedFrom, mappedFrom...))

	*a = Annotation(base)
	return nil
}

// UnmarshalJSON supports canonical keys plus profile aliases.
func (s *SourceRef) UnmarshalJSON(data []byte) error {
	type sourceAlias SourceRef
	var base sourceAlias
	if err := json.Unmarshal(data, &base); err != nil {
		return err
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	mappedFrom := make([]string, 0)
	var err error
	var mapped []string

	base.ContractRef, mapped, err = resolveStringAlias(raw, base.ContractRef, "contract_ref", "schema_ref", "spec_ref", "contract_uri", "schema_url")
	if err != nil {
		return err
	}
	mappedFrom = append(mappedFrom, mapped...)

	base.ProviderID, mapped, err = resolveStringAlias(raw, base.ProviderID, "provider_id", "provider", "external_provider")
	if err != nil {
		return err
	}
	mappedFrom = append(mappedFrom, mapped...)

	base.UpstreamSystem, mapped, err = resolveStringAlias(raw, base.UpstreamSystem, "upstream_system", "upstream_service", "upstream_source_system")
	if err != nil {
		return err
	}
	mappedFrom = append(mappedFrom, mapped...)

	base.AsOf, mapped, err = resolveStringAlias(raw, base.AsOf, "as_of", "asof", "snapshot_as_of")
	if err != nil {
		return err
	}
	mappedFrom = append(mappedFrom, mapped...)

	base.MappedFrom = uniqueSortedStrings(append(base.MappedFrom, mappedFrom...))
	*s = SourceRef(base)
	return nil
}

func resolveStringAlias(raw map[string]json.RawMessage, canonicalValue string, canonicalKey string, aliases ...string) (string, []string, error) {
	trimmedCanonical := strings.TrimSpace(canonicalValue)

	aliasMatches := make([]string, 0)
	aliasValue := ""
	for _, alias := range aliases {
		rawValue, ok := raw[alias]
		if !ok {
			continue
		}
		value, err := decodeStringRaw(rawValue)
		if err != nil {
			return "", nil, fmt.Errorf("parse %q: %w", alias, err)
		}
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		aliasMatches = append(aliasMatches, alias)
		if aliasValue == "" {
			aliasValue = value
			continue
		}
		if aliasValue != value {
			return "", nil, fmt.Errorf("conflicting alias values for %q", canonicalKey)
		}
	}

	if trimmedCanonical != "" {
		if aliasValue != "" && aliasValue != trimmedCanonical {
			return "", nil, fmt.Errorf("conflicting values for %q and aliases", canonicalKey)
		}
		return trimmedCanonical, uniqueSortedStrings(aliasMatches), nil
	}
	if aliasValue != "" {
		return aliasValue, uniqueSortedStrings(aliasMatches), nil
	}
	return trimmedCanonical, nil, nil
}

func decodeStringRaw(raw json.RawMessage) (string, error) {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", err
	}
	return value, nil
}
