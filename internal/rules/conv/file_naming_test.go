// file_naming_test.go — Tests for CONV-file-naming rule.
package conv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stricture/stricture/internal/model"
)

// ---------------------------------------------------------------------------
// 1. Interface compliance
// ---------------------------------------------------------------------------

func TestFileNaming_InterfaceCompliance(t *testing.T) {
	rule := &FileNaming{}

	// Verify it satisfies model.Rule.
	var _ model.Rule = rule

	assert.Equal(t, "CONV-file-naming", rule.ID())
	assert.Equal(t, "conv", rule.Category())
	assert.Equal(t, "error", rule.DefaultSeverity())
	assert.Equal(t, "Enforce file naming convention", rule.Description())
	assert.False(t, rule.NeedsProjectContext())
	assert.Equal(t,
		"Inconsistent naming makes files hard to find and breaks tooling assumptions.",
		rule.Why(),
	)
}

// ---------------------------------------------------------------------------
// 2. kebab-case validation
// ---------------------------------------------------------------------------

func TestFileNaming_KebabCase(t *testing.T) {
	rule := &FileNaming{}
	cfg := model.RuleConfig{Options: map[string]interface{}{"style": "kebab-case"}}

	t.Run("valid: user-service.ts", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/user-service.ts", Language: "typescript"}
		violations := rule.Check(file, nil, cfg)
		assert.Empty(t, violations)
	})

	t.Run("valid: api-client.ts", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/api-client.ts", Language: "typescript"}
		violations := rule.Check(file, nil, cfg)
		assert.Empty(t, violations)
	})

	t.Run("valid: single word index.ts", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/index.ts", Language: "typescript"}
		violations := rule.Check(file, nil, cfg)
		assert.Empty(t, violations)
	})

	t.Run("invalid: PascalCase UserService.ts", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/UserService.ts", Language: "typescript"}
		violations := rule.Check(file, nil, cfg)
		require.Len(t, violations, 1)
		assert.Equal(t, "CONV-file-naming", violations[0].RuleID)
		assert.Contains(t, violations[0].Message, "UserService.ts")
		assert.Contains(t, violations[0].Message, "kebab-case")
		assert.Contains(t, violations[0].Message, "user-service.ts")
		assert.Equal(t, 1, violations[0].StartLine)
	})

	t.Run("invalid: snake_case user_service.ts", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/user_service.ts", Language: "typescript"}
		violations := rule.Check(file, nil, cfg)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "user_service.ts")
		assert.Contains(t, violations[0].Message, "kebab-case")
		assert.Contains(t, violations[0].Message, "user-service.ts")
	})

	t.Run("invalid: camelCase userService.ts", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/userService.ts", Language: "typescript"}
		violations := rule.Check(file, nil, cfg)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "user-service.ts")
	})
}

// ---------------------------------------------------------------------------
// 3. snake_case validation
// ---------------------------------------------------------------------------

func TestFileNaming_SnakeCase(t *testing.T) {
	rule := &FileNaming{}
	cfg := model.RuleConfig{Options: map[string]interface{}{"style": "snake_case"}}

	t.Run("valid: user_service.go", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/internal/user_service.go", Language: "go"}
		violations := rule.Check(file, nil, cfg)
		assert.Empty(t, violations)
	})

	t.Run("valid: http_client.go", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/internal/http_client.go", Language: "go"}
		violations := rule.Check(file, nil, cfg)
		assert.Empty(t, violations)
	})

	t.Run("valid: single word main.go", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/cmd/main.go", Language: "go"}
		violations := rule.Check(file, nil, cfg)
		assert.Empty(t, violations)
	})

	t.Run("invalid: camelCase userService.go", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/internal/userService.go", Language: "go"}
		violations := rule.Check(file, nil, cfg)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "userService.go")
		assert.Contains(t, violations[0].Message, "snake_case")
		assert.Contains(t, violations[0].Message, "user_service.go")
	})

	t.Run("invalid: kebab-case user-service.go", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/internal/user-service.go", Language: "go"}
		violations := rule.Check(file, nil, cfg)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "user-service.go")
		assert.Contains(t, violations[0].Message, "user_service.go")
	})

	t.Run("invalid: PascalCase UserService.go", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/internal/UserService.go", Language: "go"}
		violations := rule.Check(file, nil, cfg)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "user_service.go")
	})
}

