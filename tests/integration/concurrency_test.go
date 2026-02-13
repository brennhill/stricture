// concurrency_test.go — Integration checks for --concurrency behavior.
//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeFile(t *testing.T, dir string, name string, body string) {
	t.Helper()
	fullPath := filepath.Join(dir, name)
	if err := os.WriteFile(fullPath, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestConcurrencyLevelPreservesOutput(t *testing.T) {
	tmp := t.TempDir()
	for _, name := range []string{"a.go", "b.go", "c.go"} {
		writeFile(t, tmp, name, "package main\n\nfunc main() {}\n")
	}

	stdout1, stderr1, code1 := runInDir(t, tmp,
		"--format", "json",
		"--rule", "CONV-file-header",
		"--concurrency", "1",
		".",
	)
	if code1 != 1 {
		t.Fatalf("concurrency=1 run should fail with violations: code=%d stderr=%q stdout=%q", code1, stderr1, stdout1)
	}

	stdout2, stderr2, code2 := runInDir(t, tmp,
		"--format", "json",
		"--rule", "CONV-file-header",
		"--concurrency", "4",
		".",
	)
	if code2 != 1 {
		t.Fatalf("concurrency=4 run should fail with violations: code=%d stderr=%q stdout=%q", code2, stderr2, stdout2)
	}
	if stderr1 != stderr2 {
		t.Fatalf("stderr should be stable across concurrency levels")
	}

	got1 := normalizeLintJSON(t, stdout1)
	got2 := normalizeLintJSON(t, stdout2)
	if !reflect.DeepEqual(got1, got2) {
		t.Fatalf("output should remain equivalent across concurrency levels")
	}
}

func TestInvalidConcurrencyValueExitsTwo(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, tmp, "a.go", "package main\n\nfunc main() {}\n")

	_, stderr, code := runInDir(t, tmp,
		"--rule", "CONV-file-header",
		"--concurrency", "0",
		".",
	)
	if code != 2 {
		t.Fatalf("--concurrency 0 exit code = %d, want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "concurrency") {
		t.Fatalf("stderr should explain invalid concurrency, got %q", stderr)
	}
}
