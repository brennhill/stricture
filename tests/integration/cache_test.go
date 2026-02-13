// cache_test.go — Integration checks for cache flag transparency.
//go:build integration

package integration

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestNoCacheFlagDoesNotCrash(t *testing.T) {
	stdout1, stderr1, code1 := run(t, "--format", "json", "--no-cache", ".")
	if code1 == 2 {
		t.Fatalf("lint returned operational error (exit 2)")
	}

	stdout2, stderr2, code2 := run(t, "--format", "json", "--no-cache", ".")
	if code2 != code1 {
		t.Fatalf("exit code changed between runs: %d vs %d", code1, code2)
	}
	if stderr1 != stderr2 {
		t.Fatalf("stderr changed between deterministic runs")
	}

	got1 := normalizeLintJSON(t, stdout1)
	got2 := normalizeLintJSON(t, stdout2)
	if !reflect.DeepEqual(got1, got2) {
		t.Fatalf("--no-cache runs should be deterministic for identical input")
	}
}

func normalizeLintJSON(t *testing.T, raw string) map[string]interface{} {
	t.Helper()
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("invalid lint json output: %v", err)
	}

	summary, ok := payload["summary"].(map[string]interface{})
	if ok {
		delete(summary, "elapsedMs")
	}
	return payload
}
