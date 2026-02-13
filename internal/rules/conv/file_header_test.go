// file_header_test.go — Tests for CONV-file-header rule.
package conv

import (
	"testing"

	"github.com/stricture/stricture/internal/model"
)

func TestFileHeader(t *testing.T) {
	rule := &FileHeader{}

	if rule.ID() != "CONV-file-header" {
		t.Errorf("ID() = %s, want CONV-file-header", rule.ID())
	}
	if rule.Category() != "conv" {
		t.Errorf("Category() = %s, want conv", rule.Category())
	}

	tests := []struct {
		name      string
		file      *model.UnifiedFileModel
		config    model.RuleConfig
		wantCount int
	}{
		{
			name: "Go file with proper header passes",
			file: &model.UnifiedFileModel{
				Path:     "/project/internal/user_service.go",
				Language: "go",
				Source:   []byte("// user_service.go — User service implementation.\npackage service\n"),
			},
			config:    model.RuleConfig{Options: map[string]interface{}{"pattern": "// {filename} — {purpose}"}},
			wantCount: 0,
		},
		{
			name: "Go file without header fails",
			file: &model.UnifiedFileModel{
				Path:     "/project/internal/user_service.go",
				Language: "go",
				Source:   []byte("package service\n\nfunc GetUser() {}\n"),
			},
			config:    model.RuleConfig{Options: map[string]interface{}{"pattern": "// {filename} — {purpose}"}},
			wantCount: 1,
		},
		{
			name: "TypeScript file with header passes",
			file: &model.UnifiedFileModel{
				Path:     "/project/src/user-service.ts",
				Language: "typescript",
				Source:   []byte("// user-service.ts — User service module.\n\nexport class UserService {}\n"),
			},
			config:    model.RuleConfig{Options: map[string]interface{}{"pattern": "// {filename} — {purpose}"}},
			wantCount: 0,
		},
		{
			name: "TypeScript file without header fails",
			file: &model.UnifiedFileModel{
				Path:     "/project/src/user-service.ts",
				Language: "typescript",
				Source:   []byte("export class UserService {}\n"),
			},
			config:    model.RuleConfig{Options: map[string]interface{}{"pattern": "// {filename} — {purpose}"}},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := rule.Check(tt.file, nil, tt.config)
			if len(violations) != tt.wantCount {
				t.Errorf("Check() returned %d violations, want %d", len(violations), tt.wantCount)
			}
		})
	}
}
