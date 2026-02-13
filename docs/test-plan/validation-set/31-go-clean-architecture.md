# 31 — Go Clean Architecture (Architecture Validation)

**Why included:** Go cmd/internal/pkg patterns, interface-driven design, Go package visibility. All 6 ARCH rules.

## Architecture Overview

Standard Go project layout following Clean Architecture principles:

```
cmd/server/             # Entry point, DI wiring
internal/
  domain/               # Business entities (zero deps)
  service/              # Business logic (depends on domain)
  repository/           # Data access (depends on domain)
    postgres/           # Concrete implementation
  handler/              # HTTP handlers (depends on service interfaces)
pkg/                    # Shared utilities (importable by external modules)
  httputil/
  logging/
```

## Stricture Configuration

```yaml
arch:
  dependency-direction:
    enabled: true
    layers:
      - name: handler
        depends-on: [service]
      - name: service
        depends-on: [domain]
      - name: repository
        depends-on: [domain]
      - name: domain
        depends-on: []
    shared-packages:
      - pkg/**

  import-boundary:
    enabled: true
    rules:
      - pattern: "internal/**"
        forbidden-from: ["pkg/**"]
        reason: "pkg/ must not depend on internal/"

      - pattern: "*/postgres"
        forbidden-from: ["internal/handler/**", "internal/service/**"]
        reason: "Only repository layer should import concrete DB implementations"

  layer-violation:
    enabled: true
    rules:
      - layer: handler
        forbidden-imports:
          - internal/repository/**
        reason: "Handlers must use service interfaces, not repositories directly"

      - layer: domain
        forbidden-imports:
          - internal/**
          - database/sql
          - github.com/**
        reason: "Domain must have zero dependencies"

  no-circular-deps:
    enabled: true
    max-depth: 3

  max-file-lines:
    enabled: true
    limit: 800
    exclude:
      - "**/*_test.go"
      - "**/generated/**"

  module-boundary:
    enabled: true
    rules:
      - module: internal/service
        must-import-via: "interfaces"
        reason: "Services must be consumed through interfaces, not concrete types"
```

---

## PERFECT Examples

### P01: domain/user.go — Pure Domain Entity (Zero Dependencies)

**Validates:** ARCH-dependency-direction (domain has no deps), ARCH-layer-violation

```go
// domain/user.go — User domain entity with business rules.
package domain

import (
	"errors"
	"time"
)

// User represents a user in the system.
type User struct {
	ID        string
	Email     string
	Username  string
	CreatedAt time.Time
	UpdatedAt time.Time
	IsActive  bool
}

// Validate checks if the user entity is valid.
func (u *User) Validate() error {
	if u.Email == "" {
		return errors.New("email is required")
	}
	if u.Username == "" {
		return errors.New("username is required")
	}
	if len(u.Username) < 3 {
		return errors.New("username must be at least 3 characters")
	}
	return nil
}

// Activate marks the user as active.
func (u *User) Activate() {
	u.IsActive = true
	u.UpdatedAt = time.Now()
}

// Deactivate marks the user as inactive.
func (u *User) Deactivate() {
	u.IsActive = false
	u.UpdatedAt = time.Now()
}
```

**Why perfect:** Domain package imports only standard library. Pure business logic with no infrastructure concerns.

---

### P02: service/interfaces.go — Repository Interfaces (Dependency Inversion)

**Validates:** ARCH-dependency-direction (consumer defines interface), ARCH-module-boundary

```go
// service/interfaces.go — Repository interfaces defined by service layer.
package service

import (
	"context"
	"super-lint-test/internal/domain"
)

// UserRepository defines the interface for user data access.
// The service layer defines this interface, and the repository layer implements it.
// This is dependency inversion — high-level module (service) defines the interface.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*domain.User, error)
}

// SessionRepository defines the interface for session management.
type SessionRepository interface {
	Create(ctx context.Context, userID string, token string) error
	GetByToken(ctx context.Context, token string) (userID string, err error)
	Delete(ctx context.Context, token string) error
	DeleteAllForUser(ctx context.Context, userID string) error
}
```

