// github_client_b15.go — Missing concurrency control
package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type UpdateIssueRequest struct {
	Title *string `json:"title,omitempty"`
	State *string `json:"state,omitempty"`
}

func (c *Client) UpdateIssue(ctx context.Context, owner, repo string, number int64, req *UpdateIssueRequest) (*Issue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d", c.baseURL, owner, repo, number)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// BUG: No If-Match header with ETag
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	httpReq.Header.Set("Content-Type", "application/json")
	// BUG: Should set If-Match header with ETag from previous GET

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var issue Issue
	json.NewDecoder(resp.Body).Decode(&issue)
	return &issue, nil
}

// Race condition scenario:
// 1. Client A fetches issue #123 (state="open", ETag: "abc123")
// 2. Client B fetches issue #123 (state="open", ETag: "abc123")
// 3. Client A updates issue #123 to state="closed" (ETag now "def456")
// 4. Client B updates issue #123 to state="open" (overwrites A's change)
//
// With If-Match:
// 1. Client A fetches issue (ETag: "abc123")
// 2. Client B fetches issue (ETag: "abc123")
// 3. Client A updates with If-Match: "abc123" → succeeds, new ETag "def456"
// 4. Client B updates with If-Match: "abc123" → GitHub returns 412 Precondition Failed
// 5. Client B must re-fetch and retry

// Correct implementation:
type Issue struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	ETag  string `json:"-"` // stored from ETag header, not in JSON body
}

func (c *Client) UpdateIssueSafe(ctx context.Context, issue *Issue, req *UpdateIssueRequest) (*Issue, error) {
	// ... create request

	if issue.ETag != "" {
		httpReq.Header.Set("If-Match", issue.ETag) // prevents lost updates
	}

	// ... send request
	// If 412 Precondition Failed, caller must re-fetch and retry
	return nil, nil
}
