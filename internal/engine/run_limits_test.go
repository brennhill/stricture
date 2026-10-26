// run_limits_test.go — Pins fail-fast, panic, suppression, and config-binding semantics of RunRules.
package engine

import (
	"reflect"
	"testing"

	"github.com/stricture/stricture/internal/model"
)

func viol(ruleID, file string, line int) model.Violation {
	return model.Violation{RuleID: ruleID, Severity: "error", FilePath: file, StartLine: line, Message: "m"}
}

func goFile(path, source string) *model.UnifiedFileModel {
	return &model.UnifiedFileModel{Path: path, Source: []byte(source)}
}

// configProbe records the RuleConfig it was invoked with.
type configProbe struct {
	fakeRule
	seen *model.RuleConfig
}

func (p configProbe) Check(_ *model.UnifiedFileModel, _ *model.ProjectContext, cfg model.RuleConfig) []model.Violation {
	*p.seen = cfg
	return nil
}

func TestRunRulesMaxViolations(t *testing.T) {
	three := func(id, file string) fakeRule {
		return fakeRule{id: id, violations: []model.Violation{viol(id, file, 1), viol(id, file, 2), viol(id, file, 3)}}
	}
	cases := []struct {
		name  string
		files []*model.UnifiedFileModel
		rules []model.Rule
		max   int
		want  []string // "rule@file:line"
	}{
		{
			name:  "stops mid-rule when the limit is reached",
			files: []*model.UnifiedFileModel{goFile("a.go", "x\n"), goFile("b.go", "x\n")},
			rules: []model.Rule{three("R1", "a.go")},
			max:   2,
			want:  []string{"R1@a.go:1", "R1@a.go:2"},
		},
		{
			name:  "limit spans rules within a file and skips later files",
			files: []*model.UnifiedFileModel{goFile("a.go", "x\n"), goFile("b.go", "x\n")},
			rules: []model.Rule{
				fakeRule{id: "R1", violations: []model.Violation{viol("R1", "a.go", 1), viol("R1", "a.go", 2)}},
				fakeRule{id: "R2", violations: []model.Violation{viol("R2", "a.go", 3), viol("R2", "a.go", 4)}},
			},
			max:  3,
			want: []string{"R1@a.go:1", "R1@a.go:2", "R2@a.go:3"},
		},
		{
			name:  "limit spans files",
			files: []*model.UnifiedFileModel{goFile("a.go", "x\n"), goFile("b.go", "x\n")},
			rules: []model.Rule{fakeRule{id: "R1", violations: []model.Violation{viol("R1", "x", 1)}}},
			max:   1,
			want:  []string{"R1@x:1"},
		},
		{
			name:  "a panic counts toward the limit",
			files: []*model.UnifiedFileModel{goFile("a.go", "x\n")},
			rules: []model.Rule{
				fakeRule{id: "BOOM", shouldPanic: true},
				fakeRule{id: "R2", violations: []model.Violation{viol("R2", "a.go", 1)}},
			},
			max:  1,
			want: []string{"BOOM@a.go:1"},
		},
		{
			name:  "suppressed violations do not count toward the limit",
			files: []*model.UnifiedFileModel{goFile("a.go", "// stricture-disable-next-line R1\nx\ny\n")},
			rules: []model.Rule{fakeRule{id: "R1", violations: []model.Violation{viol("R1", "a.go", 2), viol("R1", "a.go", 3)}}},
			max:   1,
			want:  []string{"R1@a.go:3"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RunRules(tc.files, tc.rules, &model.ProjectContext{}, tc.max, 1)
			var gotKeys []string
			for _, v := range got {
				gotKeys = append(gotKeys, v.RuleID+"@"+v.FilePath+":"+itoa(v.StartLine))
			}
			if !reflect.DeepEqual(gotKeys, tc.want) {
				t.Errorf("RunRules = %v, want %v", gotKeys, tc.want)
			}
		})
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := ""
	for ; n > 0; n /= 10 {
		digits = string(rune('0'+n%10)) + digits
	}
	return digits
}

func TestRunRulesKeepsUntrimmedRuleIDButSuppressesByTrimmedID(t *testing.T) {
	file := goFile("a.go", "// stricture-disable-next-line R1\nx\ny\n")
	rule := fakeRule{id: "R1", violations: []model.Violation{
		{RuleID: " R1 ", Severity: "error", FilePath: "a.go", StartLine: 2}, // suppressed via trimmed id
		{RuleID: " R1 ", Severity: "error", FilePath: "a.go", StartLine: 3}, // kept with original spacing
		{RuleID: "", Severity: "error", FilePath: "a.go", StartLine: 0},     // blank id -> rule id, line 0 -> 1
	}}
	got := RunRules([]*model.UnifiedFileModel{file}, []model.Rule{rule}, &model.ProjectContext{}, 0, 1)
	if len(got) != 2 {
		t.Fatalf("got %d violations, want 2: %+v", len(got), got)
	}
	if got[0].RuleID != " R1 " || got[0].StartLine != 3 {
		t.Errorf("first = %+v, want untrimmed RuleID at line 3", got[0])
	}
	if got[1].RuleID != "R1" {
		t.Errorf("blank RuleID should fall back to rule id, got %q", got[1].RuleID)
	}
}

func TestRunRulesUnwrapsRuleWithConfig(t *testing.T) {
	var seen model.RuleConfig
	probe := configProbe{fakeRule: fakeRule{id: "PROBE"}, seen: &seen}
	cfg := model.RuleConfig{Severity: "warn", Options: map[string]interface{}{"max": 3}}
	RunRules([]*model.UnifiedFileModel{goFile("a.go", "x\n")}, []model.Rule{RuleWithConfig{Rule: probe, Config: cfg}}, &model.ProjectContext{}, 0, 1)
	if seen.Severity != "warn" || seen.Options["max"] != 3 {
		t.Errorf("rule saw config %+v, want severity warn and max=3", seen)
	}

	var bare model.RuleConfig
	RunRules([]*model.UnifiedFileModel{goFile("a.go", "x\n")}, []model.Rule{configProbe{fakeRule: fakeRule{id: "PROBE"}, seen: &bare}}, &model.ProjectContext{}, 0, 1)
	if bare.Severity != "error" || bare.Options == nil {
		t.Errorf("unwrapped rule should get default severity and non-nil options, got %+v", bare)
	}
}

func TestRunRulesWrappedRulePanicReportsInnerRuleID(t *testing.T) {
	wrapped := RuleWithConfig{Rule: fakeRule{id: "INNER", shouldPanic: true}, Config: model.RuleConfig{Severity: "warn"}}
	got := RunRules([]*model.UnifiedFileModel{goFile("a.go", "x\n")}, []model.Rule{wrapped}, &model.ProjectContext{}, 0, 1)
	if len(got) != 1 || got[0].RuleID != "INNER" || got[0].Severity != "error" {
		t.Errorf("panic violation = %+v, want one error for INNER", got)
	}
}