**Why perfect:** Service layer defines what it needs. Repository layer will implement these interfaces. Classic dependency inversion.

---

### P03: service/user_service.go — Business Logic Layer

**Validates:** ARCH-dependency-direction (service → domain only), ARCH-import-boundary

```go
// service/user_service.go — User service implementing business logic.
package service

import (
	"context"
	"errors"
	"super-lint-test/internal/domain"
	"super-lint-test/pkg/logging"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrUserNotFound is returned when a user is not found.
	ErrUserNotFound = errors.New("user not found")
	// ErrDuplicateEmail is returned when email already exists.
	ErrDuplicateEmail = errors.New("email already exists")
)

// UserService implements user business logic.
type UserService struct {
	repo   UserRepository
	logger logging.Logger
}

// NewUserService creates a new user service.
func NewUserService(repo UserRepository, logger logging.Logger) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger,
	}
}

// CreateUser creates a new user with validation.
func (s *UserService) CreateUser(ctx context.Context, email, username string) (*domain.User, error) {
	// Check for duplicate email
	existing, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, ErrDuplicateEmail
	}

	user := &domain.User{
		ID:        uuid.New().String(),
		Email:     email,
		Username:  username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
	}

	// Validate domain rules
	if err := user.Validate(); err != nil {
		return nil, err
	}

	// Persist
	if err := s.repo.Create(ctx, user); err != nil {
		s.logger.Error("failed to create user", "error", err)
		return nil, err
	}

	s.logger.Info("user created", "user_id", user.ID, "email", user.Email)
	return user, nil
}

// GetUser retrieves a user by ID.
func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// ActivateUser activates a user account.
func (s *UserService) ActivateUser(ctx context.Context, id string) error {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return err
	}

	user.Activate()

	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.Error("failed to activate user", "user_id", id, "error", err)
		return err
	}

	s.logger.Info("user activated", "user_id", id)
	return nil
}

// DeactivateUser deactivates a user account.
func (s *UserService) DeactivateUser(ctx context.Context, id string) error {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return err
	}

	user.Deactivate()

	if err := s.repo.Update(ctx, user); err != nil {
		s.logger.Error("failed to deactivate user", "user_id", id, "error", err)
		return err
	}

	s.logger.Info("user deactivated", "user_id", id)
	return nil
}
```

**Why perfect:** Service imports only domain entities and pkg utilities. Uses repository interface, not concrete implementation.

---

### P04: repository/postgres/user_repo.go — Concrete Repository Implementation

**Validates:** ARCH-dependency-direction (repo → domain), ARCH-module-boundary

```go
// repository/postgres/user_repo.go — PostgreSQL implementation of UserRepository.
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"super-lint-test/internal/domain"
	"super-lint-test/pkg/logging"
)

// UserRepo implements service.UserRepository using PostgreSQL.
type UserRepo struct {
	db     *sql.DB
	logger logging.Logger
}

// NewUserRepo creates a new PostgreSQL user repository.
func NewUserRepo(db *sql.DB, logger logging.Logger) *UserRepo {
	return &UserRepo{
		db:     db,
		logger: logger,
	}
}

// Create inserts a new user into the database.
func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, username, created_at, updated_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Username,
		user.CreatedAt,
		user.UpdatedAt,
		user.IsActive,
	)

	if err != nil {
		r.logger.Error("failed to insert user", "error", err)
		return err
	}

	return nil
}

// GetByID retrieves a user by ID.
func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, email, username, created_at, updated_at, is_active
		FROM users
		WHERE id = $1
	`

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("failed to query user", "error", err)
		return nil, err
	}

	return &user, nil
}

