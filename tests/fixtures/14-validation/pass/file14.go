// github_client_b14.go — Missing pagination handling
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) ListAllIssues(ctx context.Context, owner, repo string) ([]*Issue, error) {
	var allIssues []*Issue
	page := 1

	for {
		url := fmt.Sprintf("%s/repos/%s/%s/issues?page=%d&per_page=100", c.baseURL, owner, repo, page)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		// BUG: Doesn't check Link header for rel="next"
		// Assumes page-based pagination but doesn't know when to stop

		var issues []*Issue
		if err := json.Unmarshal(body, &issues); err != nil {
			return nil, err
		}

		allIssues = append(allIssues, issues...)

		// BUG: Breaks after first page, missing pagination logic
		if len(issues) < 100 {
			break // Wrong: assumes < 100 means last page, but could be exactly 100 items total
		}

		page++
	}

	return allIssues, nil
}

// GitHub Link header format:
// Link: <https://api.github.com/repos/owner/repo/issues?page=2>; rel="next",
//       <https://api.github.com/repos/owner/repo/issues?page=5>; rel="last"
//
// Correct implementation must:
// 1. Parse Link header
// 2. Check for rel="next"
// 3. Use that URL for next request
// 4. Stop when rel="next" is absent
