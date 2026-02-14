// export_profile.go - Profile-aware lineage artifact encoding.
package lineage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExportProfile controls alias/proxy fields in lineage artifact output.
type ExportProfile string

const (
	ProfileStricture   ExportProfile = "stricture"
	ProfileOTel        ExportProfile = "otel"
	ProfileOpenLineage ExportProfile = "openlineage"
	ProfileOpenAPI     ExportProfile = "openapi"
	ProfileAsyncAPI    ExportProfile = "asyncapi"
)

// ParseExportProfile parses CLI profile names and aliases.
func ParseExportProfile(raw string) (ExportProfile, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "", string(ProfileStricture):
		return ProfileStricture, nil
	case string(ProfileOTel), "opentelemetry":
		return ProfileOTel, nil
	case string(ProfileOpenLineage):
		return ProfileOpenLineage, nil
	case string(ProfileOpenAPI):
		return ProfileOpenAPI, nil
	case string(ProfileAsyncAPI):
		return ProfileAsyncAPI, nil
	default:
		return "", fmt.Errorf("invalid profile %q (valid: stricture, openlineage, otel, openapi, asyncapi)", raw)
	}
}

// MarshalArtifactForProfile marshals a lineage artifact with optional profile aliases.
func MarshalArtifactForProfile(artifact Artifact, profile ExportProfile) ([]byte, error) {
	if artifact.SchemaVersion == "" {
		artifact.SchemaVersion = "1"
	}
	if profile == ProfileStricture {
		return json.MarshalIndent(artifact, "", "  ")
	}

	canonical, err := marshalArtifactMap(artifact)
	if err != nil {
		return nil, err
	}
	canonical["export_profile"] = string(profile)

	fieldsAny, ok := canonical["fields"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid artifact payload: fields must be array")
	}
	for idx, fieldAny := range fieldsAny {
		fieldMap, ok := fieldAny.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid artifact payload: field entry %d must be object", idx)
		}
		if idx >= len(artifact.Fields) {
			break
		}
		annotation := artifact.Fields[idx]
		applyAnnotationProfileAliases(fieldMap, annotation, profile)
	}
	return json.MarshalIndent(canonical, "", "  ")
}

// WriteArtifactForProfile writes profile-aware artifact JSON with deterministic formatting.
func WriteArtifactForProfile(path string, artifact Artifact, profile ExportProfile) error {
	data, err := MarshalArtifactForProfile(artifact, profile)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir artifact dir: %w", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write lineage artifact: %w", err)
	}
	return nil
}

func marshalArtifactMap(artifact Artifact) (map[string]interface{}, error) {
	encoded, err := json.Marshal(artifact)
	if err != nil {
		return nil, fmt.Errorf("marshal lineage artifact: %w", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(encoded, &payload); err != nil {
		return nil, fmt.Errorf("decode lineage artifact: %w", err)
	}
	return payload, nil
}

func applyAnnotationProfileAliases(field map[string]interface{}, annotation Annotation, profile ExportProfile) {
	field["field_path"] = annotation.Field
	field["service_name"] = annotation.SourceSystem
	field["service_version"] = annotation.SourceVersion
	field["spec_version"] = annotation.SourceVersion
	field["owner_team"] = annotation.Owner

	switch profile {
	case ProfileOTel:
		field["service.name"] = annotation.SourceSystem
		field["service.version"] = annotation.SourceVersion
		field["owner.team"] = annotation.Owner
	case ProfileOpenLineage:
		field["openlineage_job_name"] = annotation.SourceSystem
		field["openlineage.job.name"] = annotation.SourceSystem
		field["openlineage_job_version"] = annotation.SourceVersion
		field["openlineage.job.version"] = annotation.SourceVersion
	case ProfileOpenAPI:
		field["openapi_field_path"] = annotation.Field
	case ProfileAsyncAPI:
		field["asyncapi_field_path"] = annotation.Field
	}

	sourcesAny, ok := field["sources"].([]interface{})
	if !ok {
		return
	}
	for idx, sourceAny := range sourcesAny {
		sourceMap, ok := sourceAny.(map[string]interface{})
		if !ok || idx >= len(annotation.Sources) {
			continue
		}
		source := annotation.Sources[idx]
		sourceMap["schema_ref"] = source.ContractRef
		sourceMap["spec_ref"] = source.ContractRef
		sourceMap["contract_uri"] = source.ContractRef
		sourceMap["schema_url"] = source.ContractRef
		if source.ProviderID != "" {
			sourceMap["provider"] = source.ProviderID
		}
		if source.UpstreamSystem != "" {
			sourceMap["upstream_service"] = source.UpstreamSystem
		}
		if source.AsOf != "" {
			sourceMap["asof"] = source.AsOf
			sourceMap["snapshot_as_of"] = source.AsOf
		}
	}
}