// ---------------------------------------------------------------------------
// 4. camelCase validation
// ---------------------------------------------------------------------------

func TestFileNaming_CamelCase(t *testing.T) {
	rule := &FileNaming{}
	cfg := model.RuleConfig{Options: map[string]interface{}{"style": "camelCase"}}

	t.Run("valid: userService.ts", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/userService.ts", Language: "typescript"}
		violations := rule.Check(file, nil, cfg)
		assert.Empty(t, violations)
	})

	t.Run("valid: single word index.ts", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/index.ts", Language: "typescript"}
		violations := rule.Check(file, nil, cfg)
		assert.Empty(t, violations)
	})

	t.Run("invalid: PascalCase UserService.ts", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/UserService.ts", Language: "typescript"}
		violations := rule.Check(file, nil, cfg)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "UserService.ts")
		assert.Contains(t, violations[0].Message, "camelCase")
		assert.Contains(t, violations[0].Message, "userService.ts")
	})

	t.Run("invalid: kebab-case user-service.ts", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/user-service.ts", Language: "typescript"}
		violations := rule.Check(file, nil, cfg)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "userService.ts")
	})

	t.Run("invalid: snake_case user_service.ts", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/user_service.ts", Language: "typescript"}
		violations := rule.Check(file, nil, cfg)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "userService.ts")
	})
}

// ---------------------------------------------------------------------------
// 5. PascalCase validation
// ---------------------------------------------------------------------------

func TestFileNaming_PascalCase(t *testing.T) {
	rule := &FileNaming{}
	cfg := model.RuleConfig{Options: map[string]interface{}{"style": "PascalCase"}}

	t.Run("valid: UserService.java", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/UserService.java", Language: "java"}
		violations := rule.Check(file, nil, cfg)
		assert.Empty(t, violations)
	})

	t.Run("valid: HttpClient.java", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/HttpClient.java", Language: "java"}
		violations := rule.Check(file, nil, cfg)
		assert.Empty(t, violations)
	})

	t.Run("invalid: camelCase userService.java", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/userService.java", Language: "java"}
		violations := rule.Check(file, nil, cfg)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "userService.java")
		assert.Contains(t, violations[0].Message, "PascalCase")
		assert.Contains(t, violations[0].Message, "UserService.java")
	})

	t.Run("invalid: kebab-case user-service.java", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/user-service.java", Language: "java"}
		violations := rule.Check(file, nil, cfg)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "UserService.java")
	})

	t.Run("invalid: snake_case user_service.java", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/user_service.java", Language: "java"}
		violations := rule.Check(file, nil, cfg)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "UserService.java")
	})
}

// ---------------------------------------------------------------------------
// 6. Auto-detect per language (no style in config)
// ---------------------------------------------------------------------------

func TestFileNaming_AutoDetect(t *testing.T) {
	rule := &FileNaming{}
	emptyConfig := model.RuleConfig{Options: map[string]interface{}{}}

	t.Run("Go auto-detects snake_case: valid", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/internal/user_service.go", Language: "go"}
		violations := rule.Check(file, nil, emptyConfig)
		assert.Empty(t, violations)
	})

	t.Run("Go auto-detects snake_case: violation", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/internal/userService.go", Language: "go"}
		violations := rule.Check(file, nil, emptyConfig)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "snake_case")
	})

	t.Run("TypeScript auto-detects kebab-case: valid", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/user-service.ts", Language: "typescript"}
		violations := rule.Check(file, nil, emptyConfig)
		assert.Empty(t, violations)
	})

	t.Run("TypeScript auto-detects kebab-case: violation", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/UserService.ts", Language: "typescript"}
		violations := rule.Check(file, nil, emptyConfig)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "kebab-case")
	})

	t.Run("JavaScript auto-detects kebab-case: valid", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/api-client.js", Language: "javascript"}
		violations := rule.Check(file, nil, emptyConfig)
		assert.Empty(t, violations)
	})

	t.Run("JavaScript auto-detects kebab-case: violation", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/ApiClient.js", Language: "javascript"}
		violations := rule.Check(file, nil, emptyConfig)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "kebab-case")
	})

	t.Run("Python auto-detects snake_case: valid", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/user_service.py", Language: "python"}
		violations := rule.Check(file, nil, emptyConfig)
		assert.Empty(t, violations)
	})

	t.Run("Python auto-detects snake_case: violation", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/UserService.py", Language: "python"}
		violations := rule.Check(file, nil, emptyConfig)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "snake_case")
	})

	t.Run("Java auto-detects PascalCase: valid", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/UserService.java", Language: "java"}
		violations := rule.Check(file, nil, emptyConfig)
		assert.Empty(t, violations)
	})

	t.Run("Java auto-detects PascalCase: violation", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/user-service.java", Language: "java"}
		violations := rule.Check(file, nil, emptyConfig)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "PascalCase")
	})

	t.Run("Unknown language with no config returns no violations", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/readme.txt", Language: "plaintext"}
		violations := rule.Check(file, nil, emptyConfig)
		assert.Empty(t, violations)
	})
}

