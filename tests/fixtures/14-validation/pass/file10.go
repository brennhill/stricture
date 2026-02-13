// github_client_b10.go — Missing format validation
package github

import (
	"context"
	"fmt"
)

func (c *Client) GetRepository(ctx context.Context, owner, repo string) (*Repository, error) {
	// BUG: No validation of owner/repo format
	// GitHub usernames/repos must match: ^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$

	if owner == "" {
		return nil, fmt.Errorf("owner cannot be empty")
	}
	// BUG: Accepts invalid formats like "-invalid", "user-", "user..name"

	url := fmt.Sprintf("%s/repos/%s/%s", c.baseURL, owner, repo)
	// ... rest of implementation

	return nil, nil
}

// Invalid inputs that should be rejected:
// - owner="-invalid" (starts with hyphen)
// - owner="user-" (ends with hyphen)
// - owner="user@name" (invalid character)
// - owner="user..name" (consecutive dots, if applicable)
// - repo="" (empty string, already caught)
// - repo="repo name" (contains space)

// GitHub login format rules:
// - Alphanumeric and hyphens only
// - Cannot start or end with hyphen
// - Cannot have consecutive hyphens
// - Max 39 characters
