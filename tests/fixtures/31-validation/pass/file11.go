// service/order_service_cycle.go — Part 2 of cycle.
package service

import (
	"context"
	"super-lint-test/internal/repository"
	"super-lint-test/pkg/logging"
)

type OrderService struct {
	repo   repository.OrderRepository
	logger logging.Logger
}

func (s *OrderService) CreateOrder(ctx context.Context) error {
	// Service calls repository (normal flow)
	return s.repo.Create(ctx, nil)
}
