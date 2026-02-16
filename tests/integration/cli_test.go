// cli_test.go — Integration tests for the Stricture CLI.
//
// These tests run the actual compiled binary and verify end-to-end behavior:
// exit codes, output format, flag handling, and error reporting.
//
// Build requirement: run `make build` before running these tests.
// Run: go test -tags=integration -timeout=120s ./tests/integration/...

//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
)

var (
	buildBinaryOnce sync.Once
	buildBinaryPath string
	buildBinaryErr  error
)

func projectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}

// binaryPath returns the path to the strict binary.
func binaryPath(t *testing.T) string {
	t.Helper()
	root := projectRoot(t)
	bin := filepath.Join(root, "bin", "strict")

	buildBinaryOnce.Do(func() {
		cmd := exec.Command("go", "build", "-o", bin, "./cmd/strict")
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			buildBinaryErr = fmt.Errorf("build strict binary: %w\n%s", err, string(out))
			return
		}
		if _, err := os.Stat(bin); err != nil {
			buildBinaryErr = fmt.Errorf("strict binary missing after build: %w", err)
			return
		}
		buildBinaryPath = bin
	})

	if buildBinaryErr != nil {
		t.Fatalf("%v", buildBinaryErr)
	}
	return buildBinaryPath
}

// run executes the stricture binary with the given args and returns stdout, stderr, and exit code.
func run(t *testing.T, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	bin := binaryPath(t)
	cmd := exec.Command(bin, args...)

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	exitCode = 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		t.Fatalf("failed to run strict: %v", err)
	}

	return outBuf.String(), errBuf.String(), exitCode
}

// runInDir executes strict in a specific working directory.
func runInDir(t *testing.T, dir string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()
	bin := binaryPath(t)
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	exitCode = 0
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		t.Fatalf("failed to run strict in %s: %v", dir, err)
	}

	return outBuf.String(), errBuf.String(), exitCode
}

// === Exit code tests ===

func TestBinaryPathRebuildsWhenExistingBinaryIsInvalid(t *testing.T) {
	root := projectRoot(t)
	bin := filepath.Join(root, "bin", "strict")

	original, readErr := os.ReadFile(bin)
	hadOriginal := readErr == nil
	if readErr != nil && !os.IsNotExist(readErr) {
		t.Fatalf("read existing binary: %v", readErr)
	}

	if err := os.MkdirAll(filepath.Dir(bin), 0o755); err != nil {
		t.Fatalf("create bin dir: %v", err)
	}
	if err := os.WriteFile(bin, []byte("not-a-valid-binary"), 0o755); err != nil {
		t.Fatalf("write invalid binary: %v", err)
	}
	defer func() {
		if hadOriginal {
			_ = os.WriteFile(bin, original, 0o755)
			return
		}
		_ = os.Remove(bin)
	}()
	buildBinaryOnce = sync.Once{}
	buildBinaryPath = ""
	buildBinaryErr = nil

	stdout, stderr, code := run(t, "--version")
	if code != 0 {
		t.Fatalf("version should still run after rebuild (exit=%d)\nstderr=%q\nstdout=%q", code, stderr, stdout)
	}
	if !strings.Contains(stdout, "strict version") {
		t.Fatalf("version output should include marker, got %q", stdout)
	}
}

func TestVersionExitsZero(t *testing.T) {
	stdout, _, code := run(t, "--version")
	if code != 0 {
		t.Errorf("--version exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "strict version") {
		t.Errorf("--version output = %q, want to contain 'strict version'", stdout)
	}
}

func TestHelpExitsZero(t *testing.T) {
	_, _, code := run(t, "--help")
	if code != 0 {
		t.Errorf("--help exit code = %d, want 0", code)
	}
}

func TestNoArgsRunsDefaultLint(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "bad.go")
	if err := os.WriteFile(target, []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	stdout, stderr, code := runInDir(t, tmp)
	if code != 1 {
		t.Fatalf("no-args default lint exit code = %d, want 1\nstderr=%q\nstdout=%q", code, stderr, stdout)
	}
	if !strings.Contains(stdout, "CONV-file-header") {
		t.Fatalf("default lint output should include violation, got %q", stdout)
	}
}

func TestLintParsesFlagsAfterPaths(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "bad.go")
	if err := os.WriteFile(target, []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	stdout, stderr, code := runInDir(t, tmp, ".", "--format", "json", "--rule", "CONV-file-header")
	if code != 1 {
		t.Fatalf("lint should honor flags after paths (exit=%d)\nstderr=%q\nstdout=%q", code, stderr, stdout)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("output should be JSON when --format json follows path: %v\noutput=%q", err, stdout)
	}
}

func TestInvalidFlagExitsTwo(t *testing.T) {
	_, _, code := run(t, "--nonexistent-flag")
	if code != 2 {
		t.Errorf("invalid flag exit code = %d, want 2", code)
	}
}

func TestInvalidFormatExitsTwo(t *testing.T) {
	_, _, code := run(t, "--format", "xml", ".")
	if code != 2 {
		t.Errorf("invalid format exit code = %d, want 2", code)
	}
}

func TestUnknownSubcommandExitsTwo(t *testing.T) {
	_, stderr, code := run(t, "foobar", "src/")
	if code != 2 {
		t.Fatalf("unknown subcommand exit code = %d, want 2", code)
	}
	if !strings.Contains(strings.ToLower(stderr), "unknown command") {
		t.Fatalf("stderr should mention unknown command, got %q", stderr)
	}
}

// === Output format tests ===

func TestJSONOutputIsValidJSON(t *testing.T) {
	stdout, _, code := run(t, "--format", "json", ".")
	if code == 2 {
		t.Fatalf("lint returned operational error (exit 2)")
	}

	// Must be valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Errorf("JSON output is not valid JSON: %v\nOutput: %s", err, stdout)
	}

	// Must have required fields
	requiredFields := []string{"violations", "summary"}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("JSON output missing required field %q", field)
		}
	}
}

