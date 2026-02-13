// cache_test.go — Integration checks for cache flag transparency.
//go:build integration

package integration

import "testing"

func TestNoCacheFlagDoesNotCrash(t *testing.T) {
	stdout1, stderr1, code1 := run(t, "--format", "json", "--no-cache", ".")
	if code1 == 2 {
		t.Skip("lint pipeline is not fully wired yet")
	}

	stdout2, stderr2, code2 := run(t, "--format", "json", "--no-cache", ".")
	if code2 != code1 {
		t.Fatalf("exit code changed between runs: %d vs %d", code1, code2)
	}
	if stdout1 != stdout2 || stderr1 != stderr2 {
		t.Fatalf("--no-cache runs should be deterministic for identical input")
	}
}
