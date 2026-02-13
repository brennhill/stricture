// service/validation.go — Part 1 of 2-node cycle.
package service

import (
	"context"
	"super-lint-test/internal/domain"
	"super-lint-test/internal/service/util"  // <-- VIOLATION: service imports util
)

type ValidationService struct {
	formatter *util.UserFormatter
}

func (s *ValidationService) ValidateAndFormat(ctx context.Context, user *domain.User) (string, error) {
	if err := user.Validate(); err != nil {
		return "", err
	}
	// Service uses util for formatting
	return s.formatter.Format(user), nil  // <-- Creates dependency on util
}
