// fix_apply.go — Helpers for planning, backing up, and reporting auto-fixes.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/stricture/stricture/internal/fix"
)

func rewritePathsAfterFix(paths []string, ops []fix.Operation) []string {
	renames := map[string]string{}
	for _, op := range ops {
		if op.Kind != "rename" {
			continue
		}
		renames[filepath.Clean(op.Path)] = filepath.Clean(op.NewPath)
	}
	if len(renames) == 0 {
		return append([]string(nil), paths...)
	}

	rewritten := make([]string, 0, len(paths))
	for _, pathValue := range paths {
		clean := filepath.Clean(pathValue)
		if newPath, ok := renames[clean]; ok {
			rewritten = append(rewritten, newPath)
			continue
		}
		rewritten = append(rewritten, pathValue)
	}
	return rewritten
}

func renderFixOperations(ops []fix.Operation) []map[string]string {
	out := make([]map[string]string, 0, len(ops))
	for _, op := range ops {
		entry := map[string]string{
			"ruleId":      op.RuleID,
			"kind":        op.Kind,
			"path":        filepath.ToSlash(op.Path),
			"description": op.Description,
		}
		if op.NewPath != "" {
			entry["newPath"] = filepath.ToSlash(op.NewPath)
		}
		out = append(out, entry)
	}
	return out
}

func formatFixSummary(ops []fix.Operation, dryRun bool) string {
	var out strings.Builder
	mode := "apply"
	if dryRun {
		mode = "dry-run"
	}
	fmt.Fprintf(&out, "Fixes: %d operation(s) (%s)\n", len(ops), mode)
	for _, op := range ops {
		fmt.Fprintf(&out, "  - [%s] %s\n", op.RuleID, op.Description)
	}
	return out.String()
}

func writeFixBackups(ops []fix.Operation) error {
	paths := map[string]bool{}
	for _, op := range ops {
		switch op.Kind {
		case "edit", "rename":
			if strings.TrimSpace(op.Path) != "" {
				paths[filepath.Clean(op.Path)] = true
			}
		}
	}

	ordered := make([]string, 0, len(paths))
	for pathValue := range paths {
		ordered = append(ordered, pathValue)
	}
	sort.Strings(ordered)

	for _, pathValue := range ordered {
		data, err := os.ReadFile(pathValue)
		if err != nil {
			return fmt.Errorf("read %s for backup: %w", pathValue, err)
		}

		backupPath := pathValue + ".bak"
		if _, err := os.Stat(backupPath); err == nil {
			return fmt.Errorf("backup already exists: %s", backupPath)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("check backup %s: %w", backupPath, err)
		}

		//nolint:gosec // G703: pathValue is a file already selected for fixing, so the backup has the same trust as the fix write itself.
		if err := os.WriteFile(backupPath, data, 0o644); err != nil {
			return fmt.Errorf("write backup %s: %w", backupPath, err)
		}
	}
	return nil
}
