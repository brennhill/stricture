// engine.go — Auto-fix planning and application for fixable rules.
package fix

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/stricture/stricture/internal/model"
)

var (
	renameSuggestionPattern     = regexp.MustCompile(`should be '([^']+)'`)
	expectedDirectoryPattern    = regexp.MustCompile(`should be in '([^']+)'`)
	unsupportedRuleIDsForFixing = map[string]bool{
		"CONV-export-naming": true,
		"TQ-mock-scope":      true,
	}
)

// Operation describes one concrete file-system change.
type Operation struct {
	RuleID      string
	Kind        string // edit|rename
	Path        string
	NewPath     string
	Description string
	Content     []byte // only for edit
}

// Plan builds a list of file operations for fixable violations.
func Plan(violations []model.Violation) ([]Operation, error) {
	ops := make([]Operation, 0)
	seen := map[string]bool{}

	for _, v := range violations {
		if unsupportedRuleIDsForFixing[v.RuleID] {
			continue
		}
		key := v.RuleID + "|" + v.FilePath
		if seen[key] {
			continue
		}
		seen[key] = true

		switch v.RuleID {
		case "CONV-file-header":
			op, ok, err := planFileHeaderFix(v)
			if err != nil {
				return nil, err
			}
			if ok {
				ops = append(ops, op)
			}
		case "CONV-file-naming":
			op, ok := planFileNamingFix(v)
			if ok {
				ops = append(ops, op)
			}
		case "CONV-test-file-location":
			op, ok := planTestLocationFix(v)
			if ok {
				ops = append(ops, op)
			}
		}
	}

	ops = adjustHeaderFixesForRenames(ops)
	return ops, nil
}

func adjustHeaderFixesForRenames(ops []Operation) []Operation {
	renameTargets := map[string]string{}
	for _, op := range ops {
		if op.Kind != "rename" {
			continue
		}
		renameTargets[filepath.Clean(op.Path)] = filepath.Base(op.NewPath)
	}
	if len(renameTargets) == 0 {
		return ops
	}

	adjusted := make([]Operation, 0, len(ops))
	for _, op := range ops {
		if op.Kind == "edit" && op.RuleID == "CONV-file-header" {
			if newBase, ok := renameTargets[filepath.Clean(op.Path)]; ok {
				op.Content = rewriteHeaderFilename(op.Content, newBase)
				op.Description = fmt.Sprintf("Add missing file header to %s", filepath.ToSlash(filepath.Join(filepath.Dir(op.Path), newBase)))
			}
		}
		adjusted = append(adjusted, op)
	}
	return adjusted
}

func rewriteHeaderFilename(content []byte, newBase string) []byte {
	text := string(content)
	firstLineEnd := strings.Index(text, "\n")
	if firstLineEnd < 0 {
		firstLineEnd = len(text)
	}
	firstLine := text[:firstLineEnd]
	if !strings.HasPrefix(firstLine, "// ") {
		return content
	}
	remainder := ""
	if idx := strings.Index(firstLine, " — "); idx >= 0 {
		remainder = firstLine[idx:]
	} else {
		remainder = " — TODO: describe purpose"
	}
	updatedFirst := "// " + newBase + remainder
	if firstLineEnd == len(text) {
		return []byte(updatedFirst)
	}
	return []byte(updatedFirst + text[firstLineEnd:])
}

func planFileHeaderFix(v model.Violation) (Operation, bool, error) {
	data, err := os.ReadFile(v.FilePath)
	if err != nil {
		return Operation{}, false, fmt.Errorf("read %s: %w", v.FilePath, err)
	}

	filename := filepath.Base(v.FilePath)
	expectedPrefix := fmt.Sprintf("// %s — ", filename)
	first := firstNonEmptyLine(string(data))
	if strings.HasPrefix(first, expectedPrefix) {
		return Operation{}, false, nil
	}

	header := fmt.Sprintf("// %s — TODO: describe purpose\n", filename)
	return Operation{
		RuleID:      v.RuleID,
		Kind:        "edit",
		Path:        v.FilePath,
		Description: fmt.Sprintf("Add missing file header to %s", filepath.ToSlash(v.FilePath)),
		Content:     []byte(header + string(data)),
	}, true, nil
}

func planFileNamingFix(v model.Violation) (Operation, bool) {
	match := renameSuggestionPattern.FindStringSubmatch(v.Message)
	if len(match) < 2 {
		return Operation{}, false
	}
	newName := strings.TrimSpace(match[1])
	if newName == "" {
		return Operation{}, false
	}

	oldAbs := v.FilePath
	newAbs := filepath.Join(filepath.Dir(oldAbs), filepath.FromSlash(newName))
	if filepath.Clean(oldAbs) == filepath.Clean(newAbs) {
		return Operation{}, false
	}

	return Operation{
		RuleID:      v.RuleID,
		Kind:        "rename",
		Path:        oldAbs,
		NewPath:     newAbs,
		Description: fmt.Sprintf("Rename %s -> %s", filepath.ToSlash(oldAbs), filepath.ToSlash(newAbs)),
	}, true
}

func planTestLocationFix(v model.Violation) (Operation, bool) {
	match := expectedDirectoryPattern.FindStringSubmatch(v.Message)
	if len(match) < 2 {
		return Operation{}, false
	}
	expectedDir := strings.TrimSpace(match[1])
	if expectedDir == "" {
		return Operation{}, false
	}

	oldAbs := v.FilePath
	newAbs := filepath.Join(filepath.FromSlash(expectedDir), filepath.Base(oldAbs))
	if filepath.Clean(oldAbs) == filepath.Clean(newAbs) {
		return Operation{}, false
	}

	return Operation{
		RuleID:      v.RuleID,
		Kind:        "rename",
		Path:        oldAbs,
		NewPath:     newAbs,
		Description: fmt.Sprintf("Move %s -> %s", filepath.ToSlash(oldAbs), filepath.ToSlash(newAbs)),
	}, true
}

// Apply executes planned operations. It is safe to call with an empty list.
func Apply(ops []Operation) error {
	if len(ops) == 0 {
		return nil
	}

	ordered := append([]Operation(nil), ops...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Kind == ordered[j].Kind {
			return ordered[i].Path < ordered[j].Path
		}
		return ordered[i].Kind == "edit"
	})

	for _, op := range ordered {
		switch op.Kind {
		case "edit":
			if err := os.WriteFile(op.Path, op.Content, 0o644); err != nil {
				return fmt.Errorf("write %s: %w", op.Path, err)
			}
		case "rename":
			if err := os.MkdirAll(filepath.Dir(op.NewPath), 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", filepath.Dir(op.NewPath), err)
			}
			if err := os.Rename(op.Path, op.NewPath); err != nil {
				return fmt.Errorf("rename %s -> %s: %w", op.Path, op.NewPath, err)
			}
		default:
			return fmt.Errorf("unknown fix operation kind %q", op.Kind)
		}
	}
	return nil
}

func firstNonEmptyLine(source string) string {
	for _, line := range strings.Split(source, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		return trimmed
	}
	return ""
}
