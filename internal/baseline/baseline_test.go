// baseline_test.go — Tests baseline bootstrap, comparison, and error handling.
package baseline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stricture/stricture/internal/model"
)

func v(rule, file string, line int, msg string) model.Violation {
	return model.Violation{RuleID: rule, FilePath: file, StartLine: line, Message: msg}
}

func TestApplyDisabledWithBlankPath(t *testing.T) {
	violations := []model.Violation{v("R", "a.go", 1, "m")}
	state, err := Apply("  ", &violations, Options{})
	if err != nil || state.Enabled || len(violations) != 1 {
		t.Fatalf("blank path should be a no-op, got state=%+v err=%v violations=%d", state, err, len(violations))
	}
}

func TestApplyRejectsNilViolations(t *testing.T) {
	if _, err := Apply("b.json", nil, Options{}); err == nil {
		t.Fatal("nil violations pointer should fail")
	}
}

func TestApplyMissingBaselineWithoutBootstrapFails(t *testing.T) {
	violations := []model.Violation{}
	path := filepath.Join(t.TempDir(), "b.json")
	_, err := Apply(path, &violations, Options{BootstrapIfMissing: false})
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("error = %v, want does-not-exist", err)
	}
}

func TestApplyBootstrapThenCompare(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "b.json")
	first := []model.Violation{v("R2", "b.go", 2, " m2 "), v("R1", "a.go", 5, "m1")}

	state, err := Apply(path, &first, Options{BootstrapIfMissing: true})
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	if !state.Bootstrapped || state.Suppressed != 2 || state.EntryCount != 2 || len(first) != 0 {
		t.Fatalf("bootstrap state = %+v violations=%d", state, len(first))
	}
	if state.Entries[0].FilePath != "a.go" || state.Entries[1].Message != "m2" {
		t.Errorf("entries should be sorted and trimmed: %+v", state.Entries)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("baseline file not written: %v", err)
	}

	// Second run: one old violation remains, one is resolved, one is new.
	second := []model.Violation{v("R1", "a.go", 5, "m1"), v("R3", "c.go", 9, "new")}
	state, err = Apply(path, &second, Options{})
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if state.Bootstrapped || state.Suppressed != 1 || len(second) != 1 || second[0].RuleID != "R3" {
		t.Fatalf("compare state = %+v remaining=%+v", state, second)
	}
	if len(state.Added) != 1 || state.Added[0].RuleID != "R3" {
		t.Errorf("Added = %+v, want R3", state.Added)
	}
	if len(state.Resolved) != 1 || state.Resolved[0].RuleID != "R2" {
		t.Errorf("Resolved = %+v, want R2", state.Resolved)
	}
}

func TestApplyRejectsMalformedBaseline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "b.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	violations := []model.Violation{}
	if _, err := Apply(path, &violations, Options{}); err == nil || !strings.Contains(err.Error(), "parse baseline") {
		t.Fatalf("error = %v, want parse baseline", err)
	}
}
