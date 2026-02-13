// service/auth_service.go — Service importing handler types.
package service

import (
	"context"
	"errors"
	"super-lint-test/internal/domain"
	"super-lint-test/internal/handler"  // <-- VIOLATION: Service imports handler (reverse flow)
	"super-lint-test/pkg/logging"
	"time"

	"github.com/google/uuid"
)

// AuthService handles authentication with handler types.
type AuthService struct {
	userRepo UserRepository
	logger   logging.Logger
}

// NewAuthService creates auth service.
func NewAuthService(repo UserRepository, logger logging.Logger) *AuthService {
	return &AuthService{
		userRepo: repo,
		logger:   logger,
	}
}

// ValidateCredentials checks credentials and returns handler response type.
func (s *AuthService) ValidateCredentials(ctx context.Context, email, password string) (*handler.UserResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Service returning handler's response type - architectural violation
	resp := &handler.UserResponse{  // <-- VIOLATION: Using handler type in service
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		IsActive: user.IsActive,
	}

	return resp, nil
}
