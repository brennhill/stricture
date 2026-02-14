// lineage_export_profile_test.go — Integration checks for lineage-export profiles.
//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLineageExportProfileOTelIncludesAliasFields(t *testing.T) {
	tmp := t.TempDir()
	sourcePath := filepath.Join(tmp, "lineage.go")
	content := `// stricture-source annotation_schema_version=1 field_id=response_user_id field=response.user_id source_system=Identity source_version=v1 min_supported_source_version=v1 transform_type=normalize merge_strategy=single_source break_policy=additive_only confidence=declared data_classification=internal owner=team.identity escalation=slack:#identity-oncall contract_test_id=ci://contracts/identity-user-id introduced_at=2026-01-10 sources=api:identity.GetUser#response.id@cross_repo?contract_ref=git+https://github.com/acme/identity//openapi.yaml@a1b2 flow="from @Identity normalized @self" note="normalized by UserNormalizer.Apply"` + "\n"
	if err := os.WriteFile(sourcePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	stdout, stderr, code := runInDir(t, tmp, "lineage-export", "--profile", "otel", ".")
	if code != 0 {
		t.Fatalf("lineage-export otel exit code = %d, want 0\nstderr=%q\nstdout=%q", code, stderr, stdout)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\noutput=%q", err, stdout)
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
}

func TestLineageExportRejectsInvalidProfile(t *testing.T) {
	tmp := t.TempDir()
	sourcePath := filepath.Join(tmp, "lineage.go")
	content := `// stricture-source annotation_schema_version=1 field_id=response_user_id field=response.user_id source_system=Identity source_version=v1 min_supported_source_version=v1 transform_type=normalize merge_strategy=single_source break_policy=additive_only confidence=declared data_classification=internal owner=team.identity escalation=slack:#identity-oncall contract_test_id=ci://contracts/identity-user-id introduced_at=2026-01-10 sources=api:identity.GetUser#response.id@cross_repo?contract_ref=git+https://github.com/acme/identity//openapi.yaml@a1b2 flow="from @Identity normalized @self" note="normalized by UserNormalizer.Apply"` + "\n"
	if err := os.WriteFile(sourcePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	_, stderr, code := runInDir(t, tmp, "lineage-export", "--profile", "bogus", ".")
	if code != 2 {
		t.Fatalf("lineage-export invalid profile exit code = %d, want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "invalid profile") {
		t.Fatalf("stderr should mention invalid profile, got %q", stderr)
	}
}
