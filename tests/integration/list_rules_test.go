// list_rules_test.go — Integration checks for list-rules output.
//go:build integration

package integration

import (
	"strings"
	"testing"
)

func TestListRulesIncludesFixabilityMetadata(t *testing.T) {
	stdout, stderr, code := run(t, "list-rules")
	if code != 0 {
		t.Fatalf("list-rules exit code = %d, want 0\nstderr=%q", code, stderr)
	}

	if !strings.Contains(stdout, "FIXABLE") {
		t.Fatalf("list-rules output missing FIXABLE column")
	}
	if !strings.Contains(stdout, "CONV-file-header") || !strings.Contains(stdout, "Yes") {
		t.Fatalf("list-rules output missing fixable metadata for known rule")
	}
	if !strings.Contains(stdout, "CTR-manifest-conformance") || !strings.Contains(stdout, "requires manifest") {
		t.Fatalf("list-rules output missing manifest metadata")
	}
}
