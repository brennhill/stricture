// init_test.go — Integration checks for strict init command.
//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCreatesDefaultConfig(t *testing.T) {
	tmp := t.TempDir()

	stdout, stderr, code := runInDir(t, tmp, "init")
	if code != 0 {
		t.Fatalf("init exit code = %d, want 0\nstderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, ".stricture.yml") {
		t.Fatalf("init output should mention config path, got %q", stdout)
	}

	cfgPath := filepath.Join(tmp, ".stricture.yml")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read created config: %v", err)
	}
	if !strings.Contains(string(data), "version:") {
		t.Fatalf("created config missing version field")
	}
}

func TestInitRefusesOverwriteWithoutForce(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, ".stricture.yml")
	if err := os.WriteFile(cfgPath, []byte("version: \"1.0\"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, stderr, code := runInDir(t, tmp, "init")
	if code != 2 {
		t.Fatalf("init overwrite exit code = %d, want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "already exists") {
		t.Fatalf("stderr should mention existing config, got %q", stderr)
	}
}

func TestInitForceOverwritesExistingConfig(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, ".stricture.yml")
	if err := os.WriteFile(cfgPath, []byte("version: \"old\"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, stderr, code := runInDir(t, tmp, "init", "--force")
	if code != 0 {
		t.Fatalf("init --force exit code = %d, want 0\nstderr=%q", code, stderr)
	}

	after, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(after), "CONV-file-header") {
		t.Fatalf("forced init should rewrite with default config, got %q", string(after))
	}
}