func TestStdoutVsStderr(t *testing.T) {
	// Results go to stdout, errors/diagnostics go to stderr
	stdout, stderr, code := run(t, "--format", "json", ".")
	if code == 2 {
		t.Fatalf("lint returned operational error (exit 2)")
	}

	// Stdout should be parseable (JSON or text results)
	if stdout == "" && code == 0 {
		// Empty stdout is OK for --version/--help but not for lint
		t.Log("Note: stdout is empty — lint may not be implemented yet")
	}

	// Stderr should not contain JSON results
	if strings.HasPrefix(strings.TrimSpace(stderr), "{") {
		t.Error("stderr appears to contain JSON results — results should go to stdout")
	}
}

// === Self-lint test ===

func TestSelfLint(t *testing.T) {
	// Stricture must be able to lint its own codebase
	_, _, code := run(t, ".")
	if code != 0 {
		t.Errorf("self-lint exit code = %d, want 0 (Stricture must pass its own rules)", code)
	}
}

// === Golden file tests ===

func TestGoldenTextOutput(t *testing.T) {
	goldenDir := filepath.Join(projectRoot(t), "tests", "golden", "input")
	if _, err := os.Stat(goldenDir); err != nil {
		t.Fatalf("golden input directory not found: %v", err)
	}

	stdout, _, code := run(t, "--format", "text", goldenDir)
	if code == 2 {
		t.Fatalf("lint returned operational error (exit 2)")
	}

	goldenFile := filepath.Join(projectRoot(t), "tests", "golden", "output-text.txt")
	if _, err := os.Stat(goldenFile); err != nil {
		t.Fatalf("golden output file not found: %v", err)
	}

	expected, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}

	if stdout != string(expected) {
		t.Errorf("output differs from golden file\n--- got ---\n%s\n--- want ---\n%s", stdout, string(expected))
	}
}

func TestGoldenJSONOutput(t *testing.T) {
	goldenDir := filepath.Join(projectRoot(t), "tests", "golden", "input")
	if _, err := os.Stat(goldenDir); err != nil {
		t.Fatalf("golden input directory not found: %v", err)
	}

	stdout, _, code := run(t, "--format", "json", goldenDir)
	if code == 2 {
		t.Fatalf("lint returned operational error (exit 2)")
	}

	goldenFile := filepath.Join(projectRoot(t), "tests", "golden", "output.json")
	if _, err := os.Stat(goldenFile); err != nil {
		t.Fatalf("golden output file not found: %v", err)
	}

	// Compare as parsed JSON after canonicalization:
	// ignore dynamic elapsedMs and order violations deterministically.
	var gotReport, wantReport map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &gotReport); err != nil {
		t.Fatalf("parse output JSON: %v", err)
	}

	expected, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}
	if err := json.Unmarshal(expected, &wantReport); err != nil {
		t.Fatalf("parse golden JSON: %v", err)
	}

	canonicalizeReport(gotReport)
	canonicalizeReport(wantReport)

	gotJSON, _ := json.Marshal(gotReport)
	wantJSON, _ := json.Marshal(wantReport)
	if string(gotJSON) != string(wantJSON) {
		t.Errorf("JSON output differs from golden file\n--- got ---\n%s\n--- want ---\n%s", string(gotJSON), string(wantJSON))
	}
}

func canonicalizeReport(report map[string]interface{}) {
	if report == nil {
		return
	}
	if summary, ok := report["summary"].(map[string]interface{}); ok {
		// elapsedMs is runtime-dependent and not useful for golden stability.
		delete(summary, "elapsedMs")
	}

	violationsAny, ok := report["violations"].([]interface{})
	if !ok {
		return
	}

	violations := make([]map[string]interface{}, 0, len(violationsAny))
	for _, item := range violationsAny {
		v, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		violations = append(violations, v)
	}

	sort.SliceStable(violations, func(i, j int) bool {
		return violationSortKey(violations[i]) < violationSortKey(violations[j])
	})

	normalized := make([]interface{}, 0, len(violations))
	for _, v := range violations {
		normalized = append(normalized, v)
	}
	report["violations"] = normalized
}

func violationSortKey(v map[string]interface{}) string {
	return fmt.Sprintf(
		"%v|%v|%v|%v|%v",
		firstNonNil(v["RuleID"], v["ruleId"], v["rule"], v["rule_id"]),
		firstNonNil(v["File"], v["file"], v["path"]),
		firstNonNil(v["Line"], v["line"]),
		firstNonNil(v["Column"], v["column"]),
		firstNonNil(v["Message"], v["message"]),
	)
}

func firstNonNil(values ...interface{}) interface{} {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return ""
}

// === Concurrent safety test ===

func TestConcurrentRuns(t *testing.T) {
	// Run 5 instances of stricture simultaneously.
	// They must not interfere with each other (cache corruption, etc.)
	errs := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func() {
			_, _, code := run(t, "--version")
			if code != 0 {
				errs <- fmt.Errorf("exit code %d", code)
			} else {
				errs <- nil
			}
		}()
	}

	for i := 0; i < 5; i++ {
		if err := <-errs; err != nil {
			t.Errorf("concurrent run %d failed", i)
		}
	}
}
