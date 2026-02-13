# 14 — GitHub REST API (Go)

**Why included:** Go encoding/json, struct embedding, *string nullable fields, Link header pagination, context-aware requests.

## Manifest Fragment

```yaml
repos_crud:
  endpoints:
    - GET /repos/:owner/:repo
    - POST /orgs/:org/repos
    - PATCH /repos/:owner/:repo
  status_codes: [200, 201, 404, 422]

issues_list_create:
  endpoints:
    - GET /repos/:owner/:repo/issues
    - POST /repos/:owner/:repo/issues
    - PATCH /repos/:owner/:repo/issues/:number
  pagination: link_header
  per_page: [1, 100]
  status_codes: [200, 201, 404, 422]
```

---

## PERFECT

**File:** `perfect/github_client.go`

```go
// github_client.go — GitHub REST API client with proper error handling and pagination.
package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Repository represents a GitHub repository.
type Repository struct {
	ID          int64      `json:"id"`
	NodeID      string     `json:"node_id"`
	Name        string     `json:"name"`
	FullName    string     `json:"full_name"`
	Owner       User       `json:"owner"`
	Private     bool       `json:"private"`
	HTMLURL     string     `json:"html_url"`
	Description *string    `json:"description"`
	Fork        bool       `json:"fork"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	PushedAt    *time.Time `json:"pushed_at"`
	Size        int        `json:"size"`
	StargazersCount int    `json:"stargazers_count"`
	Language    *string    `json:"language"`
	HasIssues   bool       `json:"has_issues"`
	Archived    bool       `json:"archived"`
	Disabled    bool       `json:"disabled"`
	Visibility  string     `json:"visibility"`
}

// Issue represents a GitHub issue.
type Issue struct {
	ID          int64      `json:"id"`
	NodeID      string     `json:"node_id"`
	Number      int64      `json:"number"`
	Title       string     `json:"title"`
	User        User       `json:"user"`
	State       string     `json:"state"` // "open" or "closed"
	Locked      bool       `json:"locked"`
	Assignee    *User      `json:"assignee"`
	Comments    int        `json:"comments"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ClosedAt    *time.Time `json:"closed_at"`
	Body        *string    `json:"body"`
	HTMLURL     string     `json:"html_url"`
}

// User represents a GitHub user.
type User struct {
	Login     string `json:"login"`
	ID        int64  `json:"id"`
	NodeID    string `json:"node_id"`
	AvatarURL string `json:"avatar_url"`
	Type      string `json:"type"`
	SiteAdmin bool   `json:"site_admin"`
}

// CreateRepoRequest represents a request to create a repository.
type CreateRepoRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Private     bool    `json:"private"`
	HasIssues   bool    `json:"has_issues"`
	AutoInit    bool    `json:"auto_init"`
}

// UpdateRepoRequest represents a request to update a repository.
type UpdateRepoRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Private     *bool   `json:"private,omitempty"`
	HasIssues   *bool   `json:"has_issues,omitempty"`
	Archived    *bool   `json:"archived,omitempty"`
}

// CreateIssueRequest represents a request to create an issue.
type CreateIssueRequest struct {
	Title string  `json:"title"`
	Body  *string `json:"body,omitempty"`
}

// UpdateIssueRequest represents a request to update an issue.
type UpdateIssueRequest struct {
	Title *string `json:"title,omitempty"`
	Body  *string `json:"body,omitempty"`
	State *string `json:"state,omitempty"` // "open" or "closed"
}

// ListIssuesOptions represents pagination options for listing issues.
type ListIssuesOptions struct {
	State   string // "open", "closed", "all"
	PerPage int    // [1, 100]
	Page    int
}

// PaginationInfo contains pagination metadata from Link headers.
type PaginationInfo struct {
	NextPage int
	LastPage int
	HasNext  bool
}

// ErrorResponse represents a GitHub API error response.
type ErrorResponse struct {
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url"`
}

