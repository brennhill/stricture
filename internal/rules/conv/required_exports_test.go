// required_exports_test.go - Tests for CONV-required-exports rule.
package conv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stricture/stricture/internal/model"
)

func TestRequiredExports_Metadata(t *testing.T) {
	rule := &RequiredExports{}
	assert.Equal(t, "CONV-required-exports", rule.ID())
	assert.Equal(t, "conv", rule.Category())
	assert.Equal(t, "error", rule.DefaultSeverity())
	assert.NotEmpty(t, rule.Description())
	assert.NotEmpty(t, rule.Why())
}

func TestRequiredExports_MissingDefaultExport(t *testing.T) {
	rule := &RequiredExports{}
	file := &model.UnifiedFileModel{
		Path:     "src/features/auth/index.ts",
		Language: "typescript",
		Source:   []byte("export const authService = {}\n"),
	}
	cfg := model.RuleConfig{Options: map[string]interface{}{
		"patterns": map[string]interface{}{
			"src/features/*/index.ts": map[string]interface{}{
				"required": []interface{}{"default"},
			},
		},
	}}

	violations := rule.Check(file, nil, cfg)
	require.Len(t, violations, 1)
	assert.Contains(t, violations[0].Message, "missing required export 'default'")
}

func TestRequiredExports_PassWithDefaultExport(t *testing.T) {
	rule := &RequiredExports{}
	file := &model.UnifiedFileModel{
		Path:     "src/features/auth/index.ts",
		Language: "typescript",
		Source:   []byte("export default authService\n"),
	}
	cfg := model.RuleConfig{Options: map[string]interface{}{
		"patterns": map[string]interface{}{
			"src/features/*/index.ts": map[string]interface{}{
				"required": []interface{}{"default"},
			},
		},
	}}

	violations := rule.Check(file, nil, cfg)
	assert.Empty(t, violations)
}

func TestRequiredExports_WildcardRequirement(t *testing.T) {
	rule := &RequiredExports{}
	file := &model.UnifiedFileModel{
		Path:     "src/services/user/index.ts",
		Language: "typescript",
		Source:   []byte("export function createUserService() {}\n"),
	}
	cfg := model.RuleConfig{Options: map[string]interface{}{
		"patterns": map[string]interface{}{
			"src/services/*/index.ts": map[string]interface{}{
				"required": []interface{}{"create*Service"},
			},
		},
	}}

	violations := rule.Check(file, nil, cfg)
	assert.Empty(t, violations)
}

func TestRequiredExports_NonMatchingPatternIgnored(t *testing.T) {
	rule := &RequiredExports{}
	file := &model.UnifiedFileModel{
		Path:     "src/app/main.ts",
		Language: "typescript",
		Source:   []byte("export const x = 1\n"),
	}
	cfg := model.RuleConfig{Options: map[string]interface{}{
		"patterns": map[string]interface{}{
			"src/features/*/index.ts": map[string]interface{}{
				"required": []interface{}{"default"},
			},
		},
	}}

	violations := rule.Check(file, nil, cfg)
	assert.Empty(t, violations)
}
