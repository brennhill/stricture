// github_client_b12.go — Missing nil check on nullable field
package github

import (
	"fmt"
	"strings"
)

type Issue struct {
	ID    int64   `json:"id"`
	Title string  `json:"title"`
	Body  *string `json:"body"` // nullable - can be nil
}

func PrintIssue(issue *Issue) {
	fmt.Printf("Issue #%d: %s\n", issue.ID, issue.Title)

	// BUG: No nil check before dereferencing
	bodyPreview := (*issue.Body)[:100] // PANIC if Body is nil
	fmt.Printf("Preview: %s\n", bodyPreview)

	// Also broken:
	if strings.Contains(*issue.Body, "bug") { // PANIC if Body is nil
		fmt.Println("This is a bug report")
	}
}

// Correct implementation:
func PrintIssueSafe(issue *Issue) {
	fmt.Printf("Issue #%d: %s\n", issue.ID, issue.Title)

	if issue.Body != nil {
		bodyPreview := (*issue.Body)[:100]
		fmt.Printf("Preview: %s\n", bodyPreview)
	} else {
		fmt.Println("No description provided")
	}
}

// GitHub response examples:
// {"id": 1, "title": "Bug", "body": "Detailed description"}  // Body is non-nil
// {"id": 2, "title": "Feature", "body": null}                // Body is nil
