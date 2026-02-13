// error_format_test.go - Tests for CONV-error-format rule.
package conv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stricture/stricture/internal/model"
)

func TestErrorFormat_InterfaceCompliance(t *testing.T) {
	rule := &ErrorFormat{}
	var _ model.Rule = rule

	assert.Equal(t, "CONV-error-format", rule.ID())
	assert.Equal(t, "conv", rule.Category())
	assert.Equal(t, "error", rule.DefaultSeverity())
	assert.Equal(t, "Enforce consistent error message format", rule.Description())
	assert.False(t, rule.NeedsProjectContext())
	assert.Equal(t,
		"Consistent error format makes logs searchable and tells users how to recover.",
		rule.Why(),
	)
}

func TestErrorFormat_TruePositives(t *testing.T) {
	rule := &ErrorFormat{}

	tests := []struct {
		name    string
		lang    string
		source  string
		options map[string]interface{}
	}{
		{
			name:   "TP-EF-01 unstructured new Error",
			lang:   "typescript",
			source: `throw new Error("something went wrong");`,
		},
		{
			name:   "TP-EF-02 single segment",
			lang:   "typescript",
			source: `throw new Error("failed");`,
		},
		{
			name:   "TP-EF-03 go fmt.Errorf without structured format",
			lang:   "go",
			source: `return fmt.Errorf("error creating user")`,
		},
		{
			name:    "TP-EF-05 requires recovery action when minSegments=3",
			lang:    "typescript",
			source:  `throw new Error("CreateUser: email already exists");`,
			options: map[string]interface{}{"minSegments": 3},
		},
		{
			name:   "TP-EF-06 bad operation token",
			lang:   "go",
			source: `return fmt.Errorf("bad request: %v", err)`,
		},
		{
			name:   "TP-EF-07 custom Error class with bad message",
			lang:   "typescript",
			source: `throw new ValidationError("oops");`,
		},
		{
			name:   "TP-EF-08 callback new Error",
			lang:   "typescript",
			source: `callback(new Error("it broke"));`,
		},
		{
			name: "TP-EF-09 mixed errors reports only invalid",
			lang: "typescript",
			source: `
if (!name) throw new Error("missing name");
if (!email) throw new Error("CreateUser: email required. Provide a valid email.");
`,
		},
		{
			name:   "TP-EF-10 go errors.New without format",
			lang:   "go",
			source: `var ErrBad = errors.New("bad thing happened")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := model.RuleConfig{Options: tt.options}
			file := &model.UnifiedFileModel{
				Path:     testPath(tt.lang),
				Language: tt.lang,
				Source:   []byte(tt.source),
			}

			violations := rule.Check(file, nil, cfg)
			require.NotEmpty(t, violations, "expected at least one violation for %s", tt.name)
			assert.Equal(t, "CONV-error-format", violations[0].RuleID)
			assert.Equal(t, "error", violations[0].Severity)
			assert.Contains(t, violations[0].Message, "Error message")
			assert.NotEmpty(t, violations[0].Context)
			assert.NotEmpty(t, violations[0].Context.SuggestedFix)

			if tt.name == "TP-EF-09 mixed errors reports only invalid" {
				assert.Len(t, violations, 1, "only the invalid message should be flagged")
			}
		})
	}
}

func TestErrorFormat_TrueNegatives(t *testing.T) {
	rule := &ErrorFormat{}

	tests := []struct {
		name    string
		lang    string
		source  string
		options map[string]interface{}
	}{
		{
			name:   "TN-EF-01 valid structured TS error",
			lang:   "typescript",
			source: `throw new Error("CreateUser: email already exists. Use a different email address.");`,
		},
		{
			name:   "TN-EF-02 valid structured go error",
			lang:   "go",
			source: `return fmt.Errorf("FetchUser: user not found. Verify the user ID is correct.")`,
		},
		{
			name:    "TN-EF-03 minSegments=2 allows operation+cause",
			lang:    "typescript",
			source:  `throw new Error("CreateUser: email already exists.");`,
			options: map[string]interface{}{"minSegments": 2},
		},
		{
			name:   "TN-EF-04 go wrapping accepted",
			lang:   "go",
			source: `return fmt.Errorf("CreateUser: %w", err)`,
		},
		{
			name:   "TN-EF-05 applyTo excludes new Error",
			lang:   "typescript",
			source: `throw new Error("bad thing");`,
			options: map[string]interface{}{
				"applyTo": []string{"fmt.Errorf"},
			},
		},
		{
			name:   "FP-EF-01 builtin TypeError is ignored",
			lang:   "typescript",
			source: `throw new TypeError("Cannot read property 'x' of undefined");`,
		},
		{
			name:   "FP-EF-02 template literal with colon is treated as potentially valid",
			lang:   "typescript",
			source: "throw new Error(`${operation}: ${reason}`);",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := model.RuleConfig{Options: tt.options}
			file := &model.UnifiedFileModel{
				Path:     testPath(tt.lang),
				Language: tt.lang,
				Source:   []byte(tt.source),
			}
			violations := rule.Check(file, nil, cfg)
			assert.Empty(t, violations)
		})
	}
}

func TestErrorFormat_SeverityOverride(t *testing.T) {
	rule := &ErrorFormat{}
	file := &model.UnifiedFileModel{
		Path:     testPath("typescript"),
		Language: "typescript",
		Source:   []byte(`throw new Error("failed");`),
	}

	violations := rule.Check(file, nil, model.RuleConfig{Severity: "warn"})
	require.Len(t, violations, 1)
	assert.Equal(t, "warn", violations[0].Severity)
}

func testPath(language string) string {
	switch language {
	case "go":
		return "/project/internal/example.go"
	case "typescript":
		return "/project/src/example.ts"
	default:
		return "/project/src/example.txt"
	}
}
