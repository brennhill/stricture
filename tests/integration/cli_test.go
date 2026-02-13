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
	"strings"
	"testing"
)

func projectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}

// binaryPath returns the path to the stricture binary.
func binaryPath(t *testing.T) string {
	t.Helper()
	root := projectRoot(t)
	bin := filepath.Join(root, "bin", "stricture")

	if _, err := os.Stat(bin); err == nil {
		return bin
	}

	cmd := exec.Command("go", "build", "-o", bin, "./cmd/stricture")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build stricture binary: %v\n%s", err, string(out))
	}

	if _, err := os.Stat(bin); err != nil {
		t.Fatalf("stricture binary missing after build: %v", err)
	}
	return bin
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
		t.Fatalf("failed to run stricture: %v", err)
	}

	return outBuf.String(), errBuf.String(), exitCode
}

// === Exit code tests ===

func TestVersionExitsZero(t *testing.T) {
	stdout, _, code := run(t, "--version")
	if code != 0 {
		t.Errorf("--version exit code = %d, want 0", code)
	}
	if !strings.Contains(stdout, "stricture version") {
		t.Errorf("--version output = %q, want to contain 'stricture version'", stdout)
	}
}

func TestHelpExitsZero(t *testing.T) {
	_, _, code := run(t, "--help")
	if code != 0 {
		t.Errorf("--help exit code = %d, want 0", code)
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

	// Compare as parsed JSON (ignore whitespace differences)
	var got, want interface{}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("parse output JSON: %v", err)
	}

	expected, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}
	if err := json.Unmarshal(expected, &want); err != nil {
		t.Fatalf("parse golden JSON: %v", err)
	}

	gotJSON, _ := json.Marshal(got)
	wantJSON, _ := json.Marshal(want)
	if string(gotJSON) != string(wantJSON) {
		t.Errorf("JSON output differs from golden file")
	}
}

// === Concurrent safety test ===

func TestConcurrentRuns(t *testing.T) {
	// Run 5 instances of stricture simultaneously.
	// They must not interfere with each other (cache corruption, etc.)
	t.Parallel()

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
