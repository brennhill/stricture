// github_client_b11.go — Wrong timestamp type
package github

import "encoding/json"

type Issue struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"` // BUG: Should be time.Time
	UpdatedAt string `json:"updated_at"` // BUG: Should be time.Time
	ClosedAt  string `json:"closed_at"`  // BUG: Should be *time.Time (nullable)
}

type Repository struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"` // BUG: Should be time.Time
	UpdatedAt string `json:"updated_at"` // BUG: Should be time.Time
	PushedAt  string `json:"pushed_at"`  // BUG: Should be *time.Time (nullable)
}

// Example GitHub response:
// {
//   "created_at": "2011-01-26T19:01:12Z",
//   "updated_at": "2024-12-01T10:30:45Z",
//   "closed_at": null
// }
//
// By storing as string:
// - Cannot use time.Time methods (Before, After, Add, etc.)
// - Cannot parse timezone correctly
// - Cannot perform time arithmetic
// - Loses type safety for time operations

func ExampleBrokenTimeHandling() {
	issueJSON := `{"id": 1, "title": "Bug", "created_at": "2024-01-01T00:00:00Z"}`
	var issue Issue
	json.Unmarshal([]byte(issueJSON), &issue)

	// BUG: Can't do time comparisons
	// if issue.CreatedAt.Before(time.Now()) { ... } // won't compile
}
