// file_naming_test.go — Tests for CONV-file-naming rule.
package conv

import (
	"testing"

	"github.com/stricture/stricture/internal/model"
)

func TestFileNaming(t *testing.T) {
	rule := &FileNaming{}

	// Verify interface compliance
	if rule.ID() != "CONV-file-naming" {
		t.Errorf("ID() = %s, want CONV-file-naming", rule.ID())
	}
	if rule.Category() != "conv" {
		t.Errorf("Category() = %s, want conv", rule.Category())
	}
	if rule.DefaultSeverity() != "error" {
		t.Errorf("DefaultSeverity() = %s, want error", rule.DefaultSeverity())
	}
	if rule.NeedsProjectContext() {
		t.Error("NeedsProjectContext() = true, want false")
	}

	tests := []struct {
		name      string
		file      *model.UnifiedFileModel
		config    model.RuleConfig
		wantCount int
	}{
		{
			name: "kebab-case TypeScript passes",
			file: &model.UnifiedFileModel{
				Path:     "/project/src/user-service.ts",
				Language: "typescript",
			},
			config:    model.RuleConfig{Options: map[string]interface{}{"style": "kebab-case"}},
			wantCount: 0,
		},
		{
			name: "PascalCase TypeScript violates kebab-case",
			file: &model.UnifiedFileModel{
				Path:     "/project/src/UserService.ts",
				Language: "typescript",
			},
			config:    model.RuleConfig{Options: map[string]interface{}{"style": "kebab-case"}},
			wantCount: 1,
		},
		{
			name: "snake_case Go passes",
			file: &model.UnifiedFileModel{
				Path:     "/project/internal/user_service.go",
				Language: "go",
			},
			config:    model.RuleConfig{Options: map[string]interface{}{"style": "snake_case"}},
			wantCount: 0,
		},
		{
			name: "camelCase Go violates snake_case",
			file: &model.UnifiedFileModel{
				Path:     "/project/internal/userService.go",
				Language: "go",
			},
			config:    model.RuleConfig{Options: map[string]interface{}{"style": "snake_case"}},
			wantCount: 1,
		},
		{
			name: "test file follows convention",
			file: &model.UnifiedFileModel{
				Path:       "/project/src/user-service.test.ts",
				Language:   "typescript",
				IsTestFile: true,
			},
			config:    model.RuleConfig{Options: map[string]interface{}{"style": "kebab-case"}},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := rule.Check(tt.file, nil, tt.config)
			if len(violations) != tt.wantCount {
				t.Errorf("Check() returned %d violations, want %d", len(violations), tt.wantCount)
			}
			if tt.wantCount > 0 && len(violations) > 0 {
				if violations[0].RuleID != "CONV-file-naming" {
					t.Errorf("violation.RuleID = %s, want CONV-file-naming", violations[0].RuleID)
				}
				if violations[0].Message == "" {
					t.Error("violation.Message is empty")
				}
				if violations[0].Context != nil && violations[0].Context.SuggestedFix == "" {
					t.Error("violation.Context.SuggestedFix is empty")
				}
			}
		})
	}
}
