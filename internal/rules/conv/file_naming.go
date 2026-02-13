// file_naming.go — CONV-file-naming: Enforce file naming convention.
package conv

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/stricture/stricture/internal/model"
)

// Supported naming conventions.
const (
	StyleKebabCase  = "kebab-case"
	StyleSnakeCase  = "snake_case"
	StyleCamelCase  = "camelCase"
	StylePascalCase = "PascalCase"
)

// Multi-extension suffixes that should be stripped before checking the base name.
// Order matters: longer suffixes must come first so ".test.ts" matches before ".ts".
var multiExtensions = []string{
	".test.ts", ".test.tsx", ".test.js", ".test.jsx",
	".spec.ts", ".spec.tsx", ".spec.js", ".spec.jsx",
	".d.ts", ".d.tsx",
	".test.go", ".spec.go",
	".test.py", ".spec.py",
}

// Convention patterns compiled once.
var conventionPatterns = map[string]*regexp.Regexp{
	StyleKebabCase:  regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`),
	StyleSnakeCase:  regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`),
	StyleCamelCase:  regexp.MustCompile(`^[a-z][a-zA-Z0-9]*$`),
	StylePascalCase: regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`),
}

// Default convention per language.
var languageDefaults = map[string]string{
	"go":         StyleSnakeCase,
	"golang":     StyleSnakeCase,
	"typescript": StyleKebabCase,
	"javascript": StyleKebabCase,
	"tsx":        StyleKebabCase,
	"jsx":        StyleKebabCase,
	"python":     StyleSnakeCase,
	"java":       StylePascalCase,
	"kotlin":     StylePascalCase,
}

// FileNaming enforces file naming conventions (kebab-case, snake_case, etc.).
type FileNaming struct{}

func (r *FileNaming) ID() string                { return "CONV-file-naming" }
func (r *FileNaming) Category() string          { return "conv" }
func (r *FileNaming) Description() string       { return "Enforce file naming convention" }
func (r *FileNaming) DefaultSeverity() string   { return "error" }
func (r *FileNaming) NeedsProjectContext() bool { return false }

func (r *FileNaming) Why() string {
	return "Inconsistent naming makes files hard to find and breaks tooling assumptions."
}

// Check evaluates the file name against the configured or auto-detected naming convention.
func (r *FileNaming) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	convention := resolveConvention(file.Language, config)
	if convention == "" {
		return nil
	}

	pattern, ok := conventionPatterns[convention]
	if !ok {
		return nil
	}

	baseName := extractBaseName(file.Path)
	if baseName == "" {
		return nil
	}

	if pattern.MatchString(baseName) {
		return nil
	}

	severity := config.Severity
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	suggested := convertToConvention(baseName, convention)
	suggestedFull := rebuildFileName(file.Path, suggested)

	return []model.Violation{
		{
			RuleID:   "CONV-file-naming",
			Severity: severity,
			Message: fmt.Sprintf(
				"File name '%s' does not match convention '%s', should be '%s'",
				filepath.Base(file.Path), convention, suggestedFull,
			),
			FilePath:  file.Path,
			StartLine: 1,
			Context: &model.ViolationContext{
				SuggestedFix: fmt.Sprintf("Rename to '%s' using %s.", suggestedFull, convention),
			},
		},
	}
}

// resolveConvention determines which naming convention to use.
// Config option "style" takes precedence over language-based auto-detection.
func resolveConvention(language string, config model.RuleConfig) string {
	if config.Options != nil {
		if style, ok := config.Options["style"]; ok {
			if s, ok := style.(string); ok && s != "" {
				return s
			}
		}
	}
	return languageDefaults[strings.ToLower(language)]
}

// extractBaseName extracts the base name from a file path, stripping all extensions.
// Handles multi-extensions like .test.ts, .spec.ts, .d.ts.
func extractBaseName(path string) string {
	filename := filepath.Base(path)

	// Try multi-extensions first (longest match).
	for _, ext := range multiExtensions {
		if strings.HasSuffix(filename, ext) {
			return strings.TrimSuffix(filename, ext)
		}
	}

	// Fall back to single extension.
	ext := filepath.Ext(filename)
	if ext != "" {
		return strings.TrimSuffix(filename, ext)
	}

	return filename
}

// rebuildFileName reconstructs the filename with the suggested base name,
// preserving the original extension(s).
func rebuildFileName(path string, suggestedBase string) string {
	filename := filepath.Base(path)

	// Try multi-extensions first.
	for _, ext := range multiExtensions {
		if strings.HasSuffix(filename, ext) {
			return suggestedBase + ext
		}
	}

	// Fall back to single extension.
	ext := filepath.Ext(filename)
	return suggestedBase + ext
}

// convertToConvention converts a name to the target naming convention.
func convertToConvention(name string, convention string) string {
	words := splitIntoWords(name)
	if len(words) == 0 {
		return name
	}

	switch convention {
	case StyleKebabCase:
		return joinKebab(words)
	case StyleSnakeCase:
		return joinSnake(words)
	case StyleCamelCase:
		return joinCamel(words)
	case StylePascalCase:
		return joinPascal(words)
	default:
		return name
	}
}

// splitIntoWords splits a name into its component words by detecting boundaries:
// - Hyphens (kebab-case)
// - Underscores (snake_case)
// - Case transitions (camelCase, PascalCase).
func splitIntoWords(name string) []string {
	// First, split on explicit delimiters.
	if strings.Contains(name, "-") {
		return splitAndLower(name, "-")
	}
	if strings.Contains(name, "_") {
		return splitAndLower(name, "_")
	}

	// Split on case transitions for camelCase / PascalCase.
	return splitOnCaseTransitions(name)
}

// splitAndLower splits on a delimiter and lowercases each part.
func splitAndLower(name string, sep string) []string {
	parts := strings.Split(name, sep)
	words := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			words = append(words, strings.ToLower(p))
		}
	}
	return words
}

// splitOnCaseTransitions splits camelCase or PascalCase into words.
// "userService" -> ["user", "service"]
// "HTTPClient" -> ["http", "client"]
// "getAPIKey" -> ["get", "api", "key"].
func splitOnCaseTransitions(name string) []string {
	if name == "" {
		return nil
	}

	runes := []rune(name)
	var words []string
	wordStart := 0

	for i := 1; i < len(runes); i++ {
		if shouldSplitAt(runes, i) {
			word := string(runes[wordStart:i])
			if word != "" {
				words = append(words, strings.ToLower(word))
			}
			wordStart = i
		}
	}

	// Add the last word.
	last := string(runes[wordStart:])
	if last != "" {
		words = append(words, strings.ToLower(last))
	}

	return words
}

// shouldSplitAt determines if a word boundary exists at position i.
func shouldSplitAt(runes []rune, i int) bool {
	curr := runes[i]
	prev := runes[i-1]

	// Lower -> Upper transition: "userS" -> split before S.
	if unicode.IsLower(prev) && unicode.IsUpper(curr) {
		return true
	}

	// Upper -> Upper -> Lower transition: "APIKey" -> split before K.
	// This handles acronyms followed by regular words.
	if i+1 < len(runes) && unicode.IsUpper(prev) && unicode.IsUpper(curr) && unicode.IsLower(runes[i+1]) {
		return true
	}

	// Digit -> Letter or Letter -> Digit transitions.
	if unicode.IsDigit(prev) && unicode.IsLetter(curr) {
		return true
	}
	if unicode.IsLetter(prev) && unicode.IsDigit(curr) {
		return true
	}

	return false
}

// joinKebab joins words with hyphens: ["user", "service"] -> "user-service".
func joinKebab(words []string) string {
	return strings.Join(words, "-")
}

// joinSnake joins words with underscores: ["user", "service"] -> "user_service".
func joinSnake(words []string) string {
	return strings.Join(words, "_")
}

// joinCamel joins words in camelCase: ["user", "service"] -> "userService".
func joinCamel(words []string) string {
	if len(words) == 0 {
		return ""
	}
	result := strings.ToLower(words[0])
	for _, w := range words[1:] {
		result += capitalize(w)
	}
	return result
}

// joinPascal joins words in PascalCase: ["user", "service"] -> "UserService".
func joinPascal(words []string) string {
	var result string
	for _, w := range words {
		result += capitalize(w)
	}
	return result
}

// capitalize uppercases the first rune in a string.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
