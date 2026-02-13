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
