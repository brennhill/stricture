// repository/order_repo_cycle.go — Part 3 of cycle (creates circular dependency).
package repository

import (
	"context"
	"super-lint-test/internal/domain"
	"super-lint-test/internal/handler"  // <-- VIOLATION: Repository imports handler (completes cycle)
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
}

type OrderRepoImpl struct {
	responseFactory *handler.ResponseFactory  // <-- VIOLATION: Repo depends on handler
}

func (r *OrderRepoImpl) Create(ctx context.Context, order *domain.Order) error {
	// Repository uses handler code (architectural violation)
	r.responseFactory.Format(order)  // <-- VIOLATION
	return nil
}