// GetByEmail retrieves a user by email.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, username, created_at, updated_at, is_active
		FROM users
		WHERE email = $1
	`

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("failed to query user by email", "error", err)
		return nil, err
	}

	return &user, nil
}

// Update updates an existing user.
func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $2, username = $3, updated_at = $4, is_active = $5
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Username,
		user.UpdatedAt,
		user.IsActive,
	)

	if err != nil {
		r.logger.Error("failed to update user", "error", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("user not found")
	}

	return nil
}

// Delete removes a user from the database.
func (r *UserRepo) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete user", "error", err)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("user not found")
	}

	return nil
}

// List retrieves users with pagination.
func (r *UserRepo) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	query := `
		SELECT id, email, username, created_at, updated_at, is_active
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Error("failed to list users", "error", err)
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.IsActive,
		)
		if err != nil {
			r.logger.Error("failed to scan user row", "error", err)
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
```

**Why perfect:** Repository imports only domain entities and database/sql. No knowledge of handlers or services. Implements interface defined by service layer.

---

### P05: handler/user_handler.go — HTTP Handler Layer

**Validates:** ARCH-dependency-direction (handler → service interfaces), ARCH-import-boundary

```go
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
```

**Why perfect:** Handler imports only service interfaces and pkg utilities. No direct repository or database imports. Pure HTTP layer concerns.

---

## VIOLATION Examples

### V01 — Handler Directly Imports Repository (ARCH-dependency-direction)

**Violation:** Handler bypasses the service layer and imports repository package directly.

**Expected violation:** `ARCH-dependency-direction` on `handler/user_handler_violating.go`

```go
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
```

**Why Stricture catches this:** The ARCH-dependency-direction rule specifies that handler layer depends only on service layer. Importing `internal/repository/postgres` violates the layer hierarchy. Handler should depend on service interfaces, not concrete repository implementations.

---

### V02 — Service Imports Handler (ARCH-dependency-direction)

**Violation:** Service package imports from handler package, creating reverse dependency flow.

**Expected violation:** `ARCH-dependency-direction` on `service/auth_service.go`

```go
// service/auth_service.go — Service importing handler types.
package service

import (
	"context"
	"errors"
	"super-lint-test/internal/domain"
	"super-lint-test/internal/handler"  // <-- VIOLATION: Service imports handler (reverse flow)
	"super-lint-test/pkg/logging"
	"time"

	"github.com/google/uuid"
)

// AuthService handles authentication with handler types.
type AuthService struct {
	userRepo UserRepository
	logger   logging.Logger
}

// NewAuthService creates auth service.
func NewAuthService(repo UserRepository, logger logging.Logger) *AuthService {
	return &AuthService{
		userRepo: repo,
		logger:   logger,
	}
}

// ValidateCredentials checks credentials and returns handler response type.
func (s *AuthService) ValidateCredentials(ctx context.Context, email, password string) (*handler.UserResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Service returning handler's response type - architectural violation
	resp := &handler.UserResponse{  // <-- VIOLATION: Using handler type in service
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		IsActive: user.IsActive,
	}

	return resp, nil
}
```

**Why Stricture catches this:** Service layer is below handler layer in the dependency hierarchy. Services provide business logic consumed by handlers. A service importing handler types creates a reverse dependency, violating the top-down flow (handler → service → repository → domain).

---

### V03 — Cross-Module Internal Import (ARCH-import-boundary)

**Violation:** User feature directly imports order repository internals, bypassing module boundaries.

**Expected violation:** `ARCH-import-boundary` on `internal/users/service/user_billing_sync.go`

```go
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
```

**Why Stricture catches this:** The ARCH-import-boundary rule forbids imports like `internal/orders/repository/postgres` from other feature modules. If users module needs order data, it should import the orders module's public service interface, not reach into its internal repository implementation.

---

### V04 — cmd/ Imports Internal Subdirectory (ARCH-import-boundary)

**Violation:** main.go imports repository package directly instead of using service layer.

**Expected violation:** `ARCH-import-boundary` on `cmd/server/main.go`

```go
// cmd/server/main.go — Entry point with architectural violations.
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"super-lint-test/internal/handler"
	"super-lint-test/internal/repository/postgres"  // <-- VIOLATION: cmd imports internal subdirectory
	"super-lint-test/pkg/logging"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	logger := logging.NewLogger()

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// main.go directly instantiating repository - should use service layer
	userRepo := postgres.NewUserRepo(db, logger)  // <-- VIOLATION: Bypassing service layer

	// Handler wired directly to repository (no service layer!)
	userHandler := handler.NewUserHandler(userRepo, logger)  // Type mismatch, but shows intent

	mux := http.NewServeMux()
	mux.HandleFunc("/users", userHandler.CreateUser)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		logger.Info("server starting", "port", 8080)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "error", err)
	}
}
```

**Why Stricture catches this:** The ARCH-import-boundary rule states that cmd/ should not import internal subdirectories like `internal/repository/postgres`. The entry point should wire dependencies through service interfaces, not directly instantiate concrete repository implementations.

---

### V05 — Three-Package Circular Chain (ARCH-no-circular-deps)

**Violation:** Handler → Service → Repository → Handler import cycle.

**Expected violation:** `ARCH-no-circular-deps` on all three files forming the cycle.

```go
// handler/order_handler_cycle.go — Part 1 of cycle.
package handler

