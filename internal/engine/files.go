// files.go — Reads source files into UnifiedFileModels.
package engine

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/stricture/stricture/internal/adapter"
	"github.com/stricture/stricture/internal/model"
)

// BuildFiles reads each path into a UnifiedFileModel carrying source text, language, and line count.
func BuildFiles(paths []string) ([]*model.UnifiedFileModel, error) {
	files := make([]*model.UnifiedFileModel, 0, len(paths))
	for _, pathValue := range paths {
		data, err := os.ReadFile(pathValue)
		if err != nil {
			return nil, err
		}

		file := &model.UnifiedFileModel{
			Path:       filepath.ToSlash(pathValue),
			Language:   adapter.DetectLanguage(pathValue),
			Source:     data,
			LineCount:  CountLines(data),
			IsTestFile: LooksLikeTestFile(pathValue),
		}
		files = append(files, file)
	}
	return files, nil
}

// CountLines returns the number of lines in data (a trailing newline starts a final empty line).
func CountLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	count := 1
	for _, b := range data {
		if b == '\n' {
			count++
		}
	}
	return count
}

// LooksLikeTestFile reports whether a path follows common test file naming conventions.
func LooksLikeTestFile(pathValue string) bool {
	name := strings.ToLower(filepath.Base(pathValue))
	return strings.HasSuffix(name, "_test.go") ||
		strings.Contains(name, ".test.") ||
		strings.Contains(name, ".spec.") ||
		strings.HasPrefix(name, "test_") ||
		strings.HasSuffix(name, "test.java")
}
