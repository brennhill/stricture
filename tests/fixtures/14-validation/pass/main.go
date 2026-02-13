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
