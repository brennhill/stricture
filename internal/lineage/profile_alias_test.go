// profile_alias_test.go - Tests for profile aliases and profile export encoding.
package lineage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseExportProfile(t *testing.T) {
	cases := []struct {
		raw     string
		want    ExportProfile
		wantErr bool
	}{
		{raw: "", want: ProfileStricture},
		{raw: "stricture", want: ProfileStricture},
		{raw: "otel", want: ProfileOTel},
		{raw: "OpenTelemetry", want: ProfileOTel},
		{raw: "openlineage", want: ProfileOpenLineage},
		{raw: "openapi", want: ProfileOpenAPI},
		{raw: "asyncapi", want: ProfileAsyncAPI},
		{raw: "bad-profile", wantErr: true},
	}

	for _, tc := range cases {
		got, err := ParseExportProfile(tc.raw)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("ParseExportProfile(%q) expected error", tc.raw)
			}
			continue
		}
		if err != nil {
			t.Fatalf("ParseExportProfile(%q) unexpected error: %v", tc.raw, err)
		}
		if got != tc.want {
			t.Fatalf("ParseExportProfile(%q) = %q, want %q", tc.raw, got, tc.want)
		}
	}
}

func TestLoadArtifactAcceptsAliasJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "artifact.json")
	content := `{
  "schema_version": "1",
  "fields": [
    {
      "annotation_schema_version": "1",
      "field_id": "response_user_id",
      "field_path": "response.user_id",
      "service_name": "Identity",
      "service_version": "v1",
      "min_source_version": "v1",
      "owner_team": "team.identity",
      "contract_test": "ci://contracts/identity-user-id",
      "sources": [
        {
          "kind": "api",
          "target": "identity.GetUser",
          "path": "response.id",
          "scope": "cross_repo",
          "schema_ref": "https://specs.example.com/id",
          "raw": "api:identity.GetUser#response.id@cross_repo?schema_ref=https://specs.example.com/id"
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	artifact, err := LoadArtifact(path)
	if err != nil {
		t.Fatalf("LoadArtifact unexpected error: %v", err)
	}
	if len(artifact.Fields) != 1 {
		t.Fatalf("fields len = %d, want 1", len(artifact.Fields))
	}
	field := artifact.Fields[0]
	if field.Field != "response.user_id" {
		t.Fatalf("field = %q, want response.user_id", field.Field)
	}
	if field.SourceSystem != "Identity" {
		t.Fatalf("source_system = %q, want Identity", field.SourceSystem)
	}
	if field.SourceVersion != "v1" {
		t.Fatalf("source_version = %q, want v1", field.SourceVersion)
	}
	if field.MinSupportedSourceVersion != "v1" {
		t.Fatalf("min_supported_source_version = %q, want v1", field.MinSupportedSourceVersion)
	}
	if field.Owner != "team.identity" {
		t.Fatalf("owner = %q, want team.identity", field.Owner)
	}
	if field.ContractTestID != "ci://contracts/identity-user-id" {
		t.Fatalf("contract_test_id = %q, unexpected", field.ContractTestID)
	}
	if len(field.Sources) != 1 {
		t.Fatalf("sources len = %d, want 1", len(field.Sources))
	}
	if field.Sources[0].ContractRef != "https://specs.example.com/id" {
		t.Fatalf("contract_ref = %q, unexpected", field.Sources[0].ContractRef)
	}
}

func TestLoadArtifactRejectsConflictingAliasJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "artifact.json")
	content := `{
  "schema_version": "1",
  "fields": [
    {
      "annotation_schema_version": "1",
      "field_id": "response_user_id",
      "field": "response.user_id",
      "source_system": "Identity",
      "service_name": "Billing"
    }
  ]
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	_, err := LoadArtifact(path)
	if err == nil {
		t.Fatalf("expected conflicting alias error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "conflicting") {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

func TestMarshalArtifactForProfileAddsAliases(t *testing.T) {
	artifact := Artifact{
		SchemaVersion: "1",
		Fields: []Annotation{
			{
				AnnotationSchemaVersion: "1",
				FieldID:                 "response_user_id",
				Field:                   "response.user_id",
				SourceSystem:            "Identity",
				SourceVersion:           "v1",
				Owner:                   "team.identity",
				Sources: []SourceRef{
					{
						Kind:        "api",
						Target:      "identity.GetUser",
						Path:        "response.id",
						Scope:       "cross_repo",
						ContractRef: "https://specs.example.com/id",
						Raw:         "api:identity.GetUser#response.id@cross_repo?contract_ref=https://specs.example.com/id",
					},
				},
			},
		},
	}

	out, err := MarshalArtifactForProfile(artifact, ProfileOTel)
	if err != nil {
		t.Fatalf("MarshalArtifactForProfile(otel) error: %v", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(out, &payload); err != nil {
		t.Fatalf("unmarshal otel payload: %v", err)
	}
	if payload["export_profile"] != "otel" {
		t.Fatalf("export_profile = %v, want otel", payload["export_profile"])
	}

	fields, ok := payload["fields"].([]interface{})
	if !ok || len(fields) != 1 {
		t.Fatalf("fields payload invalid: %#v", payload["fields"])
	}
	field, ok := fields[0].(map[string]interface{})
	if !ok {
		t.Fatalf("field payload invalid: %#v", fields[0])
	}
	if field["service_name"] != "Identity" {
		t.Fatalf("service_name = %v, want Identity", field["service_name"])
	}
	if field["service.name"] != "Identity" {
		t.Fatalf("service.name = %v, want Identity", field["service.name"])
	}
	if field["field_path"] != "response.user_id" {
		t.Fatalf("field_path = %v, want response.user_id", field["field_path"])
	}
}

func TestMarshalArtifactForProfileStrictureIsCanonical(t *testing.T) {
	artifact := Artifact{
		SchemaVersion: "1",
		Fields: []Annotation{
			{
				AnnotationSchemaVersion: "1",
				FieldID:                 "response_user_id",
				Field:                   "response.user_id",
				SourceSystem:            "Identity",
			},
		},
	}

	out, err := MarshalArtifactForProfile(artifact, ProfileStricture)
	if err != nil {
		t.Fatalf("MarshalArtifactForProfile(stricture) error: %v", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(out, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if _, ok := payload["export_profile"]; ok {
		t.Fatalf("stricture profile should not include export_profile")
	}

	fields, ok := payload["fields"].([]interface{})
	if !ok || len(fields) != 1 {
		t.Fatalf("fields payload invalid: %#v", payload["fields"])
	}
	field := fields[0].(map[string]interface{})
	if _, ok := field["field_path"]; ok {
		t.Fatalf("stricture profile should not include field_path alias")
	}
}

func TestMarshalArtifactForProfileOpenAPIAndAsyncAPI(t *testing.T) {
	artifact := Artifact{
		SchemaVersion: "1",
		Fields: []Annotation{
			{
				AnnotationSchemaVersion: "1",
				FieldID:                 "response_user_id",
				Field:                   "response.user_id",
				SourceSystem:            "Identity",
				SourceVersion:           "v1",
				Owner:                   "team.identity",
			},
		},
	}

	for _, profile := range []ExportProfile{ProfileOpenAPI, ProfileAsyncAPI} {
		out, err := MarshalArtifactForProfile(artifact, profile)
		if err != nil {
			t.Fatalf("MarshalArtifactForProfile(%s) error: %v", profile, err)
		}
		var payload map[string]interface{}
		if err := json.Unmarshal(out, &payload); err != nil {
			t.Fatalf("unmarshal payload: %v", err)
		}
		if payload["export_profile"] != string(profile) {
			t.Fatalf("export_profile = %v, want %s", payload["export_profile"], profile)
		}
		fields := payload["fields"].([]interface{})
		field := fields[0].(map[string]interface{})
		if field["field_path"] != "response.user_id" {
			t.Fatalf("field_path = %v, want response.user_id", field["field_path"])
		}
	}
}

func TestWriteArtifactForProfile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "lineage", "artifact.json")
	artifact := Artifact{
		SchemaVersion: "1",
		Fields: []Annotation{
			{
				AnnotationSchemaVersion: "1",
				FieldID:                 "response_user_id",
				Field:                   "response.user_id",
				SourceSystem:            "Identity",
				SourceVersion:           "v1",
				Owner:                   "team.identity",
				Sources: []SourceRef{
					{
						Kind:        "api",
						Target:      "identity.GetUser",
						Path:        "response.id",
						Scope:       "cross_repo",
						ContractRef: "https://specs.example.com/id",
						Raw:         "api:identity.GetUser#response.id@cross_repo?contract_ref=https://specs.example.com/id",
					},
				},
			},
		},
	}

	if err := WriteArtifactForProfile(path, artifact, ProfileOpenLineage); err != nil {
		t.Fatalf("WriteArtifactForProfile error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if !strings.HasSuffix(string(data), "\n") {
		t.Fatalf("expected trailing newline")
	}
	if !strings.Contains(string(data), "\"export_profile\": \"openlineage\"") {
		t.Fatalf("expected openlineage profile marker in output")
	}
}

func TestWriteArtifactForProfileMkdirError(t *testing.T) {
	tmp := t.TempDir()
	parent := filepath.Join(tmp, "not-a-dir")
	if err := os.WriteFile(parent, []byte("x"), 0o644); err != nil {
		t.Fatalf("write parent file: %v", err)
	}
	path := filepath.Join(parent, "artifact.json")
	err := WriteArtifactForProfile(path, Artifact{SchemaVersion: "1"}, ProfileStricture)
	if err == nil {
		t.Fatalf("expected mkdir failure")
	}
}

func TestSourceRefUnmarshalAliasAndConflicts(t *testing.T) {
	var ref SourceRef
	if err := json.Unmarshal([]byte(`{"kind":"api","target":"identity.GetUser","path":"response.id","scope":"cross_repo","schema_ref":"https://specs.example.com/id","raw":"x"}`), &ref); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if ref.ContractRef != "https://specs.example.com/id" {
		t.Fatalf("contract_ref = %q, unexpected", ref.ContractRef)
	}

	if err := json.Unmarshal([]byte(`{"contract_ref":"https://a.example","schema_ref":"https://b.example"}`), &ref); err == nil {
		t.Fatalf("expected conflict error")
	}

	if err := json.Unmarshal([]byte(`{"schema_ref":123}`), &ref); err == nil {
		t.Fatalf("expected non-string alias parse error")
	}
}

func TestAnnotationUnmarshalSourceRefsAlias(t *testing.T) {
	var ann Annotation
	payload := `{
  "annotation_schema_version":"1",
  "field_id":"response_user_id",
  "field":"response.user_id",
  "source_system":"Identity",
  "source_version":"v1",
  "min_supported_source_version":"v1",
  "transform_type":"normalize",
  "merge_strategy":"single_source",
  "break_policy":"additive_only",
  "confidence":"declared",
  "data_classification":"internal",
  "owner":"team.identity",
  "escalation":"slack:#identity-oncall",
  "contract_test_id":"ci://contracts/identity-user-id",
  "introduced_at":"2026-01-10",
  "flow":"from @Identity normalized @self",
  "note":"n",
  "source_refs":[{"kind":"api","target":"identity.GetUser","path":"response.id","scope":"cross_repo","schema_ref":"https://specs.example.com/id","raw":"x"}]
}`
	if err := json.Unmarshal([]byte(payload), &ann); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if len(ann.Sources) != 1 {
		t.Fatalf("sources len = %d, want 1", len(ann.Sources))
	}
	if ann.Sources[0].ContractRef != "https://specs.example.com/id" {
		t.Fatalf("contract_ref = %q, unexpected", ann.Sources[0].ContractRef)
	}
}

func TestProfileRoundTripPreservesCanonicalCoreFields(t *testing.T) {
	artifact := Artifact{
		SchemaVersion: "1",
		Fields: []Annotation{
			{
				AnnotationSchemaVersion:   "1",
				FieldID:                   "response_user_id",
				Field:                     "response.user_id",
				SourceSystem:              "Identity",
				SourceVersion:             "v1",
				MinSupportedSourceVersion: "v1",
				TransformType:             "normalize",
				MergeStrategy:             "single_source",
				BreakPolicy:               "additive_only",
				Confidence:                "declared",
				DataClassification:        "internal",
				Owner:                     "team.identity",
				Escalation:                "slack:#identity-oncall",
				ContractTestID:            "ci://contracts/identity-user-id",
				IntroducedAt:              "2026-01-10",
				Flow:                      "from @Identity normalized @self",
				Note:                      "normalized by UserNormalizer.Apply",
				Sources: []SourceRef{
					{
						Kind:        "api",
						Target:      "identity.GetUser",
						Path:        "response.id",
						Scope:       "cross_repo",
						ContractRef: "https://specs.example.com/id",
						Raw:         "api:identity.GetUser#response.id@cross_repo?contract_ref=https://specs.example.com/id",
					},
				},
			},
		},
	}

	profiles := []ExportProfile{
		ProfileOTel,
		ProfileOpenLineage,
		ProfileOpenAPI,
		ProfileAsyncAPI,
	}
	for _, profile := range profiles {
		out, err := MarshalArtifactForProfile(artifact, profile)
		if err != nil {
			t.Fatalf("MarshalArtifactForProfile(%s): %v", profile, err)
		}

		tmp := t.TempDir()
		path := filepath.Join(tmp, "artifact.json")
		if err := os.WriteFile(path, out, 0o644); err != nil {
			t.Fatalf("write profile artifact: %v", err)
		}

		loaded, err := LoadArtifact(path)
		if err != nil {
			t.Fatalf("LoadArtifact(%s): %v", profile, err)
		}
		if len(loaded.Fields) != 1 {
			t.Fatalf("profile=%s fields len = %d, want 1", profile, len(loaded.Fields))
		}
		got := loaded.Fields[0]
		if got.FieldID != artifact.Fields[0].FieldID ||
			got.Field != artifact.Fields[0].Field ||
			got.SourceSystem != artifact.Fields[0].SourceSystem ||
			got.SourceVersion != artifact.Fields[0].SourceVersion ||
			got.Owner != artifact.Fields[0].Owner {
			t.Fatalf("profile=%s round-trip mismatch: got=%+v", profile, got)
		}
		if len(got.Sources) != 1 || got.Sources[0].ContractRef != artifact.Fields[0].Sources[0].ContractRef {
			t.Fatalf("profile=%s source round-trip mismatch: %+v", profile, got.Sources)
		}
	}
}
