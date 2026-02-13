// github_client_b08.go — Incomplete enum handling
package github

import (
	"context"
	"fmt"
)

type ListIssuesOptions struct {
	State   string // "open", "closed", or "all"
	PerPage int
	Page    int
}

func (c *Client) ListIssues(ctx context.Context, owner, repo string, opts *ListIssuesOptions) ([]*Issue, error) {
	// BUG: Only validates "open" and "closed", missing "all"
	if opts != nil && opts.State != "" {
		if opts.State != "open" && opts.State != "closed" {
			return nil, fmt.Errorf("invalid state: %s", opts.State)
		}
	}

	// GitHub API accepts three values:
	// - "open": only open issues
	// - "closed": only closed issues
	// - "all": both open and closed
	//
	// Code rejects "all" as invalid, breaking valid use case

	url := fmt.Sprintf("%s/repos/%s/%s/issues", c.baseURL, owner, repo)
	// ... rest of implementation
	return nil, nil
}
