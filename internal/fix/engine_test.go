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
