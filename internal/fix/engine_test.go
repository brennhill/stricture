package fix

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stricture/stricture/internal/model"
)

func TestPlanFileHeaderFix(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "user-service.ts")
	if err := os.WriteFile(target, []byte("export const value = 1;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	violations := []model.Violation{
		{
			RuleID:   "CONV-file-header",
			FilePath: target,
			Message:  "File missing header comment",
		},
	}

	ops, err := Plan(violations)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("ops len = %d, want 1", len(ops))
	}
	if ops[0].Kind != "edit" {
		t.Fatalf("op kind = %q, want edit", ops[0].Kind)
	}
	if err := Apply(ops); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if !strings.HasPrefix(string(after), "// user-service.ts — ") {
		t.Fatalf("header prefix missing after apply, got %q", string(after))
	}
}

func TestPlanFileNamingFix(t *testing.T) {
	v := model.Violation{
		RuleID:   "CONV-file-naming",
		FilePath: "/tmp/UserService.ts",
		Message:  "File name 'UserService.ts' does not match convention 'kebab-case', should be 'user-service.ts'",
	}

	ops, err := Plan([]model.Violation{v})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("ops len = %d, want 1", len(ops))
	}
	if ops[0].Kind != "rename" {
		t.Fatalf("op kind = %q, want rename", ops[0].Kind)
	}
	if filepath.Base(ops[0].NewPath) != "user-service.ts" {
		t.Fatalf("rename target = %q", ops[0].NewPath)
	}
}

func TestApplyRename(t *testing.T) {
	tmp := t.TempDir()
	oldPath := filepath.Join(tmp, "UserService.ts")
	newPath := filepath.Join(tmp, "user-service.ts")
	if err := os.WriteFile(oldPath, []byte("x\n"), 0o644); err != nil {
		t.Fatalf("write oldPath: %v", err)
	}

	if err := Apply([]Operation{{
		RuleID:      "CONV-file-naming",
		Kind:        "rename",
		Path:        oldPath,
		NewPath:     newPath,
		Description: "rename",
	}}); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if _, err := os.Stat(newPath); err != nil {
		t.Fatalf("new path missing: %v", err)
	}
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("old path should not exist after rename")
	}
}

func TestPlanSkipsUnsupportedAndDuplicateViolations(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "user-service.ts")
	if err := os.WriteFile(target, []byte("export const value = 1;\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	violations := []model.Violation{
		{RuleID: "CONV-export-naming", FilePath: target, Message: "unsupported"},
		{RuleID: "CONV-file-header", FilePath: target, Message: "missing header"},
		{RuleID: "CONV-file-header", FilePath: target, Message: "duplicate missing header"},
	}
	ops, err := Plan(violations)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("ops len = %d, want 1", len(ops))
	}
	if ops[0].RuleID != "CONV-file-header" {
		t.Fatalf("rule id = %q, want CONV-file-header", ops[0].RuleID)
	}
}

func TestPlanHeaderFixNoOpWhenAlreadyPresent(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "user-service.ts")
	content := "// user-service.ts — handles user routes\nexport const value = 1;\n"
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}

	ops, err := Plan([]model.Violation{{
		RuleID:   "CONV-file-header",
		FilePath: target,
		Message:  "missing header",
	}})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(ops) != 0 {
		t.Fatalf("ops len = %d, want 0", len(ops))
	}
}

func TestPlanHeaderFixReadError(t *testing.T) {
	_, err := Plan([]model.Violation{{
		RuleID:   "CONV-file-header",
		FilePath: filepath.Join(t.TempDir(), "missing.ts"),
		Message:  "missing header",
	}})
	if err == nil {
		t.Fatalf("expected read error")
	}
}

func TestPlanTestLocationFix(t *testing.T) {
	v := model.Violation{
		RuleID:   "CONV-test-file-location",
		FilePath: filepath.FromSlash("services/user/user_test.go"),
		Message:  "Test file should be in 'tests/unit/user'",
	}

	ops, err := Plan([]model.Violation{v})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(ops) != 1 {
		t.Fatalf("ops len = %d, want 1", len(ops))
	}
	if ops[0].Kind != "rename" {
		t.Fatalf("op kind = %q, want rename", ops[0].Kind)
	}
	want := filepath.Join(filepath.FromSlash("tests/unit/user"), "user_test.go")
	if filepath.Clean(ops[0].NewPath) != filepath.Clean(want) {
		t.Fatalf("new path = %q, want %q", ops[0].NewPath, want)
	}
}

