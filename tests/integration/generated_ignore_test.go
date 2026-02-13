// generated_ignore_test.go — Integration checks for default generated-file ignores.
//go:build integration

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGeneratedFilesAreIgnoredByDefault(t *testing.T) {
	tmp := t.TempDir()

	regular := filepath.Join(tmp, "service.ts")
	if err := os.WriteFile(regular, []byte("export const value = 1;\n"), 0o644); err != nil {
		t.Fatalf("write regular file: %v", err)
	}
	generatedTS := filepath.Join(tmp, "client.generated.ts")
	if err := os.WriteFile(generatedTS, []byte("export const generated = true;\n"), 0o644); err != nil {
		t.Fatalf("write generated ts: %v", err)
	}
	protoGo := filepath.Join(tmp, "types.pb.go")
	if err := os.WriteFile(protoGo, []byte("package tmp\n"), 0o644); err != nil {
		t.Fatalf("write pb go: %v", err)
	}
	protoTS := filepath.Join(tmp, "types.pb.ts")
	if err := os.WriteFile(protoTS, []byte("export type T = { id: string };\n"), 0o644); err != nil {
		t.Fatalf("write pb ts: %v", err)
	}

	stdout, stderr, code := runInDir(t, tmp, "--format", "json", "--rule", "CONV-file-header", ".")
	if code != 1 {
		t.Fatalf("lint should fail only for regular source file; code=%d stderr=%q stdout=%q", code, stderr, stdout)
	}

	var result struct {
		Violations []struct {
			FilePath string `json:"filePath"`
		} `json:"violations"`
		Summary struct {
			TotalViolations float64 `json:"totalViolations"`
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("unmarshal JSON output: %v\noutput=%q", err, stdout)
	}
	if int(result.Summary.TotalViolations) != 1 || len(result.Violations) != 1 {
		t.Fatalf("expected exactly one violation after generated-file filtering, got total=%v len=%d", result.Summary.TotalViolations, len(result.Violations))
	}
	if filepath.Base(result.Violations[0].FilePath) != "service.ts" {
		t.Fatalf("expected violation on service.ts only, got %q", result.Violations[0].FilePath)
	}
}
