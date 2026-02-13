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
