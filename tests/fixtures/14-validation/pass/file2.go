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
