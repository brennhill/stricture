// github_client_b05.go — Missing required field validation
package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type CreateIssueRequest struct {
	Title string  `json:"title"`           // BUG: no validation that this is required
	Body  *string `json:"body,omitempty"`
}

func (c *Client) CreateIssue(ctx context.Context, owner, repo string, req *CreateIssueRequest) (*Issue, error) {
	// BUG: No validation that title is non-empty
	url := fmt.Sprintf("%s/repos/%s/%s/issues", c.baseURL, owner, repo)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// GitHub returns 422 Unprocessable Entity if title is empty
	// but code doesn't validate before sending

	var issue Issue
	json.NewDecoder(resp.Body).Decode(&issue)
	return &issue, nil
}
