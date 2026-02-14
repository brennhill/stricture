// artifact_test.go - Tests for lineage artifact collection.
package lineage

import (
	"os"
	"path/filepath"
	"testing"
)

func validLine(fieldID string, fieldPath string) string {
	return "// stricture-source annotation_schema_version=1 field_id=" + fieldID + " field=" + fieldPath + " source_system=Identity source_version=v2026.02 min_supported_source_version=v2026.01 transform_type=normalize merge_strategy=single_source break_policy=additive_only confidence=declared data_classification=internal owner=team.identity escalation=slack:#identity-oncall contract_test_id=ci://contracts/identity-user-id introduced_at=2026-01-10 sources=api:identity.GetUser#response.id@cross_repo?contract_ref=git+https://github.com/acme/identity//openapi.yaml@a1b2 flow=\"from @Identity normalized @self\" note=\"normalized by UserNormalizer.Apply\""
}

func compactLine(fieldPath string) string {
	return "// stricture-source field=" + fieldPath + " source_system=Identity source_version=v2026.02 sources=api:identity.GetUser#response.id@cross_repo?contract_ref=git+https://github.com/acme/identity//openapi.yaml@a1b2"
}

func TestCollect_BuildsDeterministicArtifact(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "b.go"), []byte(validLine("response_user_name", "response.user_name")+"\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "a.go"), []byte(validLine("response_user_id", "response.user_id")+"\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	artifact, parseErrs, err := Collect([]string{tmp})
	if err != nil {
		t.Fatalf("collect error: %v", err)
	}
	if len(parseErrs) != 0 {
		t.Fatalf("parseErrs len = %d, want 0", len(parseErrs))
	}
	if artifact.SchemaVersion != "1" {
		t.Fatalf("schema_version = %q, want 1", artifact.SchemaVersion)
	}
	if len(artifact.Fields) != 2 {
		t.Fatalf("fields len = %d, want 2", len(artifact.Fields))
	}
	if artifact.Fields[0].FieldID != "response_user_id" {
		t.Fatalf("fields[0].field_id = %q, want response_user_id", artifact.Fields[0].FieldID)
	}
	if artifact.Fields[1].FieldID != "response_user_name" {
		t.Fatalf("fields[1].field_id = %q, want response_user_name", artifact.Fields[1].FieldID)
	}
}

func TestCollect_ReturnsParseErrorsWithFilePaths(t *testing.T) {
	tmp := t.TempDir()
	valid := filepath.Join(tmp, "valid.go")
	invalid := filepath.Join(tmp, "invalid.go")

	if err := os.WriteFile(valid, []byte(validLine("response_user_id", "response.user_id")+"\n"), 0o644); err != nil {
		t.Fatalf("write valid: %v", err)
	}
	if err := os.WriteFile(invalid, []byte("// stricture-source field=response.user_id\n"), 0o644); err != nil {
		t.Fatalf("write invalid: %v", err)
	}

	artifact, parseErrs, err := Collect([]string{tmp})
	if err != nil {
		t.Fatalf("collect error: %v", err)
	}
	if len(artifact.Fields) != 1 {
		t.Fatalf("fields len = %d, want 1", len(artifact.Fields))
	}
	if len(parseErrs) != 1 {
		t.Fatalf("parseErrs len = %d, want 1", len(parseErrs))
	}
	if parseErrs[0].FilePath == "" {
		t.Fatalf("expected parse error to include file path")
	}
}

func TestCollect_IncludesOverrides(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "override.go")
	content := `// stricture-lineage-override field_id=response_user_id change_type=field_removed expires=2099-12-31 reason="temporary migration window" ticket=INC-123` + "\n"

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	artifact, parseErrs, err := Collect([]string{tmp})
	if err != nil {
		t.Fatalf("collect error: %v", err)
	}
	if len(parseErrs) != 0 {
		t.Fatalf("parseErrs len = %d, want 0", len(parseErrs))
	}
	if len(artifact.Overrides) != 1 {
		t.Fatalf("overrides len = %d, want 1", len(artifact.Overrides))
	}
	if artifact.Overrides[0].FieldID != "response_user_id" {
		t.Fatalf("field_id = %q, want response_user_id", artifact.Overrides[0].FieldID)
	}
	if artifact.Overrides[0].FilePath == "" {
		t.Fatalf("expected override to include file_path")
	}
	if artifact.Overrides[0].Line != 1 {
		t.Fatalf("line = %d, want 1", artifact.Overrides[0].Line)
	}
}

func TestCollect_AppliesDefaultsForCompactAnnotation(t *testing.T) {
	tmp := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmp, "compact.go"), []byte(compactLine("response.user_id")+"\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	artifact, parseErrs, err := Collect([]string{tmp})
	if err != nil {
		t.Fatalf("collect error: %v", err)
	}
	if len(parseErrs) != 0 {
		t.Fatalf("parseErrs len = %d, want 0", len(parseErrs))
	}
	if len(artifact.Fields) != 1 {
		t.Fatalf("fields len = %d, want 1", len(artifact.Fields))
	}
	field := artifact.Fields[0]
	if field.FieldID != "response_user_id" {
		t.Fatalf("field_id = %q, want response_user_id", field.FieldID)
	}
	if field.Owner != "team.identity" {
		t.Fatalf("owner = %q, want team.identity", field.Owner)
	}
	if field.ContractTestID != "ci://contracts/identity/response_user_id" {
		t.Fatalf("contract_test_id = %q, want ci://contracts/identity/response_user_id", field.ContractTestID)
	}
	if field.Note != "defaulted_by=stricture" {
		t.Fatalf("note = %q, want defaulted_by=stricture", field.Note)
	}
}
