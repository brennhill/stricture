// github_client_b09.go — Missing range validation
package github

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

type ListIssuesOptions struct {
	State   string
	PerPage int // BUG: No validation that this must be [1, 100]
	Page    int
}

func (c *Client) ListIssues(ctx context.Context, owner, repo string, opts *ListIssuesOptions) ([]*Issue, error) {
	// BUG: No validation of per_page range
	url := fmt.Sprintf("%s/repos/%s/%s/issues", c.baseURL, owner, repo)
	reqURL, _ := url.Parse(url)

	query := reqURL.Query()
	if opts != nil {
		if opts.PerPage > 0 { // BUG: Allows 999, should reject > 100
			query.Set("per_page", strconv.Itoa(opts.PerPage))
		}
		// Also doesn't reject PerPage == 0, which is invalid
	}
	reqURL.RawQuery = query.Encode()

	// GitHub API will reject per_page=0 or per_page=999 with 422
	// but code doesn't validate before sending request

	return nil, nil
}

// Manifest specifies: per_page: [1, 100]
// Valid: ListIssues(ctx, "owner", "repo", &ListIssuesOptions{PerPage: 50})
// Invalid: ListIssues(ctx, "owner", "repo", &ListIssuesOptions{PerPage: 0})
// Invalid: ListIssues(ctx, "owner", "repo", &ListIssuesOptions{PerPage: 999})
