// builder.go — Fluent builders for test fixtures.
package testutil

import (
	"fmt"
	"strings"

	"github.com/stricture/stricture/internal/model"
)

// FileBuilder builds UnifiedFileModel instances for testing.
type FileBuilder struct {
	file model.UnifiedFileModel
}

// NewFile creates a FileBuilder with path and language auto-detected from extension.
func NewFile(path string) *FileBuilder {
	lang := detectLanguage(path)
	return &FileBuilder{
		file: model.UnifiedFileModel{
			Path:     path,
			Language: lang,
		},
	}
}

// WithLanguage overrides the auto-detected language.
func (b *FileBuilder) WithLanguage(lang string) *FileBuilder {
	b.file.Language = lang
	return b
}

// WithSource sets the file source content.
func (b *FileBuilder) WithSource(src string) *FileBuilder {
	b.file.Source = []byte(src)
	b.file.LineCount = countLines(src)
	return b
}

// AsTestFile marks the file as a test file.
func (b *FileBuilder) AsTestFile() *FileBuilder {
	b.file.IsTestFile = true
	return b
}

// WithImport adds an import declaration.
func (b *FileBuilder) WithImport(path string) *FileBuilder {
	b.file.Imports = append(b.file.Imports, model.ImportDecl{
		Path:      path,
		StartLine: len(b.file.Imports) + 1,
	})
	return b
}

// WithImportAlias adds an aliased import declaration.
func (b *FileBuilder) WithImportAlias(path, alias string) *FileBuilder {
	b.file.Imports = append(b.file.Imports, model.ImportDecl{
		Path:      path,
		Alias:     alias,
		StartLine: len(b.file.Imports) + 1,
	})
	return b
}

// WithExport adds an export declaration.
func (b *FileBuilder) WithExport(name, kind string) *FileBuilder {
	b.file.Exports = append(b.file.Exports, model.ExportDecl{
		Name:      name,
		Kind:      kind,
		StartLine: len(b.file.Exports) + 1,
	})
	return b
}

// WithFunc adds a function to the file.
func (b *FileBuilder) WithFunc(name string, params int) *FileBuilder {
	f := model.FuncModel{
		Name:       name,
		IsExported: isExported(name),
		StartLine:  len(b.file.Functions)*10 + 1,
		EndLine:    len(b.file.Functions)*10 + 10,
		LineCount:  10,
	}
	for i := 0; i < params; i++ {
		f.Params = append(f.Params, model.ParamModel{
			Name: fmt.Sprintf("arg%d", i),
			Type: "string",
		})
	}
	b.file.Functions = append(b.file.Functions, f)
	return b
}

// WithTestFunc adds a test function to the file.
func (b *FileBuilder) WithTestFunc(name string) *FileBuilder {
	f := model.FuncModel{
		Name:       name,
		IsTest:     true,
		IsExported: true,
		StartLine:  len(b.file.Functions)*10 + 1,
		EndLine:    len(b.file.Functions)*10 + 10,
		LineCount:  10,
	}
	b.file.Functions = append(b.file.Functions, f)
	return b
}

// WithFuncDetail adds a fully-specified function.
func (b *FileBuilder) WithFuncDetail(f model.FuncModel) *FileBuilder {
	b.file.Functions = append(b.file.Functions, f)
	return b
}

// WithType adds a type definition.
func (b *FileBuilder) WithType(name, kind string, fields ...model.FieldModel) *FileBuilder {
	b.file.Types = append(b.file.Types, model.TypeModel{
		Name:     name,
		Kind:     kind,
		Fields:   fields,
		Exported: isExported(name),
	})
	return b
}

// WithClass adds a class.
func (b *FileBuilder) WithClass(name string) *FileBuilder {
	b.file.Classes = append(b.file.Classes, model.ClassModel{
		Name:     name,
		Exported: isExported(name),
	})
	return b
}

// WithTestCase adds a test case.
func (b *FileBuilder) WithTestCase(name string, assertions ...model.Assertion) *FileBuilder {
	b.file.TestCases = append(b.file.TestCases, model.TestCase{
		Name:       name,
		Assertions: assertions,
		StartLine:  len(b.file.TestCases)*20 + 1,
		EndLine:    len(b.file.TestCases)*20 + 20,
	})
	return b
}

// WithTestTarget adds a test target (source file the test covers).
func (b *FileBuilder) WithTestTarget(target string) *FileBuilder {
	b.file.TestTargets = append(b.file.TestTargets, target)
	return b
}

// WithLineCount overrides the auto-computed line count.
func (b *FileBuilder) WithLineCount(n int) *FileBuilder {
	b.file.LineCount = n
	return b
}

// Build returns the constructed UnifiedFileModel.
func (b *FileBuilder) Build() *model.UnifiedFileModel {
	result := b.file
	return &result
}

// --- helpers ---

func detectLanguage(path string) string {
	switch {
	case strings.HasSuffix(path, ".go"):
		return "go"
	case strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".tsx"):
		return "typescript"
	case strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".jsx"):
		return "javascript"
	case strings.HasSuffix(path, ".py"):
		return "python"
	case strings.HasSuffix(path, ".java"):
		return "java"
	default:
		return "unknown"
	}
}

func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

func countLines(s string) int {
	if len(s) == 0 {
		return 0
	}
	n := 1
	for _, c := range s {
		if c == '\n' {
			n++
		}
	}
	return n
}
