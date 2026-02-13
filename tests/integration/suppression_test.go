// suppression_test.go — Integration checks for inline suppression behavior.
//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDisableFileSuppressionSkipsRule(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "user_service.ts")
	source := "// stricture-disable-file CONV-file-header\nexport const value = 1;\n"
	if err := os.WriteFile(target, []byte(source), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	_, stderr, code := run(t, "--rule", "CONV-file-header", target)
	if code != 0 {
		t.Fatalf("suppressed rule should not fail, exit=%d stderr=%q", code, stderr)
	}
}

