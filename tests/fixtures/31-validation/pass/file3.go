// service/user_service.go — User service implementing business logic.
package service

import (
	"context"
	"errors"
	"super-lint-test/internal/domain"
	"super-lint-test/pkg/logging"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrUserNotFound is returned when a user is not found.
	ErrUserNotFound = errors.New("user not found")
	// ErrDuplicateEmail is returned when email already exists.
	ErrDuplicateEmail = errors.New("email already exists")
)

// UserService implements user business logic.
type UserService struct {
	repo   UserRepository
	logger logging.Logger
}

// NewUserService creates a new user service.
func NewUserService(repo UserRepository, logger logging.Logger) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger,
	}
}

// CreateUser creates a new user with validation.
func (s *UserService) CreateUser(ctx context.Context, email, username string) (*domain.User, error) {
	// Check for duplicate email
	existing, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, ErrDuplicateEmail
	}

	user := &domain.User{
		ID:        uuid.New().String(),
		Email:     email,
		Username:  username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}

	// Validate domain rules
	if err := user.Validate(); err != nil {
		return nil, err
	}

	// Persist
	if err := s.repo.Create(ctx, user); err != nil {
		s.logger.Error("failed to create user", "error", err)
		return nil, err
	}

	s.logger.Info("user created", "user_id", user.ID, "email", user.Email)
	return user, nil
}

// GetUser retrieves a user by ID.
func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// ActivateUser activates a user account.
func (s *UserService) ActivateUser(ctx context.Context, id string) error {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return err
	}

	user.Activate()

	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.Error("failed to activate user", "user_id", id, "error", err)
		return err
	}

	s.logger.Info("user activated", "user_id", id)
	return nil
}

// DeactivateUser deactivates a user account.
func (s *UserService) DeactivateUser(ctx context.Context, id string) error {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return err
	}

	user.Deactivate()

	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.Error("failed to deactivate user", "user_id", id, "error", err)
		return err
	}

	s.logger.Info("user deactivated", "user_id", id)
	return nil
}