// ---------------------------------------------------------------------------
// 7. Multi-extension handling
// ---------------------------------------------------------------------------

func TestFileNaming_MultiExtensions(t *testing.T) {
	rule := &FileNaming{}
	kebabCfg := model.RuleConfig{Options: map[string]interface{}{"style": "kebab-case"}}
	snakeCfg := model.RuleConfig{Options: map[string]interface{}{"style": "snake_case"}}

	t.Run("valid: user-service.test.ts (kebab-case)", func(t *testing.T) {
		file := &model.UnifiedFileModel{
			Path:       "/project/src/user-service.test.ts",
			Language:   "typescript",
			IsTestFile: true,
		}
		violations := rule.Check(file, nil, kebabCfg)
		assert.Empty(t, violations)
	})

	t.Run("valid: user-service.spec.ts (kebab-case)", func(t *testing.T) {
		file := &model.UnifiedFileModel{
			Path:       "/project/src/user-service.spec.ts",
			Language:   "typescript",
			IsTestFile: true,
		}
		violations := rule.Check(file, nil, kebabCfg)
		assert.Empty(t, violations)
	})

	t.Run("valid: user-service.d.ts (kebab-case)", func(t *testing.T) {
		file := &model.UnifiedFileModel{
			Path:     "/project/src/user-service.d.ts",
			Language: "typescript",
		}
		violations := rule.Check(file, nil, kebabCfg)
		assert.Empty(t, violations)
	})

	t.Run("invalid: UserService.test.ts (kebab-case)", func(t *testing.T) {
		file := &model.UnifiedFileModel{
			Path:       "/project/src/UserService.test.ts",
			Language:   "typescript",
			IsTestFile: true,
		}
		violations := rule.Check(file, nil, kebabCfg)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "user-service.test.ts")
	})

	t.Run("invalid: UserService.spec.js (kebab-case)", func(t *testing.T) {
		file := &model.UnifiedFileModel{
			Path:       "/project/src/UserService.spec.js",
			Language:   "javascript",
			IsTestFile: true,
		}
		violations := rule.Check(file, nil, kebabCfg)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "user-service.spec.js")
	})

	t.Run("invalid: UserService.d.ts (kebab-case)", func(t *testing.T) {
		file := &model.UnifiedFileModel{
			Path:     "/project/src/UserService.d.ts",
			Language: "typescript",
		}
		violations := rule.Check(file, nil, kebabCfg)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "user-service.d.ts")
	})

	t.Run("valid: user_service_test.go (snake_case, Go test convention)", func(t *testing.T) {
		// Go test files end in _test.go, not .test.go. The _test portion
		// is part of the base name and should still pass snake_case.
		file := &model.UnifiedFileModel{
			Path:       "/project/internal/user_service_test.go",
			Language:   "go",
			IsTestFile: true,
		}
		violations := rule.Check(file, nil, snakeCfg)
		assert.Empty(t, violations)
	})
}

// ---------------------------------------------------------------------------
// 8. Edge cases: single-word names, numbers
// ---------------------------------------------------------------------------

