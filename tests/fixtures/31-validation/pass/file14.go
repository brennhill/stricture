// service/util/formatter.go — Part 2 of 2-node cycle.
package util

import (
	"super-lint-test/internal/domain"
	"super-lint-test/internal/service"  // <-- VIOLATION: util imports service (completes cycle)
)

type UserFormatter struct {
	validator *service.ValidationService  // <-- VIOLATION: util depends on service
}

func (f *UserFormatter) Format(user *domain.User) string {
	// Util calls back into service (circular dependency)
	f.validator.ValidateAndFormat(nil, user)  // <-- Creates cycle
	return user.Email
}
