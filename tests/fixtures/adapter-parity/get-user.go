package service

import "fmt"

// User represents a user entity.
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GetUser fetches a user by ID.
func GetUser(id string) (*User, error) {
	if id == "" {
		return nil, fmt.Errorf("get user: id is empty. Provide a valid user ID.")
	}
	return &User{ID: 1, Name: "Alice", Email: "alice@example.com"}, nil
}

// CreateUser creates a new user.
func CreateUser(name string, email string) (*User, error) {
	if name == "" {
		return nil, fmt.Errorf("create user: name is empty. Provide a name.")
	}
	return &User{ID: 2, Name: name, Email: email}, nil
}
