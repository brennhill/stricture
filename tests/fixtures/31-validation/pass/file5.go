// handler/user_handler.go — HTTP handlers for user endpoints.
package handler

import (
	"encoding/json"
	"net/http"
	"super-lint-test/internal/service"
	"super-lint-test/pkg/httputil"
	"super-lint-test/pkg/logging"

	"github.com/gorilla/mux"
)

// UserHandler handles HTTP requests for user operations.
type UserHandler struct {
	service service.UserService
	logger  logging.Logger
}

// NewUserHandler creates a new user handler.
func NewUserHandler(svc service.UserService, logger logging.Logger) *UserHandler {
	return &UserHandler{
		service: svc,
		logger:  logger,
	}
}

// CreateUserRequest represents the request body for creating a user.
type CreateUserRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

// UserResponse represents a user in API responses.
type UserResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

// CreateUser handles POST /users.
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.service.CreateUser(r.Context(), req.Email, req.Username)
	if err != nil {
		h.logger.Error("failed to create user", "error", err)
		httputil.RespondError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	resp := UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		IsActive: user.IsActive,
	}

	httputil.RespondJSON(w, http.StatusCreated, resp)
}

// GetUser handles GET /users/{id}.
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	user, err := h.service.GetUser(r.Context(), id)
	if err == service.ErrUserNotFound {
		httputil.RespondError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		h.logger.Error("failed to get user", "error", err)
		httputil.RespondError(w, http.StatusInternalServerError, "failed to retrieve user")
		return
	}

	resp := UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		IsActive: user.IsActive,
	}

	httputil.RespondJSON(w, http.StatusOK, resp)
}

// ActivateUser handles POST /users/{id}/activate.
func (h *UserHandler) ActivateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.ActivateUser(r.Context(), id); err != nil {
		h.logger.Error("failed to activate user", "error", err)
		httputil.RespondError(w, http.StatusInternalServerError, "failed to activate user")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]string{"status": "activated"})
}

// DeactivateUser handles POST /users/{id}/deactivate.
func (h *UserHandler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.service.DeactivateUser(r.Context(), id); err != nil {
		h.logger.Error("failed to deactivate user", "error", err)
		httputil.RespondError(w, http.StatusInternalServerError, "failed to deactivate user")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]string{"status": "deactivated"})
}
