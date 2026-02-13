// determinism_test.go — Integration determinism checks.
//go:build integration

package integration

import (
	"strings"
	"testing"
)

func TestVersionOutputDeterminism(t *testing.T) {
	const runs = 5
	outputs := make([]string, 0, runs)

	for i := 0; i < runs; i++ {
		stdout, stderr, code := run(t, "--version")
		if code != 0 {
			t.Fatalf("run %d: --version exit code = %d, want 0", i, code)
		}
		if strings.TrimSpace(stderr) != "" {
			t.Fatalf("run %d: unexpected stderr: %q", i, stderr)
		}
		outputs = append(outputs, strings.TrimSpace(stdout))
	}

	for i := 1; i < len(outputs); i++ {
		if outputs[i] != outputs[0] {
			t.Fatalf("determinism violation: run %d output %q differs from run 0 %q", i, outputs[i], outputs[0])
		}
	}
}