func TestPlanAdjustsHeaderContentWhenFileIsRenamed(t *testing.T) {
	tmp := t.TempDir()
	oldPath := filepath.Join(tmp, "UserService.ts")
	if err := os.WriteFile(oldPath, []byte("export const value = 1;\n"), 0o644); err != nil {
		t.Fatalf("write oldPath: %v", err)
	}

	ops, err := Plan([]model.Violation{
		{
			RuleID:   "CONV-file-header",
			FilePath: oldPath,
			Message:  "missing header",
		},
		{
			RuleID:   "CONV-file-naming",
			FilePath: oldPath,
			Message:  "File name 'UserService.ts' should be 'user-service.ts'",
		},
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if len(ops) != 2 {
		t.Fatalf("ops len = %d, want 2", len(ops))
	}

	var edit Operation
	for _, op := range ops {
		if op.Kind == "edit" {
			edit = op
		}
	}
	if edit.Kind != "edit" {
		t.Fatalf("missing edit op in plan")
	}
	if !strings.HasPrefix(string(edit.Content), "// user-service.ts — ") {
		t.Fatalf("header not rewritten for rename: %q", string(edit.Content))
	}
	if !strings.Contains(edit.Description, "user-service.ts") {
		t.Fatalf("description not rewritten for rename: %q", edit.Description)
	}
}

func TestRewriteHeaderFilename(t *testing.T) {
	updated := rewriteHeaderFilename([]byte("// old.ts — data access\nconst x = 1;\n"), "new.ts")
	if !strings.HasPrefix(string(updated), "// new.ts — data access") {
		t.Fatalf("unexpected rewritten header: %q", string(updated))
	}

	updatedNoDash := rewriteHeaderFilename([]byte("// old.ts\nconst x = 1;\n"), "new.ts")
	if !strings.HasPrefix(string(updatedNoDash), "// new.ts — TODO: describe purpose\n") {
		t.Fatalf("missing fallback suffix in rewritten header: %q", string(updatedNoDash))
	}

	untouched := []byte("package main\n")
	if got := rewriteHeaderFilename(untouched, "main.go"); string(got) != string(untouched) {
		t.Fatalf("non-comment first line should be unchanged: %q", string(got))
	}
}

func TestApplyEditThenRenameCreatesDirectories(t *testing.T) {
	tmp := t.TempDir()
	oldPath := filepath.Join(tmp, "UserService.ts")
	newPath := filepath.Join(tmp, "nested", "services", "user-service.ts")
	if err := os.WriteFile(oldPath, []byte("before\n"), 0o644); err != nil {
		t.Fatalf("write oldPath: %v", err)
	}

	ops := []Operation{
		{
			Kind:    "rename",
			Path:    oldPath,
			NewPath: newPath,
		},
		{
			Kind:    "edit",
			Path:    oldPath,
			Content: []byte("after\n"),
		},
	}
	if err := Apply(ops); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	after, err := os.ReadFile(newPath)
	if err != nil {
		t.Fatalf("read newPath: %v", err)
	}
	if string(after) != "after\n" {
		t.Fatalf("renamed file contents = %q, want %q", string(after), "after\n")
	}
}

func TestApplyUnknownOperationKind(t *testing.T) {
	err := Apply([]Operation{{Kind: "noop", Path: "x"}})
	if err == nil {
		t.Fatalf("expected unknown operation error")
	}
}

func TestFirstNonEmptyLine(t *testing.T) {
	if got := firstNonEmptyLine("\n  \n\t\nvalue\n"); got != "value" {
		t.Fatalf("firstNonEmptyLine = %q, want value", got)
	}
	if got := firstNonEmptyLine("\n  \n\t\n"); got != "" {
		t.Fatalf("firstNonEmptyLine = %q, want empty", got)
	}
}