func TestFileNaming_EdgeCases(t *testing.T) {
	rule := &FileNaming{}

	t.Run("single word: main.go passes snake_case", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/cmd/main.go", Language: "go"}
		cfg := model.RuleConfig{Options: map[string]interface{}{"style": "snake_case"}}
		violations := rule.Check(file, nil, cfg)
		assert.Empty(t, violations)
	})

	t.Run("single word: index.ts passes kebab-case", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/index.ts", Language: "typescript"}
		cfg := model.RuleConfig{Options: map[string]interface{}{"style": "kebab-case"}}
		violations := rule.Check(file, nil, cfg)
		assert.Empty(t, violations)
	})

	t.Run("numbers: v2-api.ts passes kebab-case", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/v2-api.ts", Language: "typescript"}
		cfg := model.RuleConfig{Options: map[string]interface{}{"style": "kebab-case"}}
		violations := rule.Check(file, nil, cfg)
		assert.Empty(t, violations)
	})

	t.Run("numbers: stage2.go passes snake_case (single word with trailing digit)", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/cmd/stage2.go", Language: "go"}
		cfg := model.RuleConfig{Options: map[string]interface{}{"style": "snake_case"}}
		violations := rule.Check(file, nil, cfg)
		assert.Empty(t, violations)
	})

	t.Run("numbers: api2client.ts passes kebab-case (single lowercase token)", func(t *testing.T) {
		// "api2client" is a single lowercase+digit token, which is valid kebab-case.
		// The regex ^[a-z][a-z0-9]*(-[a-z0-9]+)*$ matches it as one segment.
		file := &model.UnifiedFileModel{Path: "/project/src/api2client.ts", Language: "typescript"}
		cfg := model.RuleConfig{Options: map[string]interface{}{"style": "kebab-case"}}
		violations := rule.Check(file, nil, cfg)
		assert.Empty(t, violations)
	})

	t.Run("numbers: Api2Client.ts violates kebab-case", func(t *testing.T) {
		// PascalCase with digits should be caught and converted.
		file := &model.UnifiedFileModel{Path: "/project/src/Api2Client.ts", Language: "typescript"}
		cfg := model.RuleConfig{Options: map[string]interface{}{"style": "kebab-case"}}
		violations := rule.Check(file, nil, cfg)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "kebab-case")
	})

	t.Run("nil options uses auto-detect", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/internal/user_service.go", Language: "go"}
		cfg := model.RuleConfig{} // nil Options
		violations := rule.Check(file, nil, cfg)
		assert.Empty(t, violations)
	})

	t.Run("empty language with no style returns no violations", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/Makefile", Language: ""}
		cfg := model.RuleConfig{}
		violations := rule.Check(file, nil, cfg)
		assert.Empty(t, violations)
	})
}

// ---------------------------------------------------------------------------
// 9. Suggested name generation (detailed verification)
// ---------------------------------------------------------------------------

