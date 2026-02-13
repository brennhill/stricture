// error_format.go - CONV-error-format: Enforce structured error messages.
package conv

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/stricture/stricture/internal/model"
)

// ErrorFormat enforces a consistent error message shape.
type ErrorFormat struct{}

func (r *ErrorFormat) ID() string                { return "CONV-error-format" }
func (r *ErrorFormat) Category() string          { return "conv" }
func (r *ErrorFormat) Description() string       { return "Enforce consistent error message format" }
func (r *ErrorFormat) DefaultSeverity() string   { return "error" }
func (r *ErrorFormat) NeedsProjectContext() bool { return false }
func (r *ErrorFormat) Why() string {
	return "Consistent error format makes logs searchable and tells users how to recover."
}

var (
	goFmtErrorRe = regexp.MustCompile("fmt\\.Errorf\\(\\s*(\"(?:\\\\.|[^\"\\\\])*\"|`[^`]*`)\\s*(?:,|\\))")
	goNewErrRe   = regexp.MustCompile("errors\\.New\\(\\s*(\"(?:\\\\.|[^\"\\\\])*\"|`[^`]*`)\\s*\\)")
	jsErrorRe    = regexp.MustCompile("new\\s+([A-Za-z_][A-Za-z0-9_]*)\\s*\\(\\s*(\"(?:\\\\.|[^\"\\\\])*\"|'(?:\\\\.|[^'\\\\])*'|`[^`]*`)\\s*\\)")

	operationTokenRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`)
	upperSnakeRe     = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)
)

var builtInJSErrors = map[string]bool{
	"TypeError":      true,
	"RangeError":     true,
	"ReferenceError": true,
	"SyntaxError":    true,
	"EvalError":      true,
	"URIError":       true,
	"AggregateError": true,
}

var defaultApplyTargets = map[string]bool{
	"fmt.Errorf":        true,
	"errors.New":        true,
	"new Error":         true,
	"throw new .*Error": true,
}

// Check validates error message literals against a structured format:
// {OPERATION}: {ROOT_CAUSE}. {RECOVERY_ACTION}.
func (r *ErrorFormat) Check(file *model.UnifiedFileModel, _ *model.ProjectContext, config model.RuleConfig) []model.Violation {
	if file == nil || len(file.Source) == 0 {
		return nil
	}

	severity := config.Severity
	if severity == "" {
		severity = r.DefaultSeverity()
	}

	minSegments := resolveMinSegments(config)
	applyTargets := resolveApplyTargets(config)
	lines := strings.Split(string(file.Source), "\n")
	var violations []model.Violation

	for i, line := range lines {
		lineNo := i + 1

		if applyTargets["fmt.Errorf"] {
			for _, match := range goFmtErrorRe.FindAllStringSubmatch(line, -1) {
				literal := match[1]
				message, isTemplate := literalMessage(literal)
				if !isStructuredError(message, minSegments, isTemplate) {
					violations = append(violations, r.newViolation(file.Path, lineNo, severity, message))
				}
			}
		}

		if applyTargets["errors.New"] {
			for _, match := range goNewErrRe.FindAllStringSubmatch(line, -1) {
				literal := match[1]
				message, isTemplate := literalMessage(literal)
				if !isStructuredError(message, minSegments, isTemplate) {
					violations = append(violations, r.newViolation(file.Path, lineNo, severity, message))
				}
			}
		}

		for _, match := range jsErrorRe.FindAllStringSubmatch(line, -1) {
			errorType := match[1]
			literal := match[2]

			if !shouldCheckJSError(errorType, applyTargets) {
				continue
			}

			message, isTemplate := literalMessage(literal)
			if !isStructuredError(message, minSegments, isTemplate) {
				violations = append(violations, r.newViolation(file.Path, lineNo, severity, message))
			}
		}
	}

	return violations
}

func resolveMinSegments(config model.RuleConfig) int {
	const defaultMinSegments = 2

	if config.Options == nil {
		return defaultMinSegments
	}

	value, ok := config.Options["minSegments"]
	if !ok {
		return defaultMinSegments
	}

	switch v := value.(type) {
	case int:
		if v >= 2 && v <= 3 {
			return v
		}
	case int64:
		if v >= 2 && v <= 3 {
			return int(v)
		}
	case float64:
		n := int(v)
		if float64(n) == v && n >= 2 && n <= 3 {
			return n
		}
	}

	return defaultMinSegments
}

func resolveApplyTargets(config model.RuleConfig) map[string]bool {
	targets := map[string]bool{}
	for key, value := range defaultApplyTargets {
		targets[key] = value
	}

	if config.Options == nil {
		return targets
	}

	value, ok := config.Options["applyTo"]
	if !ok {
		return targets
	}

	custom := map[string]bool{}
	switch raw := value.(type) {
	case []string:
		for _, item := range raw {
			item = strings.TrimSpace(item)
			if item != "" {
				custom[item] = true
			}
		}
	case []interface{}:
		for _, item := range raw {
			s, ok := item.(string)
			if !ok {
				continue
			}
			s = strings.TrimSpace(s)
			if s != "" {
				custom[s] = true
			}
		}
	default:
		return targets
	}

	if len(custom) == 0 {
		return targets
	}
	return custom
}

func shouldCheckJSError(errorType string, targets map[string]bool) bool {
	if builtInJSErrors[errorType] {
		return false
	}

	if errorType == "Error" {
		return targets["new Error"]
	}

	if strings.HasSuffix(errorType, "Error") {
		return targets["throw new .*Error"]
	}

	return false
}

func literalMessage(literal string) (message string, isTemplate bool) {
	if len(literal) < 2 {
		return literal, false
	}

	first := literal[0]
	last := literal[len(literal)-1]
	if first != last {
		return literal, false
	}

	switch first {
	case '`':
		return literal[1 : len(literal)-1], true
	case '\'', '"':
		return literal[1 : len(literal)-1], false
	default:
		return literal, false
	}
}

