// characterization_test.go — Pins current CLI behavior of previously untested commands before refactoring.
//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	policyHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	policyURL  = "https://policy.example.com/p.yml"

	identityAnnotation = `// strict-source annotation_schema_version=1 field_id=response_user_id field=response.user_id source_system=Identity source_version=v1 min_supported_source_version=v1 transform_type=normalize merge_strategy=single_source break_policy=additive_only confidence=declared data_classification=internal owner=team.identity escalation=slack:#identity-oncall contract_test_id=ci://contracts/identity-user-id introduced_at=2026-01-10 sources=api:identity.GetUser#response.id@cross_repo?contract_ref=git+https://github.com/acme/identity//openapi.yaml@a1b2 flow="from @Identity normalized @self" note="normalized by UserNormalizer.Apply"` + "\n"
)

// writeFiles creates files (relative path -> content) under dir.
func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir for %s: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
}

// assertGolden compares got with testdata/characterization/<name>; UPDATE_GOLDEN=1 rewrites it.
func assertGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join(projectRoot(t), "tests", "integration", "testdata", "characterization", name)
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir golden dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden %s: %v", name, err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v. Run with UPDATE_GOLDEN=1 to create it", name, err)
	}
	if string(want) != got {
		t.Errorf("output differs from golden %s\n--- want\n%s\n--- got\n%s", name, want, got)
	}
}

func expect(t *testing.T, label string, code, wantCode int, stderr, wantStderr string) {
	t.Helper()
	if code != wantCode {
		t.Errorf("%s: exit code = %d, want %d (stderr=%q)", label, code, wantCode, stderr)
	}
	if wantStderr != "" && !strings.Contains(stderr, wantStderr) {
		t.Errorf("%s: stderr = %q, want to contain %q", label, stderr, wantStderr)
	}
}

// === policy ===

func TestPolicyCommand(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"good.yml":   "strict:policy_url: " + policyURL + "\nstrict:policy_sha256: " + policyHash + "\nversion: \"1.0\"\n",
		"nosha.yml":  "strict:policy_url: " + policyURL + "\nversion: \"1.0\"\n",
		"badsha.yml": "strict:policy_url: " + policyURL + "\nstrict:policy_sha256: zz\nversion: \"1.0\"\n",
		"nourl.yml":  "version: \"1.0\"\n",
	})

	cases := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{"no subcommand prints usage", []string{"policy"}, 0, "Usage: strict policy <command>", ""},
		{"help", []string{"policy", "--help"}, 0, "verify-ref      Verify repo", ""},
		{"verify-ref help", []string{"policy", "verify-ref", "--help"}, 0, "CI/CD guard", ""},
		{"unknown subcommand", []string{"policy", "bogus"}, 2, "", `unknown policy command "bogus"`},
		{"match by flag", []string{"policy", "verify-ref", "--config", "good.yml", "--expected-url", policyURL}, 0,
			"Policy binding verified: strict:policy_url=" + policyURL + " source=flag:expected-url", ""},
		{"mismatch", []string{"policy", "verify-ref", "--config", "good.yml", "--expected-url", "https://other.example.com"}, 1, "",
			`strict:policy_url mismatch: got "` + policyURL + `" want "https://other.example.com"`},
		{"nothing to compare", []string{"policy", "verify-ref", "--config", "good.yml"}, 2, "",
			"provide --expected-url, --expected-url-env, or at least one --allow-url"},
		{"allowlist hit", []string{"policy", "verify-ref", "--config", "good.yml", "--allow-url", policyURL}, 0, "Policy binding verified", ""},
		{"allowlist miss", []string{"policy", "verify-ref", "--config", "good.yml", "--allow-url", "https://nope"}, 1, "", "is not in allowlist"},
		{"missing url in config", []string{"policy", "verify-ref", "--config", "nourl.yml", "--expected-url", "x"}, 2, "", "is missing strict:policy_url"},
		{"missing config file", []string{"policy", "verify-ref", "--config", "missing.yml", "--expected-url", "x"}, 2, "", "read config missing.yml"},
		{"require-sha satisfied", []string{"policy", "verify-ref", "--config", "good.yml", "--expected-url", policyURL, "--require-sha"}, 0, "Policy binding verified", ""},
		{"require-sha missing", []string{"policy", "verify-ref", "--config", "nosha.yml", "--expected-url", policyURL, "--require-sha"}, 1, "", "strict:policy_sha256 is required but missing"},
		{"malformed sha", []string{"policy", "verify-ref", "--config", "badsha.yml", "--expected-url", policyURL}, 1, "", "must be 64 hex chars"},
		{"positional args rejected", []string{"policy", "verify-ref", "extra"}, 2, "", "does not accept positional args: extra"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, code := runInDir(t, dir, tc.args...)
			expect(t, tc.name, code, tc.wantCode, stderr, tc.wantStderr)
			if tc.wantStdout != "" && !strings.Contains(stdout, tc.wantStdout) {
				t.Errorf("stdout = %q, want to contain %q", stdout, tc.wantStdout)
			}
		})
	}
}