func TestFileNaming_SuggestedNameGeneration(t *testing.T) {
	rule := &FileNaming{}

	tests := []struct {
		name          string
		path          string
		language      string
		style         string
		wantSuggested string
	}{
		{
			name:          "PascalCase to kebab-case",
			path:          "/project/src/UserService.ts",
			language:      "typescript",
			style:         "kebab-case",
			wantSuggested: "user-service.ts",
		},
		{
			name:          "camelCase to kebab-case",
			path:          "/project/src/userService.ts",
			language:      "typescript",
			style:         "kebab-case",
			wantSuggested: "user-service.ts",
		},
		{
			name:          "snake_case to kebab-case",
			path:          "/project/src/user_service.ts",
			language:      "typescript",
			style:         "kebab-case",
			wantSuggested: "user-service.ts",
		},
		{
			name:          "kebab-case to snake_case",
			path:          "/project/internal/user-service.go",
			language:      "go",
			style:         "snake_case",
			wantSuggested: "user_service.go",
		},
		{
			name:          "PascalCase to snake_case",
			path:          "/project/internal/UserService.go",
			language:      "go",
			style:         "snake_case",
			wantSuggested: "user_service.go",
		},
		{
			name:          "kebab-case to camelCase",
			path:          "/project/src/user-service.ts",
			language:      "typescript",
			style:         "camelCase",
			wantSuggested: "userService.ts",
		},
		{
			name:          "snake_case to PascalCase",
			path:          "/project/src/user_service.java",
			language:      "java",
			style:         "PascalCase",
			wantSuggested: "UserService.java",
		},
		{
			name:          "PascalCase test file to kebab-case preserves .test.ts",
			path:          "/project/src/UserService.test.ts",
			language:      "typescript",
			style:         "kebab-case",
			wantSuggested: "user-service.test.ts",
		},
		{
			name:          "PascalCase type decl to kebab-case preserves .d.ts",
			path:          "/project/src/UserService.d.ts",
			language:      "typescript",
			style:         "kebab-case",
			wantSuggested: "user-service.d.ts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := &model.UnifiedFileModel{Path: tt.path, Language: tt.language}
			cfg := model.RuleConfig{Options: map[string]interface{}{"style": tt.style}}

			violations := rule.Check(file, nil, cfg)
			require.Len(t, violations, 1, "expected exactly one violation")
			assert.Contains(t, violations[0].Message, tt.wantSuggested,
				"violation message should contain the suggested name")

			require.NotNil(t, violations[0].Context, "violation should have Context")
			assert.Contains(t, violations[0].Context.SuggestedFix, tt.wantSuggested,
				"SuggestedFix should contain the suggested name")
		})
	}
}

// ---------------------------------------------------------------------------
// 10. Violation structure verification
// ---------------------------------------------------------------------------

func TestFileNaming_ViolationStructure(t *testing.T) {
	rule := &FileNaming{}
	file := &model.UnifiedFileModel{
		Path:     "/project/src/UserService.ts",
		Language: "typescript",
	}
	cfg := model.RuleConfig{
		Severity: "warn",
		Options:  map[string]interface{}{"style": "kebab-case"},
	}

	violations := rule.Check(file, nil, cfg)
	require.Len(t, violations, 1)

	v := violations[0]
	assert.Equal(t, "CONV-file-naming", v.RuleID)
	assert.Equal(t, "warn", v.Severity, "should use config severity when set")
	assert.Equal(t, "/project/src/UserService.ts", v.FilePath)
	assert.Equal(t, 1, v.StartLine, "file-level violations use line 1")

	// Verify message matches error-catalog template:
	// "File name '{actual_name}' does not match convention '{expected_convention}', should be '{suggested_name}'"
	assert.Equal(t,
		"File name 'UserService.ts' does not match convention 'kebab-case', should be 'user-service.ts'",
		v.Message,
	)

	// Verify context.
	require.NotNil(t, v.Context)
	assert.Equal(t, "Rename to 'user-service.ts' using kebab-case.", v.Context.SuggestedFix)
}

// ---------------------------------------------------------------------------
// 11. Default severity when config.Severity is empty
// ---------------------------------------------------------------------------

func TestFileNaming_DefaultSeverity(t *testing.T) {
	rule := &FileNaming{}
	file := &model.UnifiedFileModel{
		Path:     "/project/src/UserService.ts",
		Language: "typescript",
	}
	cfg := model.RuleConfig{
		// Severity intentionally empty — should fall back to "error".
		Options: map[string]interface{}{"style": "kebab-case"},
	}

	violations := rule.Check(file, nil, cfg)
	require.Len(t, violations, 1)
	assert.Equal(t, "error", violations[0].Severity)
}

// ---------------------------------------------------------------------------
// 12. Config style overrides language auto-detect
// ---------------------------------------------------------------------------

