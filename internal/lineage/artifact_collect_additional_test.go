package lineage

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCollect_DefaultsToCurrentDirectoryWhenPathsEmpty(t *testing.T) {
	tmp := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("restore cwd: %v", chdirErr)
		}
	})

	if err := os.WriteFile(filepath.Join(tmp, "service.go"), []byte(validLine("response_user_id", "response.user_id")+"\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	artifact, parseErrs, err := Collect(nil)
	if err != nil {
		t.Fatalf("collect error: %v", err)
	}
	if len(parseErrs) != 0 {
		t.Fatalf("parseErrs len = %d, want 0", len(parseErrs))
	}
	if len(artifact.Fields) != 1 {
		t.Fatalf("fields len = %d, want 1", len(artifact.Fields))
	}
	if artifact.Fields[0].FieldID != "response_user_id" {
		t.Fatalf("field_id = %q, want response_user_id", artifact.Fields[0].FieldID)
	}
}

func TestCollect_SkipsIgnoredDirectoriesAndNonSourceFiles(t *testing.T) {
	tmp := t.TempDir()

	ignoredDirs := []string{
		filepath.Join(tmp, ".git"),
		filepath.Join(tmp, "node_modules"),
		filepath.Join(tmp, "bin"),
	}
	for _, dir := range ignoredDirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "ignored.go"), []byte("// strict-source field=response.user_id\n"), 0o644); err != nil {
			t.Fatalf("write ignored file: %v", err)
		}
	}

	if err := os.WriteFile(filepath.Join(tmp, "README.md"), []byte(validLine("response_from_readme", "response.readme")+"\n"), 0o644); err != nil {
		t.Fatalf("write markdown file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "service.go"), []byte(validLine("response_user_id", "response.user_id")+"\n"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
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
	if artifact.Fields[0].FieldID != "response_user_id" {
		t.Fatalf("field_id = %q, want response_user_id", artifact.Fields[0].FieldID)
	}
}

func TestCollect_SkipsStandaloneNonSourceFiles(t *testing.T) {
	tmp := t.TempDir()
	nonSource := filepath.Join(tmp, "notes.txt")
	if err := os.WriteFile(nonSource, []byte("not a source file"), 0o644); err != nil {
		t.Fatalf("write non-source file: %v", err)
	}

	artifact, parseErrs, err := Collect([]string{nonSource})
	if err != nil {
		t.Fatalf("collect error: %v", err)
	}
	if len(parseErrs) != 0 {
		t.Fatalf("parseErrs len = %d, want 0", len(parseErrs))
	}
	if len(artifact.Fields) != 0 {
		t.Fatalf("fields len = %d, want 0", len(artifact.Fields))
	}
}

func TestCollect_ReturnsErrorWhenReadingFileFails(t *testing.T) {
	tmp := t.TempDir()
	brokenPath := filepath.Join(tmp, "broken.go")

	if err := os.Symlink(filepath.Join(tmp, "missing.go"), brokenPath); err != nil {
		lower := strings.ToLower(err.Error())
		if runtime.GOOS == "windows" || strings.Contains(lower, "operation not permitted") {
			t.Skipf("symlink not supported in this environment: %v", err)
		}
		t.Fatalf("create symlink: %v", err)
	}

	_, _, err := Collect([]string{tmp})
	if err == nil {
		t.Fatalf("expected collect error from unreadable/broken file")
	}
	if !strings.Contains(err.Error(), "read") {
		t.Fatalf("error = %q, want read error", err)
	}
}