func TestPolicyVerifyRefReadsExpectedURLFromEnv(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{"good.yml": "strict:policy_url: " + policyURL + "\nversion: \"1.0\"\n"})
	t.Setenv("STRICTURE_POLICY_URL", policyURL)
	stdout, stderr, code := runInDir(t, dir, "policy", "verify-ref", "--config", "good.yml")
	expect(t, "env url", code, 0, stderr, "")
	if !strings.Contains(stdout, "source=env:STRICTURE_POLICY_URL") {
		t.Errorf("stdout = %q, want source=env:STRICTURE_POLICY_URL", stdout)
	}
}

// === inspect ===

const goSample = `// svc.go — Sample.
package svc

import (
	"errors"
	"fmt"
)

// User is a user.
type User struct {
	ID   string ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}

func GetUser(id string) (*User, error) {
	if id == "" {
		return nil, errors.New("empty")
	}
	return &User{ID: id}, nil
}

func (u *User) String() string { return fmt.Sprint(u.ID) }
`

func TestInspectGoFileMatchesGolden(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{"svc.go": goSample})
	stdout, stderr, code := runInDir(t, dir, "inspect", "svc.go")
	expect(t, "inspect go", code, 0, stderr, "")
	assertGolden(t, "inspect_go.json", stdout)
}

func TestInspectErrors(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{"f.unknownext": "x", "blob.go": "ab\x00cd\x00\x00"})
	cases := []struct {
		name, wantStderr string
		args             []string
	}{
		{"no argument", "inspect requires a file path argument", []string{"inspect"}},
		{"missing file", "file not found: nope.go", []string{"inspect", "nope.go"}},
		{"unknown language", `no language adapter for ".unknownext" files. Supported: go, typescript, javascript, python, java`, []string{"inspect", "f.unknownext"}},
		{"binary file", "cannot inspect binary file: blob.go", []string{"inspect", "blob.go"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, stderr, code := runInDir(t, dir, tc.args...)
			expect(t, tc.name, code, 2, stderr, tc.wantStderr)
		})
	}
}

// === lineage ===

func TestInspectLineageCommand(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{"ok.go": identityAnnotation, "bad.go": "// strict-source bogus=1\n"})

	stdout, stderr, code := runInDir(t, dir, "inspect-lineage", "ok.go")
	expect(t, "valid annotation", code, 0, stderr, "")
	assertGolden(t, "inspect_lineage_ok.json", stdout)

	stdout, stderr, code = runInDir(t, dir, "inspect-lineage", "bad.go")
	expect(t, "invalid annotation", code, 1, stderr, "")
	assertGolden(t, "inspect_lineage_bad.json", stdout)

	stdout, stderr, code = runInDir(t, dir, "inspect-lineage")
	expect(t, "no argument", code, 2, stderr, "inspect-lineage requires a file path argument")
	if !strings.Contains(stdout, "Usage: strict inspect-lineage <file>") {
		t.Errorf("usage should print to stdout, got %q", stdout)
	}

	_, stderr, code = runInDir(t, dir, "inspect-lineage", "nope.go")
	expect(t, "missing file", code, 1, stderr, "cannot read nope.go")
}

