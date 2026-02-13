// handler/user_handler_violating.go — Handler directly accessing repository.
package handler

import (
	"encoding/json"
	"net/http"
	"super-lint-test/internal/domain"
	"super-lint-test/internal/repository/postgres"  // <-- VIOLATION: Handler imports repository
	"super-lint-test/pkg/httputil"
	"super-lint-test/pkg/logging"

	"github.com/gorilla/mux"
)

// UserHandlerViolating demonstrates handler bypassing service layer.
type UserHandlerViolating struct {
	repo   *postgres.UserRepo  // <-- VIOLATION: Direct repository dependency
	logger logging.Logger
}

// NewUserHandlerViolating creates a violating handler.
func NewUserHandlerViolating(repo *postgres.UserRepo, logger logging.Logger) *UserHandlerViolating {
	return &UserHandlerViolating{
		repo:   repo,
		logger: logger,
	}
}

// GetUser handles GET /users/{id} by directly querying repository.
func (h *UserHandlerViolating) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Handler directly calls repository - bypassing all business logic
	user, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		h.logger.Error("failed to get user", "error", err)
		httputil.RespondError(w, http.StatusInternalServerError, "failed to retrieve user")
		return
	}

	if user == nil {
		httputil.RespondError(w, http.StatusNotFound, "user not found")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, user)
}
