// internal/billing/invoice_service.go — Accessing user module internals.
package billing

import (
	"context"
	"super-lint-test/internal/domain"
	"super-lint-test/internal/users/repository/postgres"  // <-- VIOLATION: Importing internal implementation
	"super-lint-test/pkg/logging"
)

// InvoiceService generates invoices with user data.
type InvoiceService struct {
	userRepo *postgres.UserRepo  // <-- VIOLATION: Billing module depends on users module internals
	logger   logging.Logger
}

// NewInvoiceService creates invoice service.
func NewInvoiceService(userRepo *postgres.UserRepo, logger logging.Logger) *InvoiceService {
	return &InvoiceService{
		userRepo: userRepo,
		logger:   logger,
	}
}

// GenerateInvoice creates invoice by directly accessing user repository.
func (s *InvoiceService) GenerateInvoice(ctx context.Context, userID string) (*domain.Invoice, error) {
	// Billing module reaching into users module's repository internals
	user, err := s.userRepo.GetByID(ctx, userID)  // <-- VIOLATION: Bypassing users module boundary
	if err != nil {
		s.logger.Error("failed to get user for invoice", "error", err)
		return nil, err
	}

	invoice := &domain.Invoice{
		UserID:    user.ID,
		UserEmail: user.Email,
		Amount:    1000,
	}

	return invoice, nil
}
