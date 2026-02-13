// adapter.go — Lightweight Python adapter implementation.
package python

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/stricture/stricture/internal/adapter"
	"github.com/stricture/stricture/internal/model"
)

var functionPattern = regexp.MustCompile(`(?m)^\s*def\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(`)

// Adapter parses Python files into a UnifiedFileModel.
type Adapter struct{}

func (a *Adapter) Name() string {
	return "python"
}

func (a *Adapter) Extensions() []string {
	return []string{".py"}
}

func (a *Adapter) IsTestFile(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	return strings.HasPrefix(name, "test_") || strings.HasSuffix(name, "_test.py")
}

func (a *Adapter) Parse(path string, source []byte, _ adapter.AdapterConfig) (*model.UnifiedFileModel, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return nil, fmt.Errorf("parse python file: %w", model.ErrParseFailure)
	}

	result := &model.UnifiedFileModel{
		Path:       filepath.ToSlash(trimmedPath),
		Language:   "python",
		Source:     append([]byte(nil), source...),
		LineCount:  countLines(source),
		IsTestFile: a.IsTestFile(trimmedPath),
	}

	for _, match := range functionPattern.FindAllSubmatch(source, -1) {
		if len(match) < 2 {
			continue
		}
		result.Functions = append(result.Functions, model.FuncModel{Name: string(match[1])})
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