func TestLineageExportCommand(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{"a/l.go": identityAnnotation, "bad.go": "// strict-source bogus=1\n"})

	stdout, stderr, code := runInDir(t, dir, "lineage-export", "a")
	expect(t, "export to stdout", code, 0, stderr, "")
	assertGolden(t, "lineage_export_default.json", stdout)

	stdout, stderr, code = runInDir(t, dir, "lineage-export", "--out", "out.json", "a")
	expect(t, "export to file", code, 0, stderr, "")
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("stdout should be empty with --out, got %q", stdout)
	}
	written, err := os.ReadFile(filepath.Join(dir, "out.json"))
	if err != nil {
		t.Fatalf("read out.json: %v", err)
	}
	if !strings.Contains(string(written), `"field_id": "response_user_id"`) {
		t.Errorf("out.json missing field: %s", written)
	}

	stdout, stderr, code = runInDir(t, dir, "lineage-export", "bad.go")
	expect(t, "strict parse error", code, 1, stderr, "Lineage parse errors (1)")
	assertGolden(t, "lineage_export_parse_error.json", stdout)

	_, stderr, code = runInDir(t, dir, "lineage-export", "--strict=false", "bad.go")
	expect(t, "non-strict parse error", code, 0, stderr, "Lineage parse errors (1)")
}

func TestLineageDiffCommand(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{"a/l.go": identityAnnotation, "empty/x.go": "package x\n"})
	if _, stderr, code := runInDir(t, dir, "lineage-export", "--out", "base.json", "a"); code != 0 {
		t.Fatalf("export base: %d %s", code, stderr)
	}
	if _, stderr, code := runInDir(t, dir, "lineage-export", "--out", "head.json", "empty"); code != 0 {
		t.Fatalf("export head: %d %s", code, stderr)
	}

	_, stderr, code := runInDir(t, dir, "lineage-diff")
	expect(t, "missing flags", code, 2, stderr, "--base and --head are required")

	stdout, stderr, code := runInDir(t, dir, "lineage-diff", "--base", "base.json", "--head", "base.json")
	expect(t, "no drift", code, 0, stderr, "")
	assertGolden(t, "lineage_diff_no_drift.json", stdout)

	stdout, stderr, code = runInDir(t, dir, "lineage-diff", "--base", "base.json", "--head", "head.json")
	expect(t, "high drift blocks", code, 1, stderr, "")
	assertGolden(t, "lineage_diff_field_removed.json", stdout)

	_, stderr, code = runInDir(t, dir, "lineage-diff", "--base", "base.json", "--head", "head.json", "--mode", "warn")
	expect(t, "warn mode", code, 0, stderr, "WARN: drift at/above high detected, but mode=warn so exit code remains 0")

	_, stderr, code = runInDir(t, dir, "lineage-diff", "--base", "base.json", "--head", "head.json", "--fail-on", "none")
	expect(t, "fail-on none", code, 0, stderr, "")

	_, stderr, code = runInDir(t, dir, "lineage-diff", "--base", "base.json", "--head", "head.json", "--fail-on", "bogus")
	expect(t, "invalid severity", code, 2, stderr, `invalid severity "bogus" (valid: high|medium|low|info|none)`)

	_, stderr, code = runInDir(t, dir, "lineage-diff", "--base", "nope.json", "--head", "head.json")
	expect(t, "missing base", code, 1, stderr, "load base artifact")

	stdout, stderr, code = runInDir(t, dir, "lineage-diff", "--base", "base.json", "--head", "head.json", "--out", "d.json")
	expect(t, "diff to file", code, 1, stderr, "")
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("stdout should be empty with --out, got %q", stdout)
	}
	if _, err := os.Stat(filepath.Join(dir, "d.json")); err != nil {
		t.Errorf("--out file not written: %v", err)
	}
}