func TestFileNaming_ConfigOverridesAutoDetect(t *testing.T) {
	rule := &FileNaming{}

	t.Run("Go file with kebab-case override", func(t *testing.T) {
		// Go normally defaults to snake_case, but config says kebab-case.
		file := &model.UnifiedFileModel{Path: "/project/internal/user-service.go", Language: "go"}
		cfg := model.RuleConfig{Options: map[string]interface{}{"style": "kebab-case"}}
		violations := rule.Check(file, nil, cfg)
		assert.Empty(t, violations, "kebab-case file should pass when config overrides to kebab-case")
	})

	t.Run("Go file with kebab-case override rejects snake_case", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/internal/user_service.go", Language: "go"}
		cfg := model.RuleConfig{Options: map[string]interface{}{"style": "kebab-case"}}
		violations := rule.Check(file, nil, cfg)
		require.Len(t, violations, 1)
		assert.Contains(t, violations[0].Message, "kebab-case")
	})

	t.Run("TypeScript file with PascalCase override", func(t *testing.T) {
		file := &model.UnifiedFileModel{Path: "/project/src/UserService.ts", Language: "typescript"}
		cfg := model.RuleConfig{Options: map[string]interface{}{"style": "PascalCase"}}
		violations := rule.Check(file, nil, cfg)
		assert.Empty(t, violations, "PascalCase file should pass when config overrides to PascalCase")
	})
}

// ---------------------------------------------------------------------------
// 13. Suppression comment documentation test
// ---------------------------------------------------------------------------

func TestFileNaming_SuppressionComment(t *testing.T) {
	// NOTE: The rule itself does not handle suppression — that is the engine's
	// responsibility. This test documents the expected suppression comment syntax
	// from the error catalog. The engine should skip running this rule when it
	// finds: // stricture-disable-next-line CONV-file-naming
	//
	// A file with a suppression comment should still trigger the rule if the
	// engine does not strip it. This test verifies the rule fires independently
	// of suppression handling.
	rule := &FileNaming{}
	file := &model.UnifiedFileModel{
		Path:     "/project/src/UserService.ts",
		Language: "typescript",
		Source:   []byte("// stricture-disable-next-line CONV-file-naming\nexport class UserService {}\n"),
	}
	cfg := model.RuleConfig{Options: map[string]interface{}{"style": "kebab-case"}}

	violations := rule.Check(file, nil, cfg)
	require.Len(t, violations, 1, "rule should fire; suppression is handled by the engine, not the rule")
}

// ---------------------------------------------------------------------------
// 14. Internal helper: extractBaseName
// ---------------------------------------------------------------------------

func TestExtractBaseName(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/project/src/user-service.ts", "user-service"},
		{"/project/src/user-service.test.ts", "user-service"},
		{"/project/src/user-service.spec.ts", "user-service"},
		{"/project/src/user-service.d.ts", "user-service"},
		{"/project/src/user-service.test.tsx", "user-service"},
		{"/project/src/user-service.spec.jsx", "user-service"},
		{"/project/internal/user_service.go", "user_service"},
		{"/project/internal/user_service_test.go", "user_service_test"},
		{"/project/cmd/main.go", "main"},
		{"/project/Makefile", "Makefile"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := extractBaseName(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// 15. Internal helper: convertToConvention
// ---------------------------------------------------------------------------

func TestConvertToConvention(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		convention string
		want       string
	}{
		{"PascalCase to kebab", "UserService", "kebab-case", "user-service"},
		{"camelCase to kebab", "userService", "kebab-case", "user-service"},
		{"snake to kebab", "user_service", "kebab-case", "user-service"},
		{"kebab to snake", "user-service", "snake_case", "user_service"},
		{"PascalCase to snake", "UserService", "snake_case", "user_service"},
		{"kebab to camel", "user-service", "camelCase", "userService"},
		{"snake to Pascal", "user_service", "PascalCase", "UserService"},
		{"single word to kebab", "main", "kebab-case", "main"},
		{"single word to Pascal", "main", "PascalCase", "Main"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertToConvention(tt.input, tt.convention)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// 16. Internal helper: splitIntoWords
// ---------------------------------------------------------------------------

func TestSplitIntoWords(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"user-service", []string{"user", "service"}},
		{"user_service", []string{"user", "service"}},
		{"userService", []string{"user", "service"}},
		{"UserService", []string{"user", "service"}},
		{"HTTPClient", []string{"http", "client"}},
		{"getAPIKey", []string{"get", "api", "key"}},
		{"main", []string{"main"}},
		{"v2API", []string{"v", "2", "api"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := splitIntoWords(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
