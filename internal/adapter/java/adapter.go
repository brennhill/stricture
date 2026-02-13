// adapter.go — Lightweight Java adapter implementation.
package java

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/stricture/stricture/internal/adapter"
	"github.com/stricture/stricture/internal/model"
)

var classPattern = regexp.MustCompile(`(?m)^\s*(?:public\s+)?class\s+([A-Za-z_][A-Za-z0-9_]*)`)

// Adapter parses Java files into a UnifiedFileModel.
type Adapter struct{}

func (a *Adapter) Name() string {
	return "java"
}

func (a *Adapter) Extensions() []string {
	return []string{".java"}
}

func (a *Adapter) IsTestFile(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	return strings.HasSuffix(name, "test.java")
}

func (a *Adapter) Parse(path string, source []byte, _ adapter.AdapterConfig) (*model.UnifiedFileModel, error) {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return nil, fmt.Errorf("parse java file: %w", model.ErrParseFailure)
	}

	result := &model.UnifiedFileModel{
		Path:       filepath.ToSlash(trimmedPath),
		Language:   "java",
		Source:     append([]byte(nil), source...),
		LineCount:  countLines(source),
		IsTestFile: a.IsTestFile(trimmedPath),
	}

	for _, match := range classPattern.FindAllSubmatch(source, -1) {
		if len(match) < 2 {
			continue
		}
		result.Classes = append(result.Classes, model.ClassModel{Name: string(match[1]), Exported: true})
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
