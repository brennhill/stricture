// github_client_b06.go — Missing response field
package github

import "time"

type Repository struct {
	ID          int64     `json:"id"`
	// BUG: Missing node_id field - GitHub returns this in all responses
	Name        string    `json:"name"`
	FullName    string    `json:"full_name"`
	Owner       User      `json:"owner"`
	Private     bool      `json:"private"`
	HTMLURL     string    `json:"html_url"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type User struct {
	Login  string `json:"login"`
	ID     int64  `json:"id"`
	// BUG: Also missing node_id here
	Type   string `json:"type"`
}

// When unmarshaling GitHub responses, node_id field is silently dropped
// This violates response shape conformance - all documented fields should be captured