func isStructuredError(message string, minSegments int, isTemplate bool) bool {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return false
	}

	// Dynamic template literals are hard to evaluate statically.
	// If they contain the segment separator, treat as potentially valid.
	if isTemplate && strings.Contains(trimmed, ":") {
		return true
	}

	parts := strings.SplitN(trimmed, ":", 2)
	if len(parts) != 2 {
		return false
	}

	operation := strings.TrimSpace(parts[0])
	body := strings.TrimSpace(parts[1])
	if !validOperation(operation) || body == "" {
		return false
	}

	root, recovery := splitBody(body)
	if root == "" {
		return false
	}

	segments := 2
	if recovery != "" {
		segments = 3
	}

	return segments >= minSegments
}

func validOperation(operation string) bool {
	if operation == "" || strings.ContainsAny(operation, " \t\n\r") {
		return false
	}
	if operationTokenRe.MatchString(operation) {
		return true
	}
	return upperSnakeRe.MatchString(operation)
}

func splitBody(body string) (root string, recovery string) {
	parts := strings.SplitN(body, ".", 2)
	root = strings.TrimSpace(parts[0])
	if len(parts) < 2 {
		return root, ""
	}
	recovery = strings.TrimSpace(parts[1])
	return root, recovery
}

func (r *ErrorFormat) newViolation(filePath string, line int, severity string, message string) model.Violation {
	suggested := suggestedMessage(filePath)
	return model.Violation{
		RuleID:    r.ID(),
		Severity:  severity,
		Message:   fmt.Sprintf("Error message '%s' does not follow format: {OPERATION}: {ROOT_CAUSE}. {RECOVERY_ACTION}", message),
		FilePath:  filePath,
		StartLine: line,
		Context: &model.ViolationContext{
			SuggestedFix: fmt.Sprintf("Rewrite as: '%s'.", suggested),
		},
	}
}

func suggestedMessage(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	parts := strings.FieldsFunc(base, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	if len(parts) == 0 {
		return "Operation: describe root cause. Describe recovery action."
	}

	var operation strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		lower := strings.ToLower(part)
		operation.WriteString(strings.ToUpper(lower[:1]))
		if len(lower) > 1 {
			operation.WriteString(lower[1:])
		}
	}

	op := operation.String()
	if op == "" {
		op = "Operation"
	}

	return fmt.Sprintf("%s: describe root cause. Describe recovery action.", op)
}