import (
	"net/http"
	"super-lint-test/internal/service"
	"super-lint-test/pkg/logging"
)

type OrderHandler struct {
	service *service.OrderService
	logger  logging.Logger
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// Handler calls service (normal flow)
	h.service.CreateOrder(r.Context())
}
```

```go
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
```

```go
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
```

**Why Stricture catches this:** The ARCH-no-circular-deps rule detects import cycles of any depth. This creates a 3-node cycle: handler imports service, service imports repository, repository imports handler. Each file appears correct in isolation, but together they form a circular dependency chain.

---

### V06 — Two-Package Circular Dependency (ARCH-no-circular-deps)

**Violation:** Service ↔ Util mutual imports.

**Expected violation:** `ARCH-no-circular-deps` on both `service/validation.go` and `service/util/formatter.go`

```go
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
```

```go
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
```

**Why Stricture catches this:** This is a classic 2-node circular dependency. service/validation.go imports service/util, and service/util/formatter.go imports service. Even though they're in the same parent directory (service/), the cycle exists at the package level. Go's import rules prevent compilation, but Stricture detects it during static analysis.

---

### V07 — 900-Line Handler File (ARCH-max-file-lines)

**Violation:** Handler file exceeds 800 LOC limit with excessive methods.

**Expected violation:** `ARCH-max-file-lines` on `handler/admin_handler.go` (900 lines)

```go
// handler/admin_handler.go — Handler with too many lines (900 total).
package handler

import (
	"encoding/json"
	"net/http"
	"super-lint-test/internal/service"
	"super-lint-test/pkg/httputil"
	"super-lint-test/pkg/logging"

	"github.com/gorilla/mux"
)

// AdminHandler handles all admin operations (90+ methods below).
type AdminHandler struct {
	userService    *service.UserService
	orderService   *service.OrderService
	billingService *service.BillingService
	logger         logging.Logger
}

// NewAdminHandler creates admin handler.
func NewAdminHandler(
	userSvc *service.UserService,
	orderSvc *service.OrderService,
	billingSvc *service.BillingService,
	logger logging.Logger,
) *AdminHandler {
	return &AdminHandler{
		userService:    userSvc,
		orderService:   orderSvc,
		billingService: billingSvc,
		logger:         logger,
	}
}

