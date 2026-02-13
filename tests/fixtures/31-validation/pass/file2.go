// service/interfaces.go — Repository interfaces defined by service layer.
package service

import (
	"context"
	"super-lint-test/internal/domain"
)

// UserRepository defines the interface for user data access.
// The service layer defines this interface, and the repository layer implements it.
// This is dependency inversion — high-level module (service) defines the interface.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*domain.User, error)
}

// SessionRepository defines the interface for session management.
type SessionRepository interface {
	Create(ctx context.Context, userID string, token string) error
	GetByToken(ctx context.Context, token string) (userID string, err error)
	Delete(ctx context.Context, token string) error
	DeleteAllForUser(ctx context.Context, userID string) error
}
