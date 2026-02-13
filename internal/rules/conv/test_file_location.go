// test_location.go - CONV-test-file-location: Enforce test file placement conventions.
package conv

import (
	"fmt"
	"path"
	"strings"

	"github.com/stricture/stricture/internal/model"
)

const (
	testStrategyColocated = "colocated"
	testStrategyMirrored  = "mirrored"
	testStrategySubfolder = "subfolder"
)

var defaultTestSuffixes = map[string][]string{
	"go":         {"_test.go"},
	"typescript": {".test.ts", ".spec.ts", ".test.tsx", ".spec.tsx"},
	"javascript": {".test.js", ".spec.js", ".test.jsx", ".spec.jsx"},
	"python":     {"_test.py"},
	"java":       {"Test.java"},
}

// TestFileLocation enforces test file path conventions.
type TestFileLocation struct{}

func (r *TestFileLocation) ID() string       { return "CONV-test-file-location" }
func (r *TestFileLocation) Category() string { return "conv" }
func (r *TestFileLocation) Description() string {
	return "Enforce where test files live relative to source files"
}
func (r *TestFileLocation) DefaultSeverity() string   { return "error" }
func (r *TestFileLocation) NeedsProjectContext() bool { return false }
func (r *TestFileLocation) Why() string {
	return "Scattered test files make test discovery and coverage analysis unreliable."
}

func (r *TestFileLocation) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	if file == nil {
		return nil
	}

	strategy := resolveTestStrategy(config.Options)
	suffixes := resolveTestSuffixes(file.Language, config.Options)
	if len(suffixes) == 0 {
		return nil
	}

	pathValue := filepathToSlash(file.Path)
	if !looksLikeTestFile(pathValue, file.Language, suffixes, file.IsTestFile) {
		return nil
	}

	severity := config.Severity
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	if !hasAnySuffix(pathValue, suffixes) {
		expected := pathValue + suffixes[0]
		return []model.Violation{r.newViolation(file.Path, severity, path.Dir(pathValue), path.Dir(expected), 1)}
	}

	if locationMatchesStrategy(pathValue, strategy) {
		return nil
	}

	expectedPath := expectedTestPath(pathValue, strategy)
	if expectedPath == "" {
		expectedPath = pathValue
	}

	return []model.Violation{r.newViolation(file.Path, severity, path.Dir(pathValue), path.Dir(expectedPath), 1)}
}

func (r *TestFileLocation) newViolation(filePath string, severity string, actual string, expected string, line int) model.Violation {
	return model.Violation{
		RuleID:    r.ID(),
		Severity:  severity,
		Message:   fmt.Sprintf("Test file '%s' is in '%s', should be in '%s' per convention", path.Base(filepathToSlash(filePath)), actual, expected),
		FilePath:  filePath,
		StartLine: line,
		Context: &model.ViolationContext{
			SuggestedFix: fmt.Sprintf("Move test file to '%s'.", expected),
		},
	}
}

func resolveTestStrategy(options map[string]interface{}) string {
	if options != nil {
		if raw, ok := options["strategy"]; ok {
			if s, ok := raw.(string); ok {
				s = strings.ToLower(strings.TrimSpace(s))
				switch s {
				case testStrategyColocated, testStrategyMirrored, testStrategySubfolder:
					return s
				}
			}
		}
	}
	return testStrategyColocated
}

func resolveTestSuffixes(language string, options map[string]interface{}) []string {
	lang := normalizeLanguage(language)
	suffixes := append([]string{}, defaultTestSuffixes[lang]...)

	if options == nil {
		return suffixes
	}

	raw, ok := options["suffixes"]
	if !ok {
		return suffixes
	}
	m, ok := toStringMap(raw)
	if !ok {
		return suffixes
	}

	custom, ok := m[lang]
	if !ok {
		return suffixes
	}
	parsed := toStringSlice(custom)
	if len(parsed) == 0 {
		return suffixes
	}
	return parsed
}

func normalizeLanguage(language string) string {
	switch strings.ToLower(strings.TrimSpace(language)) {
	case "go", "golang":
		return "go"
	case "typescript", "tsx":
		return "typescript"
	case "javascript", "jsx":
		return "javascript"
	case "python":
		return "python"
	case "java":
		return "java"
	default:
		return strings.ToLower(strings.TrimSpace(language))
	}
}

func looksLikeTestFile(pathValue string, language string, suffixes []string, isTestFlag bool) bool {
	if isTestFlag {
		return true
	}
	if hasAnySuffix(pathValue, suffixes) {
		return true
	}
	lang := normalizeLanguage(language)
	switch lang {
	case "go":
		return strings.HasSuffix(pathValue, "_test.go")
	case "typescript", "javascript":
		return strings.Contains(pathValue, ".test.") || strings.Contains(pathValue, ".spec.")
	case "python":
		base := path.Base(pathValue)
		return strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.py")
	case "java":
		return strings.HasSuffix(pathValue, "Test.java")
	default:
		return false
	}
}

func hasAnySuffix(pathValue string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(pathValue, suffix) {
			return true
		}
	}
	return false
}

func locationMatchesStrategy(pathValue string, strategy string) bool {
	hasTestsRoot := strings.HasPrefix(pathValue, "tests/") || strings.HasPrefix(pathValue, "test/")
	hasSubfolder := strings.Contains(pathValue, "/__tests__/")

	switch strategy {
	case testStrategyColocated:
		return !hasTestsRoot && !hasSubfolder
	case testStrategyMirrored:
		return hasTestsRoot
	case testStrategySubfolder:
		return hasSubfolder
	default:
		return true
	}
}

func expectedTestPath(pathValue string, strategy string) string {
	switch strategy {
	case testStrategyColocated:
		if strings.HasPrefix(pathValue, "tests/") {
			return "src/" + strings.TrimPrefix(pathValue, "tests/")
		}
		if strings.HasPrefix(pathValue, "test/") {
			return "src/" + strings.TrimPrefix(pathValue, "test/")
		}
		return strings.Replace(pathValue, "/__tests__/", "/", 1)
	case testStrategyMirrored:
		trimmed := strings.TrimPrefix(pathValue, "src/")
		trimmed = strings.Replace(trimmed, "/__tests__/", "/", 1)
		if strings.HasPrefix(pathValue, "tests/") {
			return pathValue
		}
		if strings.HasPrefix(pathValue, "test/") {
			return "tests/" + strings.TrimPrefix(pathValue, "test/")
		}
		return "tests/" + trimmed
	case testStrategySubfolder:
		if strings.Contains(pathValue, "/__tests__/") {
			return pathValue
		}
		if strings.HasPrefix(pathValue, "tests/") {
			trimmed := strings.TrimPrefix(pathValue, "tests/")
			dir, file := path.Split(trimmed)
			dir = strings.TrimSuffix(dir, "/")
			if dir == "" {
				return "src/__tests__/" + file
			}
			return "src/" + dir + "/__tests__/" + file
		}
		dir, file := path.Split(pathValue)
		dir = strings.TrimSuffix(dir, "/")
		if dir == "" {
			return "__tests__/" + file
		}
		return dir + "/__tests__/" + file
	default:
		return pathValue
	}
}

func filepathToSlash(value string) string {
	value = strings.ReplaceAll(value, "\\", "/")
	value = strings.TrimPrefix(value, "./")
	return value
}