// The file continues with 80+ more methods (each 8-12 lines):
// ListUsers, GetUser, CreateUser, UpdateUser, DeleteUser, ActivateUser, DeactivateUser,
// BulkCreateUsers, BulkUpdateUsers, BulkDeleteUsers, ExportUsers, ImportUsers,
// ListOrders, GetOrder, CreateOrder, UpdateOrder, DeleteOrder, CancelOrder, RefundOrder,
// BulkCancelOrders, BulkRefundOrders, ExportOrders, ImportOrders,
// ListInvoices, GetInvoice, CreateInvoice, UpdateInvoice, DeleteInvoice, SendInvoice,
// ListPayments, GetPayment, RefundPayment, ChargePayment, ListRefunds, ProcessRefund,
// GetDashboardStats, GetUserAnalytics, GetOrderAnalytics, GetRevenueReport,
// GetUserGrowthReport, GetChurnReport, GetRetentionReport, GetCohortAnalysis,
// ... (60+ more methods) ...
// Total: 900 lines

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Implementation here (10 lines)
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Implementation here (10 lines)
}

// ... 80+ more methods (870 lines) ...
```

**Why Stricture catches this:** The ARCH-max-file-lines rule is set to 800 lines. This handler file has 900 lines due to cramming 90+ methods into a single file. Stricture counts all lines (including comments, blank lines, and code). The file should be refactored into multiple focused handlers (UserAdminHandler, OrderAdminHandler, BillingAdminHandler).

---

### V08 — 1200-Line Service File (ARCH-max-file-lines)

**Violation:** Service file with too many methods and business logic.

**Expected violation:** `ARCH-max-file-lines` on `service/product_service.go` (1200 lines)

```go
// service/product_service.go — Service with excessive LOC (1200 total).
package service

import (
	"context"
	"errors"
	"super-lint-test/internal/domain"
	"super-lint-test/pkg/logging"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProductNotFound     = errors.New("product not found")
	ErrInsufficientStock   = errors.New("insufficient stock")
	ErrInvalidPrice        = errors.New("invalid price")
	ErrDuplicateSKU        = errors.New("duplicate SKU")
	ErrCategoryNotFound    = errors.New("category not found")
	ErrVariantNotFound     = errors.New("variant not found")
	ErrInvalidDiscount     = errors.New("invalid discount")
	ErrPromotionExpired    = errors.New("promotion expired")
)

// ProductRepository interface (20 methods).
type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	GetBySKU(ctx context.Context, sku string) (*domain.Product, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filters ProductFilters) ([]*domain.Product, error)
	// ... 14 more repository methods ...
}

// ProductService implements product business logic.
type ProductService struct {
	repo               ProductRepository
	inventoryService   *InventoryService
	pricingService     *PricingService
	categoryService    *CategoryService
	variantService     *VariantService
	promotionService   *PromotionService
	logger             logging.Logger
}

// NewProductService creates product service.
func NewProductService(
	repo ProductRepository,
	inventorySvc *InventoryService,
	pricingSvc *PricingService,
	categorySvc *CategoryService,
	variantSvc *VariantService,
	promotionSvc *PromotionService,
	logger logging.Logger,
) *ProductService {
	return &ProductService{
		repo:             repo,
		inventoryService: inventorySvc,
		pricingService:   pricingSvc,
		categoryService:  categorySvc,
		variantService:   variantSvc,
		promotionService: promotionSvc,
		logger:           logger,
	}
}

// The file continues with 100+ methods (each 10-15 lines):
// CreateProduct, GetProduct, UpdateProduct, DeleteProduct, ListProducts, SearchProducts,
// AddCategory, RemoveCategory, UpdateCategory, ListCategories,
// AddVariant, RemoveVariant, UpdateVariant, ListVariants, GetVariantBySKU,
// AddPromotion, RemovePromotion, UpdatePromotion, ListPromotions, ApplyPromotion,
// UpdateInventory, CheckStock, ReserveStock, ReleaseStock, GetStockLevel,
// UpdatePrice, GetPrice, CalculateDiscount, ApplyBulkDiscount,
// GenerateSKU, ValidateSKU, ImportProducts, ExportProducts,
// GetProductAnalytics, GetTopSellingProducts, GetLowStockProducts,
// ... (80+ more methods) ...
// Total: 1200 lines

