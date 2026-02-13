// violation.go — Violation and ViolationContext types.
package model

// Violation represents a rule violation.
type Violation struct {
	RuleID      string
	Severity    string
	Message     string
	FilePath    string
	StartLine   int
	EndLine     int
	StartColumn int
	EndColumn   int
	Context     *ViolationContext
}

// ViolationContext provides additional context for a violation.
type ViolationContext struct {
	Snippet      string
	SuggestedFix string
	References   []string
	Metadata     map[string]interface{}
}
