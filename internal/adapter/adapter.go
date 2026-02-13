// adapter.go — LanguageAdapter interface and AdapterConfig.
package adapter

import "github.com/stricture/stricture/internal/model"

// LanguageAdapter defines the interface for language-specific parsers.
type LanguageAdapter interface {
	// Name returns the language name (e.g., "go", "typescript", "python").
	Name() string

	// Extensions returns the file extensions this adapter handles (e.g., [".go"]).
	Extensions() []string

	// Parse parses a file and returns a UnifiedFileModel.
	Parse(path string, source []byte, config AdapterConfig) (*model.UnifiedFileModel, error)

	// IsTestFile determines if a file is a test file.
	IsTestFile(path string) bool
}

// AdapterConfig holds configuration for a language adapter.
//
//nolint:revive // AdapterConfig is intentionally explicit at package boundaries.
type AdapterConfig struct {
	IncludeComments bool
	MaxFileSize     int64
	Timeout         int
}
