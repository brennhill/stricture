// reporter.go — Reporter interface and Summary type.
package reporter

import "github.com/stricture/stricture/internal/model"

// Reporter defines the interface for output formatters.
type Reporter interface {
	// Format returns the format name (e.g., "text", "json", "sarif").
	Format() string

	// Report outputs violations in the reporter's format.
	Report(violations []model.Violation, summary Summary) error
}

// Summary holds aggregate statistics about a lint run.
type Summary struct {
	TotalFiles      int
	FilesWithIssues int
	TotalViolations int
	ErrorCount      int
	WarningCount    int
	Duration        int64
}
