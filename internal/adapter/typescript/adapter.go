// adapter.go — Lightweight TypeScript adapter implementation.
package typescript

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/stricture/stricture/internal/adapter"
	"github.com/stricture/stricture/internal/model"
)

var exportPattern = regexp.MustCompile(`(?m)^\s*export\s+(?:const|function|class|interface|type)\s+([A-Za-z_][A-Za-z0-9_]*)`)

// Adapter parses TypeScript/JavaScript files into a UnifiedFileModel.
type Adapter struct{}

func (a *Adapter) Name() string {
	return "typescript"
}

func (a *Adapter) Extensions() []string {
	return []string{".ts", ".tsx", ".js", ".jsx"}
}

func (a *Adapter) IsTestFile(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	return strings.Contains(name, ".test.") || strings.Contains(name, ".spec.")
}

func (a *Adapter) Parse(path string, source []byte, _ adapter.AdapterConfig) (*model.UnifiedFileModel, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return nil, fmt.Errorf("parse typescript file: %w", model.ErrParseFailure)
	}

	language := "typescript"
	ext := strings.ToLower(filepath.Ext(trimmedPath))
	if ext == ".js" || ext == ".jsx" {
		language = "javascript"
	}

	result := &model.UnifiedFileModel{
		Path:       filepath.ToSlash(trimmedPath),
		Language:   language,
		Source:     append([]byte(nil), source...),
		LineCount:  countLines(source),
		IsTestFile: a.IsTestFile(trimmedPath),
	}

	for _, match := range exportPattern.FindAllSubmatch(source, -1) {
		if len(match) < 2 {
			continue
		}
		result.Exports = append(result.Exports, model.ExportDecl{
			Name: string(match[1]),
			Kind: "value",
		})
	}

	return result, nil
}

func countLines(source []byte) int {
	if len(source) == 0 {
		return 0
	}
	count := 1
	for _, b := range source {
		if b == '\n' {
			count++
		}
	}
	return count
}
