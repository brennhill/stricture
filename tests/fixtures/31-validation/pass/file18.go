// repository/postgres/user_repo_violating.go — Repository with HTTP concerns.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"  // <-- VIOLATION: Repository doing JSON serialization
	"net/http"       // <-- VIOLATION: Repository importing HTTP package
	"super-lint-test/internal/domain"
	"super-lint-test/pkg/logging"
)

// UserRepoViolating mixes data access with HTTP response handling.
type UserRepoViolating struct {
	db     *sql.DB
	logger logging.Logger
}

// NewUserRepoViolating creates violating repository.
func NewUserRepoViolating(db *sql.DB, logger logging.Logger) *UserRepoViolating {
	return &UserRepoViolating{
		db:     db,
		logger: logger,
	}
}

// GetByIDWithResponse retrieves user and writes HTTP response.
func (r *UserRepoViolating) GetByIDWithResponse(
	ctx context.Context,
	id string,
	w http.ResponseWriter,  // <-- VIOLATION: Repository method takes http.ResponseWriter
) error {
	query := `SELECT id, email, username FROM users WHERE id = $1`

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
	)

	if err == sql.ErrNoRows {
		// Repository writing HTTP responses - architectural violation
		w.WriteHeader(http.StatusNotFound)  // <-- VIOLATION
		json.NewEncoder(w).Encode(map[string]string{  // <-- VIOLATION
			"error": "user not found",
		})
		return nil
	}

	if err != nil {
		r.logger.Error("failed to query user", "error", err)
		w.WriteHeader(http.StatusInternalServerError)  // <-- VIOLATION
		json.NewEncoder(w).Encode(map[string]string{  // <-- VIOLATION
			"error": "internal server error",
		})
		return err
	}

	// Repository formatting and writing HTTP response
	w.Header().Set("Content-Type", "application/json")  // <-- VIOLATION
	w.WriteHeader(http.StatusOK)                         // <-- VIOLATION
	return json.NewEncoder(w).Encode(user)               // <-- VIOLATION
}
