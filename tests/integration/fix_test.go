// fix_test.go — Integration checks for fix command behavior.
//go:build integration

package integration

import "testing"

func TestFixFlagBehavior(t *testing.T) {
	_, _, code := run(t, "--fix", ".")
	if code != 2 {
		t.Fatalf("--fix exit code = %d, want 2 until flag is implemented", code)
	}
}
