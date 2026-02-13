// test_location_test.go - Tests for CONV-test-file-location rule.
package conv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stricture/stricture/internal/model"
)

func TestTestFileLocation_Metadata(t *testing.T) {
	rule := &TestFileLocation{}
	assert.Equal(t, "CONV-test-file-location", rule.ID())
	assert.Equal(t, "conv", rule.Category())
	assert.Equal(t, "error", rule.DefaultSeverity())
	assert.NotEmpty(t, rule.Description())
	assert.NotEmpty(t, rule.Why())
}

func TestTestFileLocation_MirroredStrategyViolation(t *testing.T) {
	rule := &TestFileLocation{}
	file := &model.UnifiedFileModel{
		Path:       "src/services/user.test.ts",
		Language:   "typescript",
		IsTestFile: true,
	}
	cfg := model.RuleConfig{Options: map[string]interface{}{"strategy": "mirrored"}}

	violations := rule.Check(file, nil, cfg)
	require.Len(t, violations, 1)
	assert.Contains(t, violations[0].Message, "should be in")
	assert.Contains(t, violations[0].Message, "tests/services")
}

func TestTestFileLocation_MirroredStrategyPass(t *testing.T) {
	rule := &TestFileLocation{}
	file := &model.UnifiedFileModel{
		Path:       "tests/services/user.test.ts",
		Language:   "typescript",
		IsTestFile: true,
	}
	cfg := model.RuleConfig{Options: map[string]interface{}{"strategy": "mirrored"}}

	violations := rule.Check(file, nil, cfg)
	assert.Empty(t, violations)
}

func TestTestFileLocation_SubfolderViolation(t *testing.T) {
	rule := &TestFileLocation{}
	file := &model.UnifiedFileModel{
		Path:       "src/services/user.test.ts",
		Language:   "typescript",
		IsTestFile: true,
	}
	cfg := model.RuleConfig{Options: map[string]interface{}{"strategy": "subfolder"}}

	violations := rule.Check(file, nil, cfg)
	require.Len(t, violations, 1)
	assert.Contains(t, violations[0].Message, "__tests__")
}

func TestTestFileLocation_IgnoresNonTestFiles(t *testing.T) {
	rule := &TestFileLocation{}
	file := &model.UnifiedFileModel{
		Path:     "src/services/user.ts",
		Language: "typescript",
	}

	violations := rule.Check(file, nil, model.RuleConfig{})
	assert.Empty(t, violations)
}