func (s *ProductService) CreateProduct(ctx context.Context, name, sku string, price int64) (*domain.Product, error) {
	// Validation and business logic (15 lines)
	return nil, nil
}

func (s *ProductService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	// Retrieval logic (12 lines)
	return nil, nil
}

// ... 98+ more methods (1150 lines) ...
```

**Why Stricture catches this:** This service file has 1200 lines, far exceeding the 800-line limit. The file tries to handle all product-related operations in one place (catalog management, inventory, pricing, promotions, analytics). Stricture reports this as ARCH-max-file-lines violation. The file should be split into focused services: ProductCatalogService, ProductInventoryService, ProductPricingService, ProductPromotionService.

---

### V09 — SQL Query in Handler (ARCH-layer-violation)

**Violation:** Handler directly constructs and executes SQL queries.

**Expected violation:** `ARCH-layer-violation` on `handler/report_handler.go`

```go
// handler/report_handler.go — Handler with SQL queries (layer violation).
package handler

import (
	"database/sql"  // <-- VIOLATION: Handler importing database/sql
	"encoding/json"
	"net/http"
	"super-lint-test/pkg/httputil"
	"super-lint-test/pkg/logging"

	"github.com/gorilla/mux"
)

// ReportHandler generates reports with direct SQL.
type ReportHandler struct {
	db     *sql.DB  // <-- VIOLATION: Handler has database dependency
	logger logging.Logger
}

// NewReportHandler creates report handler with DB connection.
func NewReportHandler(db *sql.DB, logger logging.Logger) *ReportHandler {
	return &ReportHandler{
		db:     db,
		logger: logger,
	}
}

