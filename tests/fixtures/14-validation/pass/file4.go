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
