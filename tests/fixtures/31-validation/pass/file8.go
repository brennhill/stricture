// internal/users/service/user_billing_sync.go — Cross-module internal access.
package service

import (
	"context"
	"super-lint-test/internal/domain"
	"super-lint-test/internal/orders/repository/postgres"  // <-- VIOLATION: Cross-module internal import
	"super-lint-test/pkg/logging"
)

// UserBillingSyncService syncs user data with orders.
type UserBillingSyncService struct {
	userRepo   UserRepository
	orderRepo  *postgres.OrderRepo  // <-- VIOLATION: Direct access to orders module internals
	logger     logging.Logger
}

// NewUserBillingSyncService creates sync service.
func NewUserBillingSyncService(
	userRepo UserRepository,
	orderRepo *postgres.OrderRepo,
	logger logging.Logger,
) *UserBillingSyncService {
	return &UserBillingSyncService{
		userRepo:  userRepo,
		orderRepo: orderRepo,
		logger:    logger,
	}
}

// SyncUserOrders directly queries order repository from user module.
func (s *UserBillingSyncService) SyncUserOrders(ctx context.Context, userID string) error {
	// Users module reaching into orders module's repository layer
	orders, err := s.orderRepo.GetByUserID(ctx, userID)  // <-- VIOLATION
	if err != nil {
		s.logger.Error("failed to get orders", "error", err)
		return err
	}

	s.logger.Info("synced orders", "user_id", userID, "order_count", len(orders))
	return nil
}
