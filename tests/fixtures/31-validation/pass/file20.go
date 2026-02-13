// service/notification_service.go — Type assertion bypassing interface.
package service

import (
	"context"
	"super-lint-test/internal/domain"
	"super-lint-test/internal/repository/postgres"  // <-- VIOLATION: Importing concrete type
	"super-lint-test/pkg/logging"
)

// NotificationService sends notifications with internal optimizations.
type NotificationService struct {
	userRepo UserRepository
	logger   logging.Logger
}

// NewNotificationService creates notification service.
func NewNotificationService(repo UserRepository, logger logging.Logger) *NotificationService {
	return &NotificationService{
		userRepo: repo,
		logger:   logger,
	}
}

// SendBulkNotification casts interface to concrete type.
func (s *NotificationService) SendBulkNotification(ctx context.Context, userIDs []string) error {
	// Type assertion to access internal methods - violates interface contract
	if concreteRepo, ok := s.userRepo.(*postgres.UserRepo); ok {  // <-- VIOLATION: Type assertion
		// Accessing internal method not in UserRepository interface
		users := concreteRepo.BatchGetOptimized(ctx, userIDs)  // <-- VIOLATION: Internal method

		for _, user := range users {
			s.logger.Info("sending notification", "user_id", user.ID)
		}
		return nil
	}

	// Fallback to interface method (slow path)
	for _, id := range userIDs {
		user, err := s.userRepo.GetByID(ctx, id)
		if err != nil {
			continue
		}
		s.logger.Info("sending notification", "user_id", user.ID)
	}

	return nil
}