func TestLineageEscalateCommand(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{"a/l.go": identityAnnotation})
	if _, stderr, code := runInDir(t, dir, "lineage-export", "--out", "base.json", "a"); code != 0 {
		t.Fatalf("export base: %d %s", code, stderr)
	}

	_, stderr, code := runInDir(t, dir, "lineage-escalate")
	expect(t, "missing flags", code, 2, stderr, "--service and --artifact are required")

	stdout, stderr, code := runInDir(t, dir, "lineage-escalate", "--service", "Identity", "--artifact", "base.json")
	expect(t, "escalate", code, 0, stderr, "")
	assertGolden(t, "lineage_escalate.json", stdout)

	_, stderr, code = runInDir(t, dir, "lineage-escalate", "--service", "Identity", "--artifact", "nope.json")
	expect(t, "missing artifact", code, 1, stderr, "load artifact")

	_, stderr, code = runInDir(t, dir, "lineage-escalate", "--service", "Identity", "--artifact", "base.json", "--systems", "nope.yml")
	expect(t, "missing systems registry", code, 1, stderr, "load systems registry")
}

// === fix subcommand ===

func TestFixSubcommandAppliesFixes(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		".stricture.yml": "version: \"1.0\"\nrules:\n  CONV-file-header: [error, { pattern: \"// {filename} — {purpose}\" }]\n",
		"no_header.go":   "package x\n\nfunc F() {}\n",
	})
	stdout, stderr, code := runInDir(t, dir, "fix", ".")
	expect(t, "fix", code, 0, stderr, "")
	if !strings.Contains(stdout, "Fixes: 1 operation(s) (apply)") {
		t.Errorf("stdout = %q, want fix summary", stdout)
	}
	got, err := os.ReadFile(filepath.Join(dir, "no_header.go"))
	if err != nil {
		t.Fatalf("read fixed file: %v", err)
	}
	if !strings.HasPrefix(string(got), "// no_header.go — TODO: describe purpose\n") {
		t.Errorf("header not applied: %q", got)
	}
}

func TestFixSubcommandHelp(t *testing.T) {
	dir := t.TempDir()
	stdout, stderr, code := runInDir(t, dir, "fix", "--help")
	expect(t, "fix --help", code, 0, stderr, "")
	if !strings.Contains(stdout, "Usage: strict fix [options] [paths...]") {
		t.Errorf("stdout should describe fix, got %q", stdout)
	}
	for _, want := range []string{"--dry-run", "--fix-backup"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("fix usage missing %s: %q", want, stdout)
		}
	}
	if strings.Contains(stdout+stderr, "Usage of lint") {
		t.Errorf("fix help must not present itself as lint: %q %q", stdout, stderr)
	}
}

func TestFixSubcommandDryRunDoesNotModifyFiles(t *testing.T) {
	for _, flagName := range []string{"--dry-run", "--fix-dry-run"} {
		t.Run(flagName, func(t *testing.T) {
			dir := t.TempDir()
			original := "package x\n\nfunc F() {}\n"
			writeFiles(t, dir, map[string]string{
				".stricture.yml": "version: \"1.0\"\nrules:\n  CONV-file-header: [error, { pattern: \"// {filename} — {purpose}\" }]\n",
				"no_header.go":   original,
			})
			stdout, stderr, code := runInDir(t, dir, "fix", flagName, ".")
			// The violation is still present after a preview, so lint's exit code is 1.
			expect(t, "fix "+flagName, code, 1, stderr, "")
			if !strings.Contains(stdout, "Fixes: 1 operation(s) (dry-run)") {
				t.Errorf("stdout = %q, want dry-run summary", stdout)
			}
			got, err := os.ReadFile(filepath.Join(dir, "no_header.go"))
			if err != nil {
				t.Fatalf("read file: %v", err)
			}
			if string(got) != original {
				t.Errorf("dry run modified the file: %q", got)
			}
		})
	}
}
