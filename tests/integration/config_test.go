// config_test.go — Integration checks for config behavior.
//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateConfig_Success(t *testing.T) {
	tmp := t.TempDir()
	cfg := filepath.Join(tmp, ".stricture.yml")
	content := "version: \"1.0\"\nrules:\n  CONV-file-naming: error\n"
	if err := os.WriteFile(cfg, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	stdout, stderr, code := run(t, "validate-config", cfg)
	if code != 0 {
		t.Fatalf("validate-config exit code = %d, want 0\nstderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "valid YAML") {
		t.Fatalf("stdout missing success marker: %q", stdout)
	}
}

func TestValidateConfig_InvalidYAML(t *testing.T) {
	tmp := t.TempDir()
	cfg := filepath.Join(tmp, ".stricture.yml")
	if err := os.WriteFile(cfg, []byte("rules:\n  bad: [\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, stderr, code := run(t, "validate-config", cfg)
	if code == 0 {
		t.Fatalf("validate-config must fail on invalid YAML")
	}
	if !strings.Contains(stderr, "invalid YAML") {
		t.Fatalf("stderr missing invalid YAML marker: %q", stderr)
	}
}