// GetUserReport handles GET /reports/users.
func (h *ReportHandler) GetUserReport(w http.ResponseWriter, r *http.Request) {
	// Handler directly executing SQL queries - massive architectural violation
	query := `
		SELECT
			DATE(created_at) as date,
			COUNT(*) as user_count,
			COUNT(CASE WHEN is_active THEN 1 END) as active_count
		FROM users
		WHERE created_at >= NOW() - INTERVAL '30 days'
		GROUP BY DATE(created_at)
		ORDER BY date DESC
	`  // <-- VIOLATION: SQL query in handler layer

	rows, err := h.db.QueryContext(r.Context(), query)  // <-- VIOLATION: Direct DB access
	if err != nil {
		h.logger.Error("failed to query report", "error", err)
		httputil.RespondError(w, http.StatusInternalServerError, "failed to generate report")
		return
	}
	defer rows.Close()

	type ReportRow struct {
		Date        string `json:"date"`
		UserCount   int    `json:"user_count"`
		ActiveCount int    `json:"active_count"`
	}

	var results []ReportRow
	for rows.Next() {
		var row ReportRow
		if err := rows.Scan(&row.Date, &row.UserCount, &row.ActiveCount); err != nil {
			h.logger.Error("failed to scan row", "error", err)
			continue
		}
		results = append(results, row)
	}

	httputil.RespondJSON(w, http.StatusOK, results)
}
```

**Why Stricture catches this:** The ARCH-layer-violation rule forbids handler layer from importing `database/sql` or repository packages. Handlers should delegate all data access to service layer. This handler constructs SQL queries and executes them directly - bypassing both service and repository layers. Stricture detects the `database/sql` import in a handler file.

---

### V10 — HTTP Response Formatting in Repository (ARCH-layer-violation)

**Violation:** Repository imports net/http and writes JSON responses.

**Expected violation:** `ARCH-layer-violation` on `repository/postgres/user_repo_violating.go`

```go
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
```

**Why Stricture catches this:** The ARCH-layer-violation rule forbids repository layer from importing `net/http` or performing HTTP response formatting. Repositories should return domain entities or errors - HTTP concerns belong in the handler layer. This repository imports net/http and writes responses directly. Stricture detects net/http import in repository package.

---

### V11 — Accessing Internal Struct Directly (ARCH-module-boundary)

**Violation:** Another module accesses internal unexported types by importing internal package.

**Expected violation:** `ARCH-module-boundary` on `internal/billing/invoice_service.go`

```go
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
```

**Why Stricture catches this:** The ARCH-module-boundary rule enforces that modules should interact through public service interfaces, not by importing internal repository implementations. The billing module should depend on users module's service layer (e.g., `internal/users/service.UserService`), not reach into `internal/users/repository/postgres`. Stricture detects cross-module internal imports.

---

### V12 — Bypassing Interface Contract (ARCH-module-boundary)

**Violation:** Casting interface to concrete type to access internal methods.

**Expected violation:** `ARCH-module-boundary` on `service/notification_service.go`

```go
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
```

**Why Stricture catches this:** The ARCH-module-boundary rule enforces interface-based dependencies. This service receives `UserRepository` interface but performs type assertion to `*postgres.UserRepo` to access internal methods not in the interface contract. This violates the module boundary by creating a hidden dependency on concrete implementation. Stricture detects when code imports concrete types and performs type assertions to bypass interfaces.

---

## Violation Summary

| ID | Rule | File | Issue |
|----|------|------|-------|
| V01 | ARCH-dependency-direction | `handler/user_handler_violating.go` | Handler imports repository package (bypasses service) |
| V02 | ARCH-dependency-direction | `service/auth_service.go` | Service imports handler types (reverse flow) |
| V03 | ARCH-import-boundary | `internal/users/service/user_billing_sync.go` | Users module imports orders repository internals |
| V04 | ARCH-import-boundary | `cmd/server/main.go` | main.go imports repository subdirectory directly |
| V05 | ARCH-no-circular-deps | `handler/order_handler_cycle.go` + 2 others | Handler → Service → Repository → Handler cycle |
| V06 | ARCH-no-circular-deps | `service/validation.go` ↔ `service/util/formatter.go` | Two-package mutual import cycle |
| V07 | ARCH-max-file-lines | `handler/admin_handler.go` | 900 lines (limit: 800) |
| V08 | ARCH-max-file-lines | `service/product_service.go` | 1200 lines (limit: 800) |
| V09 | ARCH-layer-violation | `handler/report_handler.go` | Handler imports database/sql and executes queries |
| V10 | ARCH-layer-violation | `repository/postgres/user_repo_violating.go` | Repository imports net/http and writes responses |
| V11 | ARCH-module-boundary | `internal/billing/invoice_service.go` | Billing imports users repository internals |
| V12 | ARCH-module-boundary | `service/notification_service.go` | Type assertion to concrete type bypasses interface |

---

## Detection Strategy

### How to Run This Validation

1. **False positive test:** Scan the PERFECT codebase (P01-P05). Stricture must report zero ARCH violations.
2. **Detection test:** For each V01-V12, apply the violation to a copy of the PERFECT codebase and scan. Stricture must report exactly the expected violation.
3. **Isolation test:** Each violation is independent. Applying V01 must not trigger V02's rule, and vice versa.
4. **Combination test:** Apply V01 + V05 + V09 simultaneously. Stricture must report all three violations without interference.

### Rule Coverage Matrix

| Rule | PERFECT (TN) | Violations (TP) |
|------|--------------|-----------------|
| ARCH-dependency-direction | All imports flow handler→service→repository→domain | V01, V02 |
| ARCH-import-boundary | No cross-module internal imports, cmd/ uses service layer | V03, V04 |
| ARCH-no-circular-deps | No cycles in dependency graph | V05, V06 |
| ARCH-max-file-lines | All files under 800 LOC | V07, V08 |
| ARCH-layer-violation | Layers respect boundaries (no SQL in handler, no HTTP in repo) | V09, V10 |
| ARCH-module-boundary | Modules consumed through interfaces, no internal access | V11, V12 |

---

