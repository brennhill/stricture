// plugin_test.go — Integration checks for custom plugin rules.
//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestYAMLPluginRuleRunsViaConfig(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "main.go")
	if err := os.WriteFile(src, []byte("package main\nimport \"fmt\"\nfunc main(){fmt.Print(\"x\")}\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	pluginPath := filepath.Join(tmp, "custom.yml")
	pluginContent := `rules:
  - id: CUSTOM-no-fmt-print
    category: custom
    severity: error
    description: "Disallow fmt.Print"
    match:
      languages: ["go"]
      paths: ["**/*.go"]
    check:
      must_not_contain:
        pattern: "fmt\\.Print"
        message: "Use structured logging"
`
	if err := os.WriteFile(pluginPath, []byte(pluginContent), 0o644); err != nil {
		t.Fatalf("write plugin: %v", err)
	}

	cfgPath := filepath.Join(tmp, ".stricture.yml")
	cfg := `version: "1.0"
plugins:
  - ./custom.yml
`
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	stdout, stderr, code := runInDir(t, tmp, "--config", cfgPath, "--rule", "CUSTOM-no-fmt-print", src)
	if code != 1 {
		t.Fatalf("plugin lint exit code = %d, want 1\nstderr=%q\nstdout=%q", code, stderr, stdout)
	}
	if !strings.Contains(stdout, "CUSTOM-no-fmt-print") {
		t.Fatalf("expected plugin rule id in output, got %q", stdout)
	}
}
