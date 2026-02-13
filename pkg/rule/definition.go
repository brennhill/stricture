// Package rule defines the public plugin API for Stricture Go plugins.
package rule

// UnifiedFileModel is the plugin-facing file model.
type UnifiedFileModel struct {
	Path       string
	Language   string
	IsTestFile bool
	Source     string
	LineCount  int
}

// ProjectContext is the plugin-facing project context.
type ProjectContext struct {
	FileCount int
}

// Violation is the plugin-facing violation type.
type Violation struct {
	RuleID       string
	Severity     string
	Message      string
	StartLine    int
	EndLine      int
	SuggestedFix string
}

// Definition is the required exported symbol type for Go plugins.
//
// Plugins must export:
//
//	var Rule = rule.Definition{ ... }
type Definition struct {
	ID                  string
	Category            string
	Severity            string
	Description         string
	Why                 string
	NeedsProjectContext bool
	Check               func(file *UnifiedFileModel, context *ProjectContext, options map[string]interface{}) []Violation
}
