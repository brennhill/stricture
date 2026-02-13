// export_naming_test.go - Tests for CONV-export-naming rule.
package conv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stricture/stricture/internal/model"
)

func TestExportNaming_Metadata(t *testing.T) {
	rule := &ExportNaming{}
	assert.Equal(t, "CONV-export-naming", rule.ID())
	assert.Equal(t, "conv", rule.Category())
	assert.Equal(t, "error", rule.DefaultSeverity())
	assert.False(t, rule.NeedsProjectContext())
	assert.NotEmpty(t, rule.Description())
	assert.NotEmpty(t, rule.Why())
}

func TestExportNaming_TSFunctionMustBeCamelCase(t *testing.T) {
	rule := &ExportNaming{}
	file := &model.UnifiedFileModel{
		Path:     "src/services/user.ts",
		Language: "typescript",
		Source:   []byte("export function CreateUser() {}\n"),
	}

	violations := rule.Check(file, nil, model.RuleConfig{})
	require.Len(t, violations, 1)
	assert.Equal(t, "CONV-export-naming", violations[0].RuleID)
	assert.Contains(t, violations[0].Message, "does not follow convention")
	assert.Contains(t, violations[0].Message, "camelCase")
}

func TestExportNaming_TSConstantMustBeUpperSnakeCase(t *testing.T) {
	rule := &ExportNaming{}
	file := &model.UnifiedFileModel{
		Path:     "src/constants.ts",
		Language: "typescript",
		Source:   []byte("export const maxRetries = 3\n"),
	}

	violations := rule.Check(file, nil, model.RuleConfig{})
	require.Len(t, violations, 1)
	assert.Contains(t, violations[0].Message, "UPPER_SNAKE_CASE")
	assert.Contains(t, violations[0].Message, "MAX_RETRIES")
}

func TestExportNaming_AllowsConfiguredOverrides(t *testing.T) {
	rule := &ExportNaming{}
	file := &model.UnifiedFileModel{
		Path:     "src/services/user.ts",
		Language: "typescript",
		Source:   []byte("export function CreateUser() {}\n"),
	}

	cfg := model.RuleConfig{Options: map[string]interface{}{
		"typescript": map[string]interface{}{
			"exportedFunctions": "PascalCase",
		},
	}}

	violations := rule.Check(file, nil, cfg)
	assert.Empty(t, violations)
}

func TestExportNaming_GoAcronymCasing(t *testing.T) {
	rule := &ExportNaming{}
	file := &model.UnifiedFileModel{
		Path:     "internal/service/user.go",
		Language: "go",
		Source:   []byte("func CreateUSER() {}\n"),
	}

	violations := rule.Check(file, nil, model.RuleConfig{})
	require.Len(t, violations, 1)
	assert.Contains(t, violations[0].Message, "PascalCase")
	assert.Contains(t, violations[0].Message, "CreateUser")
}

func TestExportNaming_PassesWhenConformant(t *testing.T) {
	rule := &ExportNaming{}
	file := &model.UnifiedFileModel{
		Path:     "src/services/user.ts",
		Language: "typescript",
		Source:   []byte("export function createUser() {}\nexport class UserService {}\nexport const MAX_RETRIES = 3\nexport type UserInput = { id: string }\n"),
	}

	violations := rule.Check(file, nil, model.RuleConfig{})
	assert.Empty(t, violations)
}
