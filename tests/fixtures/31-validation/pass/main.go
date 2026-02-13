// domain/user.go — User domain entity with business rules.
package domain

import (
	"errors"
	"time"
)

// User represents a user in the system.
type User struct {
	ID        string
	Email     string
	Username  string
	CreatedAt time.Time
	UpdatedAt time.Time
	IsActive  bool
}

// Validate checks if the user entity is valid.
func (u *User) Validate() error {
	if u.Email == "" {
		return errors.New("email is required")
	}
	if u.Username == "" {
		return errors.New("username is required")
	}
	if len(u.Username) < 3 {
		return errors.New("username must be at least 3 characters")
	}
	return nil
}

// Activate marks the user as active.
func (u *User) Activate() {
	u.IsActive = true
	u.UpdatedAt = time.Now()
}

// Deactivate marks the user as inactive.
func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now()
}