// Client is a GitHub REST API client.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new GitHub API client.
func NewClient(token string) *Client {
	return &Client{
		baseURL:    "https://api.github.com",
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// ValidateRepoName validates owner and repo names.
func ValidateRepoName(owner, repo string) error {
	if owner == "" {
		return fmt.Errorf("owner cannot be empty")
	}
	if repo == "" {
		return fmt.Errorf("repo cannot be empty")
	}
	if strings.Contains(owner, "/") || strings.Contains(repo, "/") {
		return fmt.Errorf("owner and repo cannot contain slashes")
	}
	return nil
}

// ValidatePerPage validates per_page parameter [1, 100].
func ValidatePerPage(perPage int) error {
	if perPage < 1 || perPage > 100 {
		return fmt.Errorf("per_page must be between 1 and 100, got %d", perPage)
	}
	return nil
}

// GetRepository fetches a repository by owner and name.
func (c *Client) GetRepository(ctx context.Context, owner, repo string) (*Repository, error) {
	if err := ValidateRepoName(owner, repo); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/repos/%s/%s", c.baseURL, owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var repository Repository
	if err := json.Unmarshal(body, &repository); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &repository, nil
}

// CreateRepository creates a new repository in an organization.
func (c *Client) CreateRepository(ctx context.Context, org string, req *CreateRepoRequest) (*Repository, error) {
	if org == "" {
		return nil, fmt.Errorf("org cannot be empty")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("repository name cannot be empty")
	}

	url := fmt.Sprintf("%s/orgs/%s/repos", c.baseURL, org)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	httpReq.Header.Set("Accept", "application/vnd.github+json")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, c.handleErrorResponse(resp)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var repository Repository
	if err := json.Unmarshal(respBody, &repository); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &repository, nil
}

// UpdateRepository updates an existing repository.
func (c *Client) UpdateRepository(ctx context.Context, owner, repo string, req *UpdateRepoRequest) (*Repository, error) {
	if err := ValidateRepoName(owner, repo); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/repos/%s/%s", c.baseURL, owner, repo)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	httpReq.Header.Set("Accept", "application/vnd.github+json")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var repository Repository
	if err := json.Unmarshal(respBody, &repository); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &repository, nil
}

// ListIssues lists issues for a repository with pagination.
func (c *Client) ListIssues(ctx context.Context, owner, repo string, opts *ListIssuesOptions) ([]*Issue, *PaginationInfo, error) {
	if err := ValidateRepoName(owner, repo); err != nil {
		return nil, nil, err
	}

	if opts != nil && opts.PerPage > 0 {
		if err := ValidatePerPage(opts.PerPage); err != nil {
			return nil, nil, err
		}
	}

	url := fmt.Sprintf("%s/repos/%s/%s/issues", c.baseURL, owner, repo)
	reqURL, err := url.Parse(url)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	query := reqURL.Query()
	if opts != nil {
		if opts.State != "" {
			query.Set("state", opts.State)
		}
		if opts.PerPage > 0 {
			query.Set("per_page", strconv.Itoa(opts.PerPage))
		}
		if opts.Page > 0 {
			query.Set("page", strconv.Itoa(opts.Page))
		}
	}
	reqURL.RawQuery = query.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	httpReq.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, c.handleErrorResponse(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var issues []*Issue
	if err := json.Unmarshal(body, &issues); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	pagination := c.parseLinkHeader(resp.Header.Get("Link"))

	return issues, pagination, nil
}

```

---

## BUGS

### B01 — No Error Handling (TQ-error-path-coverage)

**Bug:** Missing `if err != nil` check after `http.Get` call, causing potential nil pointer dereference on resp.Body.
**Expected violation:** `TQ-error-path-coverage` on line 8

```go
// github_client_b01.go — Missing error handling on HTTP request
package github

import (
	"encoding/json"
	"io"
	"net/http"
)

type Repository struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

func GetRepository(owner, repo string) (*Repository, error) {
	url := "https://api.github.com/repos/" + owner + "/" + repo
	resp, _ := http.Get(url) // BUG: ignoring error
	defer resp.Body.Close()   // crashes if resp is nil

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var repository Repository
	if err := json.Unmarshal(body, &repository); err != nil {
		return nil, err
	}

	return &repository, nil
}
```

**Why Stricture catches this:** TQ-error-path-coverage requires all error-returning function calls to have corresponding error checks. The `http.Get` call returns `(*http.Response, error)` but the error is discarded with `_`, violating error handling coverage.

### B02 — No Status Code Check (CTR-status-code-handling)

**Bug:** Missing HTTP status code validation after GitHub API call, causing JSON unmarshal errors on 404/403/500 responses.
**Expected violation:** `CTR-status-code-handling` on line 24

```go
// github_client_b02.go — Missing status code check
package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Repository struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

func GetRepository(owner, repo string) (*Repository, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// BUG: No status code check - will fail to unmarshal error responses
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var repository Repository
	if err := json.Unmarshal(body, &repository); err != nil {
		return nil, err // misleading error on 404/403/500
	}

	return &repository, nil
}
```

**Why Stricture catches this:** CTR-status-code-handling requires status code validation after HTTP requests. The manifest specifies `status_codes: [200, 201, 404, 422]`, but the code doesn't check `resp.StatusCode` before processing the response body.

### B03 — Shallow Test Assertions (TQ-no-shallow-assertions)

**Bug:** Test only checks `assert.NotNil(t, repos)` instead of validating repository field values.
**Expected violation:** `TQ-no-shallow-assertions` on line 20

```go
// github_client_b03_test.go — Shallow assertions
package github

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestListRepositories(t *testing.T) {
	client := NewClient("fake-token")

	repos, err := client.ListRepositories("octocat")

	assert.NoError(t, err)
	assert.NotNil(t, repos) // BUG: shallow assertion

	// Missing assertions:
	// - assert.Greater(t, len(repos), 0)
	// - assert.NotEmpty(t, repos[0].Name)
	// - assert.NotEmpty(t, repos[0].Owner.Login)
	// - assert.True(t, repos[0].ID > 0)
	// - assert.NotEmpty(t, repos[0].NodeID)
}

func TestCreateRepository(t *testing.T) {
	client := NewClient("fake-token")

	req := &CreateRepoRequest{
		Name:    "test-repo",
		Private: true,
	}

	repo, err := client.CreateRepository("myorg", req)

	assert.NoError(t, err)
	assert.NotNil(t, repo) // BUG: shallow assertion

	// Missing assertions:
	// - assert.Equal(t, "test-repo", repo.Name)
	// - assert.True(t, repo.Private)
	// - assert.Greater(t, repo.ID, int64(0))
}
```

**Why Stricture catches this:** TQ-no-shallow-assertions detects assertions that only check for nil/non-nil without validating actual content. Tests must verify struct fields match expected values, not just existence.

### B04 — Missing Negative Tests (TQ-negative-cases)

**Bug:** Test suite only covers successful 200/201 responses, missing 404/403/422/rate-limit failure cases.
**Expected violation:** `TQ-negative-cases` on test file

```go
// github_client_b04_test.go — Missing negative test cases
package github

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

// BUG: Only tests happy path
func TestGetRepository(t *testing.T) {
	client := NewClient("fake-token")

	repo, err := client.GetRepository("octocat", "Hello-World")

	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, "Hello-World", repo.Name)
}

// Missing negative tests:
// - TestGetRepository_NotFound (404)
// - TestGetRepository_Forbidden (403)
// - TestGetRepository_Unauthorized (401)
// - TestGetRepository_RateLimited (429)
// - TestGetRepository_InvalidOwner (400)
// - TestGetRepository_NetworkError
// - TestGetRepository_MalformedJSON

func TestCreateRepository(t *testing.T) {
	client := NewClient("fake-token")

	req := &CreateRepoRequest{
		Name:    "test-repo",
		Private: true,
	}

	repo, err := client.CreateRepository("myorg", req)

	assert.NoError(t, err)
	assert.NotNil(t, repo)
}

// Missing negative tests:
// - TestCreateRepository_EmptyName (422)
// - TestCreateRepository_DuplicateName (422)
// - TestCreateRepository_Forbidden (403)
```

**Why Stricture catches this:** TQ-negative-cases requires test coverage for all documented error states. The manifest specifies `status_codes: [200, 201, 404, 422]`, but tests only cover 200/201 success cases.

### B05 — Request Missing Required Fields (CTR-request-shape)

**Bug:** CreateIssueRequest allows empty `title` field, violating GitHub API requirement that title is mandatory.
**Expected violation:** `CTR-request-shape` on line 9

```go
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
```

**Why Stricture catches this:** CTR-request-shape validates that request structs enforce required fields per API manifest. GitHub's create issue endpoint requires `title`, but the code allows empty strings.

### B06 — Response Type Mismatch (CTR-response-shape)

**Bug:** Go struct missing `node_id` field that GitHub API returns in all repository responses.
**Expected violation:** `CTR-response-shape` on line 6

```go
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
```

**Why Stricture catches this:** CTR-response-shape validates that response structs capture all documented API fields. GitHub's API specification includes `node_id` in repository and user objects, but the struct omits it.

### B07 — Wrong Field Types (CTR-manifest-conformance)

**Bug:** Repository.ID field is string instead of int64, causing unmarshal errors.
**Expected violation:** `CTR-manifest-conformance` on line 6

```go
// github_client_b07.go — Wrong field type
package github

import "time"

type Repository struct {
	ID        string    `json:"id"` // BUG: Should be int64, GitHub returns numeric ID
	NodeID    string    `json:"node_id"`
	Name      string    `json:"name"`
	FullName  string    `json:"full_name"`
	Owner     User      `json:"owner"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	Login  string `json:"login"`
	ID     string `json:"id"` // BUG: Also wrong - should be int64
	NodeID string `json:"node_id"`
}

type Issue struct {
	ID     string `json:"id"`     // BUG: Wrong type
	Number string `json:"number"` // BUG: Wrong type - should be int64
	Title  string `json:"title"`
}

// Example GitHub response:
// {
//   "id": 1296269,           // numeric
//   "node_id": "MDEwOlJl...", // string
//   "number": 1347           // numeric
// }
//
// json.Unmarshal will fail when trying to unmarshal 1296269 into a string field
```

**Why Stricture catches this:** CTR-manifest-conformance validates that struct field types match API specification. GitHub documents `id` and `number` as integers, but the code declares them as strings.

### B08 — Incomplete Enum Handling (CTR-strictness-parity)

**Bug:** Issue state handling only checks "open"/"closed" but GitHub API also accepts "all" for filtering.
**Expected violation:** `CTR-strictness-parity` on line 18

```go
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
```

**Why Stricture catches this:** CTR-strictness-parity validates that enum validation matches API specification. GitHub's state parameter accepts three values ("open", "closed", "all"), but the validation only allows two.

### B09 — Missing Range Validation (CTR-strictness-parity)

**Bug:** per_page parameter accepts 0 or 999, violating GitHub's documented range of [1, 100].
**Expected violation:** `CTR-strictness-parity` on line 15

```go
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
```

**Why Stricture catches this:** CTR-strictness-parity validates that parameter validation matches API constraints. The manifest specifies `per_page: [1, 100]`, but the code doesn't enforce this range.

### B10 — Format Not Validated (CTR-strictness-parity)

**Bug:** No regex validation for GitHub login format (alphanumeric + hyphens, no leading/trailing hyphens).
**Expected violation:** `CTR-strictness-parity` on line 14

```go
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
```

**Why Stricture catches this:** CTR-strictness-parity validates that string parameters match documented format constraints. GitHub has specific format requirements for usernames and repository names that aren't enforced.

### B11 — Precision Loss on Timestamps (CTR-strictness-parity)

**Bug:** Timestamp parsed as string instead of time.Time, losing timezone and comparison capabilities.
**Expected violation:** `CTR-strictness-parity` on line 9

```go
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
```

**Why Stricture catches this:** CTR-strictness-parity validates that timestamp fields use proper time types. GitHub returns ISO8601 timestamps that should be parsed into time.Time for proper handling.

### B12 — Nullable Field Crashes (CTR-response-shape)

**Bug:** Dereferences *string body field without nil check, causing panic on issues without descriptions.
**Expected violation:** `CTR-response-shape` on line 18

```go
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
```

**Why Stricture catches this:** CTR-response-shape validates that nullable fields (*string, *int, *time.Time) are checked for nil before dereferencing. The API returns null for optional fields.

### B13 — Missing Webhook Verification (CTR-request-shape)

**Bug:** No X-Hub-Signature-256 verification for webhook payloads, allowing spoofed requests.
**Expected violation:** `CTR-request-shape` on line 15

```go
// github_client_b13.go — Missing webhook signature verification
package github

import (
	"encoding/json"
	"io"
	"net/http"
)

type WebhookPayload struct {
	Action     string     `json:"action"`
	Issue      *Issue     `json:"issue,omitempty"`
	Repository Repository `json:"repository"`
}

func HandleWebhook(w http.ResponseWriter, r *http.Request) {
	// BUG: No signature verification - anyone can send fake webhooks

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// BUG: Processing webhook without verifying it came from GitHub
	// Attacker can send POST to webhook endpoint with fake data

	w.WriteHeader(http.StatusOK)
}

// Correct implementation requires:
// 1. Read X-Hub-Signature-256 header
// 2. Compute HMAC-SHA256 of body using webhook secret
// 3. Compare computed signature with header value
// 4. Reject if signatures don't match

// Example attack:
// curl -X POST https://myapp.com/webhook \
//   -H "Content-Type: application/json" \
//   -d '{"action":"closed","issue":{"id":999,"title":"Fake"}}'
//
// Without signature verification, this fake webhook is accepted
```

**Why Stricture catches this:** CTR-request-shape validates that webhook handlers verify request signatures. GitHub requires HMAC-SHA256 verification of webhook payloads to prevent spoofing.

### B14 — Pagination Terminated Early (CTR-response-shape)

**Bug:** Ignores Link header rel="next", preventing full pagination through all results.
**Expected violation:** `CTR-response-shape` on line 38

```go
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
```

**Why Stricture catches this:** CTR-response-shape validates that paginated endpoints correctly handle Link headers. The manifest specifies `pagination: link_header`, but the code doesn't parse or follow these links.

### B15 — Race Condition (CTR-request-shape)

**Bug:** No If-Match/ETag on issue update, allowing lost updates when multiple clients modify simultaneously.
**Expected violation:** `CTR-request-shape` on line 24

```go
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
```

**Why Stricture catches this:** CTR-request-shape validates that mutation endpoints use optimistic locking (ETag/If-Match) to prevent race conditions. GitHub supports ETags for issue updates, but the code doesn't use them.
