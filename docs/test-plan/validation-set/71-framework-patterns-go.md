# 71 — Go Framework Patterns

Route detection for Go HTTP frameworks. Shows how Stricture identifies route handlers across 4 major frameworks.

**Frameworks Covered:**
- `net/http` (stdlib)
- Chi
- Gin
- Echo

**Pattern:** Each framework section contains:
- 1 PERFECT implementation (production-ready handler)
- 3 VIOLATIONS (V1: input validation, V2: status codes, V3: auth/middleware)

---

## net/http (Standard Library)

### PERFECT: Standard HTTP Handler

```go
// handlers/user_handler.go — User management HTTP handlers using stdlib.

package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
)

// User represents a user entity.
type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// CreateUserRequest represents the request body for user creation.
type CreateUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// ErrorResponse represents a standard error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// CreateUserHandler handles POST /users requests.
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Method check
	if r.Method != http.MethodPost {
		sendErrorResponse(w, "method_not_allowed", "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req CreateUserRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		sendErrorResponse(w, "invalid_request", "Failed to parse request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate email
	if req.Email == "" {
		sendErrorResponse(w, "validation_error", "Email is required", http.StatusBadRequest)
		return
	}
	if !emailRegex.MatchString(req.Email) {
		sendErrorResponse(w, "validation_error", "Invalid email format", http.StatusBadRequest)
		return
	}

	// Validate name
	if req.Name == "" {
		sendErrorResponse(w, "validation_error", "Name is required", http.StatusBadRequest)
		return
	}
	if len(req.Name) < 2 || len(req.Name) > 100 {
		sendErrorResponse(w, "validation_error", "Name must be between 2 and 100 characters", http.StatusBadRequest)
		return
	}

	// Create user (mock implementation)
	user := User{
		ID:    fmt.Sprintf("user_%d", 12345),
		Email: req.Email,
		Name:  req.Name,
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// sendErrorResponse sends a standardized error response.
func sendErrorResponse(w http.ResponseWriter, errorType, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errResp := ErrorResponse{
		Error:   errorType,
		Message: message,
		Code:    statusCode,
	}

	if err := json.NewEncoder(w).Encode(errResp); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

// SetupRoutes configures all HTTP routes.
func SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// User routes
	mux.HandleFunc("/users", CreateUserHandler)
	mux.HandleFunc("/users/", GetUserHandler)

	// Health check
	mux.HandleFunc("/health", HealthHandler)

	return mux
}

// GetUserHandler handles GET /users/{id} requests.
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendErrorResponse(w, "method_not_allowed", "Only GET requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from path
	userID := r.URL.Path[len("/users/"):]
	if userID == "" {
		sendErrorResponse(w, "validation_error", "User ID is required", http.StatusBadRequest)
		return
	}

	// Mock user retrieval
	user := User{
		ID:    userID,
		Email: "user@example.com",
		Name:  "Test User",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// HealthHandler handles GET /health requests.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{
		"status": "healthy",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode health response: %v", err)
	}
}
```

**What makes this PERFECT:**
- Comprehensive input validation (email format, name length)
- Proper HTTP status codes (201 Created, 400 Bad Request, 405 Method Not Allowed)
- Structured error responses with consistent format
- Content-Type headers set correctly
- No panics, all errors handled gracefully
- DisallowUnknownFields prevents injection of unexpected data

---

### VIOLATION 1: Missing Input Validation

```go
// handlers/product_handler.go — Product handler with missing validation.

package handlers

import (
	"encoding/json"
	"net/http"
)

type Product struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type CreateProductRequest struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// CreateProductHandler creates a new product.
// VIOLATION: No input validation on name or price
func CreateProductHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateProductRequest

	// Decode request without validation
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// PROBLEM: Accepts empty name, negative price, price = 0, etc.
	product := Product{
		ID:    "prod_123",
		Name:  req.Name,  // Could be empty string
		Price: req.Price, // Could be negative or zero
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}
```

**Issues:**
- No validation on `Name` (accepts empty string)
- No validation on `Price` (accepts negative, zero, or extreme values)
- No email/format validation where applicable
- Missing DisallowUnknownFields on decoder

---

### VIOLATION 2: Missing or Wrong Status Codes

```go
// handlers/order_handler.go — Order handler with wrong status codes.

package handlers

import (
	"encoding/json"
	"net/http"
)

type Order struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Total  float64 `json:"total"`
}

// CreateOrderHandler creates a new order.
// VIOLATION: Uses wrong HTTP status codes
func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	var order Order

	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		w.Header().Set("Content-Type", "application/json")
		// PROBLEM: Returns 200 OK for invalid input instead of 400 Bad Request
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request",
		})
		return
	}
	defer r.Body.Close()

	// PROBLEM: Returns 200 OK for resource creation instead of 201 Created
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

// GetOrderHandler retrieves an order.
// VIOLATION: Wrong status code for not found
func GetOrderHandler(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("id")

	// Simulate order not found
	if orderID == "" {
		// PROBLEM: Returns 400 Bad Request for missing resource instead of 404 Not Found
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Order not found",
		})
		return
	}

	order := Order{ID: orderID, UserID: "user_1", Total: 99.99}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}
```

**Issues:**
- Returns 200 OK for validation errors (should be 400)
- Returns 200 OK for resource creation (should be 201)
- Returns 400 Bad Request for missing resource (should be 404)
- Inconsistent error response format

---

### VIOLATION 3: Panic in Handler

```go
// handlers/payment_handler.go — Payment handler with panic.

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Payment struct {
	ID     string  `json:"id"`
	Amount float64 `json:"amount"`
	Status string  `json:"status"`
}

// ProcessPaymentHandler processes a payment.
// VIOLATION: Can panic and crash the server
func ProcessPaymentHandler(w http.ResponseWriter, r *http.Request) {
	var payment Payment

	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// PROBLEM: Division by zero can panic
	discount := payment.Amount / 0.0
	finalAmount := payment.Amount - discount

	// PROBLEM: Map access without initialization can panic
	var statusMap map[string]string
	payment.Status = statusMap["pending"] // nil map panic

	// PROBLEM: Slice index out of bounds can panic
	var items []string
	firstItem := items[0] // panic if empty

	// PROBLEM: Type assertion without check can panic
	var data interface{} = "not a number"
	amount := data.(float64) // panic on wrong type

	response := map[string]interface{}{
		"payment":      payment,
		"final_amount": finalAmount,
		"first_item":   firstItem,
		"amount_typed": amount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// AnotherPanicHandler demonstrates nil pointer dereference.
func AnotherPanicHandler(w http.ResponseWriter, r *http.Request) {
	var user *User

	// PROBLEM: Nil pointer dereference
	fmt.Fprintf(w, "User email: %s", user.Email) // panic
}
```

**Issues:**
- Division by zero causes panic
- Nil map access causes panic
- Slice index out of bounds causes panic
- Type assertion without check causes panic
- Nil pointer dereference causes panic
- No recovery middleware to catch panics

---

## Chi Router

### PERFECT: Chi Router with Middleware

```go
// routes/api.go — Chi router with proper middleware and handlers.

package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// APIRouter sets up the Chi router with all routes and middleware.
func APIRouter() chi.Router {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Public routes
	r.Group(func(r chi.Router) {
		r.Post("/auth/login", LoginHandler)
		r.Get("/health", HealthCheckHandler)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)

		r.Route("/api/v1", func(r chi.Router) {
			r.Route("/users", func(r chi.Router) {
				r.Post("/", CreateUserHandlerChi)
				r.Get("/{userID}", GetUserHandlerChi)
				r.Put("/{userID}", UpdateUserHandlerChi)
				r.Delete("/{userID}", DeleteUserHandlerChi)
			})

			r.Route("/posts", func(r chi.Router) {
				r.Post("/", CreatePostHandler)
				r.Get("/", ListPostsHandler)
				r.Get("/{postID}", GetPostHandler)
			})
		})
	})

	return r
}

// UserRequest represents user creation/update request.
type UserRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password,omitempty"`
}

// UserResponse represents user data in responses.
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// APIError represents a structured error response.
type APIError struct {
	Error      string                 `json:"error"`
	Message    string                 `json:"message"`
	StatusCode int                    `json:"status_code"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

var (
	emailRegexChi = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// CreateUserHandlerChi handles POST /api/v1/users.
func CreateUserHandlerChi(w http.ResponseWriter, r *http.Request) {
	var req UserRequest

	// Decode with validation
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, "invalid_request", "Failed to parse request body", http.StatusBadRequest, nil)
		return
	}
	defer r.Body.Close()

	// Validate email
	if req.Email == "" {
		respondWithError(w, "validation_error", "Email is required", http.StatusBadRequest, map[string]interface{}{
			"field": "email",
		})
		return
	}
	if !emailRegexChi.MatchString(req.Email) {
		respondWithError(w, "validation_error", "Invalid email format", http.StatusBadRequest, map[string]interface{}{
			"field": "email",
			"value": req.Email,
		})
		return
	}

	// Validate name
	if req.Name == "" {
		respondWithError(w, "validation_error", "Name is required", http.StatusBadRequest, map[string]interface{}{
			"field": "name",
		})
		return
	}
	if len(req.Name) < 2 || len(req.Name) > 100 {
		respondWithError(w, "validation_error", "Name must be between 2 and 100 characters", http.StatusBadRequest, map[string]interface{}{
			"field":  "name",
			"length": len(req.Name),
		})
		return
	}

	// Validate password
	if req.Password == "" {
		respondWithError(w, "validation_error", "Password is required", http.StatusBadRequest, map[string]interface{}{
			"field": "password",
		})
		return
	}
	if len(req.Password) < 8 {
		respondWithError(w, "validation_error", "Password must be at least 8 characters", http.StatusBadRequest, map[string]interface{}{
			"field":        "password",
			"min_length":   8,
			"given_length": len(req.Password),
		})
		return
	}

	// Create user (mock)
	user := UserResponse{
		ID:        fmt.Sprintf("user_%d", time.Now().Unix()),
		Email:     req.Email,
		Name:      req.Name,
		CreatedAt: time.Now(),
	}

	respondWithJSON(w, http.StatusCreated, user)
}

// GetUserHandlerChi handles GET /api/v1/users/{userID}.
func GetUserHandlerChi(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	if userID == "" {
		respondWithError(w, "validation_error", "User ID is required", http.StatusBadRequest, nil)
		return
	}

	// Mock user retrieval
	user := UserResponse{
		ID:        userID,
		Email:     "user@example.com",
		Name:      "Test User",
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}

	respondWithJSON(w, http.StatusOK, user)
}

// UpdateUserHandlerChi handles PUT /api/v1/users/{userID}.
func UpdateUserHandlerChi(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	if userID == "" {
		respondWithError(w, "validation_error", "User ID is required", http.StatusBadRequest, nil)
		return
	}

	var req UserRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		respondWithError(w, "invalid_request", "Failed to parse request body", http.StatusBadRequest, nil)
		return
	}
	defer r.Body.Close()

	// Validate fields (same as create, but password optional)
	if req.Email != "" && !emailRegexChi.MatchString(req.Email) {
		respondWithError(w, "validation_error", "Invalid email format", http.StatusBadRequest, nil)
		return
	}

	if req.Name != "" && (len(req.Name) < 2 || len(req.Name) > 100) {
		respondWithError(w, "validation_error", "Name must be between 2 and 100 characters", http.StatusBadRequest, nil)
		return
	}

	// Mock update
	user := UserResponse{
		ID:        userID,
		Email:     req.Email,
		Name:      req.Name,
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}

	respondWithJSON(w, http.StatusOK, user)
}

// DeleteUserHandlerChi handles DELETE /api/v1/users/{userID}.
func DeleteUserHandlerChi(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	if userID == "" {
		respondWithError(w, "validation_error", "User ID is required", http.StatusBadRequest, nil)
		return
	}

	// Mock deletion
	respondWithJSON(w, http.StatusNoContent, nil)
}

// AuthMiddleware validates authentication tokens.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			respondWithError(w, "unauthorized", "Authorization header is required", http.StatusUnauthorized, nil)
			return
		}

		// Validate Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondWithError(w, "unauthorized", "Invalid authorization header format", http.StatusUnauthorized, nil)
			return
		}

		token := parts[1]
		if len(token) < 10 {
			respondWithError(w, "unauthorized", "Invalid token", http.StatusUnauthorized, nil)
			return
		}

		// Mock token validation (in production, validate JWT, check DB, etc.)
		// For this example, accept any token with length >= 10

		next.ServeHTTP(w, r)
	})
}

// LoginHandler handles POST /auth/login.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "invalid_request", "Failed to parse request body", http.StatusBadRequest, nil)
		return
	}
	defer r.Body.Close()

	// Mock authentication
	token := fmt.Sprintf("mock_token_%d", time.Now().Unix())

	respondWithJSON(w, http.StatusOK, map[string]string{
		"token": token,
	})
}

// HealthCheckHandler handles GET /health.
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// Helper functions

func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if payload != nil {
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			log.Printf("Failed to encode JSON response: %v", err)
		}
	}
}

func respondWithError(w http.ResponseWriter, errorType, message string, statusCode int, details map[string]interface{}) {
	apiErr := APIError{
		Error:      errorType,
		Message:    message,
		StatusCode: statusCode,
		Details:    details,
	}

	respondWithJSON(w, statusCode, apiErr)
}
```

**What makes this PERFECT:**
- Proper Chi middleware chain (RequestID, Logger, Recoverer, Timeout)
- Route grouping for public vs protected endpoints
- AuthMiddleware validates Bearer token format
- All handlers validate input thoroughly
- Proper HTTP status codes (201 Created, 204 No Content, 401 Unauthorized)
- Structured error responses with details
- DisallowUnknownFields prevents injection

---

### VIOLATION 1: Missing Request Body Validation

```go
// routes/blog.go — Chi router with missing validation.

package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Post struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	UserID  string `json:"user_id"`
}

// CreatePostHandler creates a new blog post.
// VIOLATION: No input validation
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	var post Post

	// PROBLEM: No DisallowUnknownFields, no validation
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// PROBLEM: Accepts empty title, empty content, invalid user_id
	post.ID = "post_123"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

// ListPostsHandler lists all posts.
// VIOLATION: No query parameter validation
func ListPostsHandler(w http.ResponseWriter, r *http.Request) {
	// PROBLEM: No validation on limit/offset query params
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	// Could be empty, negative, or non-numeric values
	_ = limit
	_ = offset

	posts := []Post{
		{ID: "1", Title: "Post 1", Content: "Content 1"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// GetPostHandler retrieves a single post.
func GetPostHandler(w http.ResponseWriter, r *http.Request) {
	postID := chi.URLParam(r, "postID")

	// PROBLEM: No validation that postID is non-empty or valid format
	post := Post{
		ID:      postID,
		Title:   "Sample Post",
		Content: "Sample content",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}
```

**Issues:**
- No DisallowUnknownFields on JSON decoder
- No validation on required fields (title, content, user_id)
- No validation on URL parameters
- No validation on query parameters
- Accepts invalid data silently

---

### VIOLATION 2: Wrong Error Response Shape

```go
// routes/comments.go — Chi router with inconsistent error handling.

package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Comment struct {
	ID      string `json:"id"`
	PostID  string `json:"post_id"`
	Author  string `json:"author"`
	Content string `json:"content"`
}

// CreateCommentHandler creates a new comment.
// VIOLATION: Inconsistent error response formats
func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	var comment Comment

	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		// PROBLEM: Plain text error instead of JSON
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if comment.Content == "" {
		// PROBLEM: Different error format (just a string)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Content is required")
		return
	}

	if comment.PostID == "" {
		// PROBLEM: Another different format (map with different keys)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"err": "Post ID missing",
		})
		return
	}

	// Success response has yet another shape
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":    comment,
		"success": true,
	})
}

// UpdateCommentHandler updates a comment.
func UpdateCommentHandler(w http.ResponseWriter, r *http.Request) {
	commentID := chi.URLParam(r, "commentID")

	var comment Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		// PROBLEM: Yet another error format
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(struct {
			ErrorMessage string `json:"error_message"`
		}{
			ErrorMessage: err.Error(),
		})
		return
	}
	defer r.Body.Close()

	comment.ID = commentID

	// PROBLEM: Success response has different shape than create
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}
```

**Issues:**
- Multiple error response formats (plain text, string, map, struct)
- No consistent error structure across handlers
- Some errors return JSON, others plain text
- Success responses have different shapes
- No standard error codes or types

---

### VIOLATION 3: Middleware Not Checking Auth

```go
// routes/admin.go — Chi router with broken auth middleware.

package routes

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// AdminRouter sets up admin routes.
func AdminRouter() chi.Router {
	r := chi.NewRouter()

	// VIOLATION: Auth middleware doesn't actually check anything
	r.Use(BrokenAuthMiddleware)

	r.Get("/admin/users", AdminListUsersHandler)
	r.Delete("/admin/users/{userID}", AdminDeleteUserHandler)

	return r
}

// BrokenAuthMiddleware pretends to check auth but doesn't.
// VIOLATION: Does not actually validate authentication
func BrokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// PROBLEM: Gets header but never validates it
		authHeader := r.Header.Get("Authorization")
		_ = authHeader

		// PROBLEM: Always calls next, even without valid auth
		next.ServeHTTP(w, r)
	})
}

// AdminListUsersHandler lists all users (admin only).
func AdminListUsersHandler(w http.ResponseWriter, r *http.Request) {
	// PROBLEM: No additional auth check in handler
	users := []map[string]string{
		{"id": "1", "email": "user1@example.com"},
		{"id": "2", "email": "user2@example.com"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// AdminDeleteUserHandler deletes a user (admin only).
func AdminDeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	// PROBLEM: No auth check, anyone can delete users
	w.WriteHeader(http.StatusNoContent)
}

// AnotherBrokenMiddleware shows more auth failures.
func AnotherBrokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-API-Key")

		// PROBLEM: Accepts any token, even empty
		if token != "" || token == "" {
			next.ServeHTTP(w, r)
			return
		}

		// This code is unreachable
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}
```

**Issues:**
- Middleware gets auth header but never validates it
- Always calls next handler regardless of auth status
- No rejection of missing or invalid tokens
- Logic errors make auth checks useless (if true OR true)
- Protected endpoints accessible without authentication

---

## Gin Framework

### PERFECT: Gin Handler with Validation

```go
// controllers/user_controller.go — Gin controller with proper validation.

package controllers

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

// UserController handles user-related requests.
type UserController struct{}

// UserCreateRequest represents user creation input.
type UserCreateRequest struct {
	Email    string `json:"email" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
	Age      int    `json:"age" binding:"required,gte=18,lte=120"`
}

// UserUpdateRequest represents user update input.
type UserUpdateRequest struct {
	Email string `json:"email" binding:"omitempty,email"`
	Name  string `json:"name" binding:"omitempty,min=2,max=100"`
	Age   int    `json:"age" binding:"omitempty,gte=18,lte=120"`
}

// UserDTO represents user data in responses.
type UserDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
}

// ErrorResponse represents a structured error.
type ErrorResponseGin struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

var emailRegexGin = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Create handles POST /api/users.
func (uc *UserController) Create(c *gin.Context) {
	var req UserCreateRequest

	// ShouldBindJSON validates struct tags
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponseGin{
			Error:   "validation_error",
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Additional email format validation
	if !emailRegexGin.MatchString(req.Email) {
		c.JSON(http.StatusBadRequest, ErrorResponseGin{
			Error:   "validation_error",
			Message: "Invalid email format",
			Details: gin.H{"field": "email"},
		})
		return
	}

	// Additional password strength check
	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, ErrorResponseGin{
			Error:   "validation_error",
			Message: "Password must be at least 8 characters",
			Details: gin.H{"field": "password", "min_length": 8},
		})
		return
	}

	// Create user (mock)
	user := UserDTO{
		ID:        fmt.Sprintf("user_%d", time.Now().Unix()),
		Email:     req.Email,
		Name:      req.Name,
		Age:       req.Age,
		CreatedAt: time.Now(),
	}

	c.JSON(http.StatusCreated, user)
}

// Get handles GET /api/users/:id.
func (uc *UserController) Get(c *gin.Context) {
	userID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponseGin{
			Error:   "validation_error",
			Message: "User ID is required",
		})
		return
	}

	// Mock user retrieval
	user := UserDTO{
		ID:        userID,
		Email:     "user@example.com",
		Name:      "Test User",
		Age:       25,
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}

	c.JSON(http.StatusOK, user)
}

// Update handles PUT /api/users/:id.
func (uc *UserController) Update(c *gin.Context) {
	userID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponseGin{
			Error:   "validation_error",
			Message: "User ID is required",
		})
		return
	}

	var req UserUpdateRequest

	// Validate request with struct tags
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponseGin{
			Error:   "validation_error",
			Message: "Invalid request body",
			Details: err.Error(),
		})
		return
	}

	// Additional email validation if provided
	if req.Email != "" && !emailRegexGin.MatchString(req.Email) {
		c.JSON(http.StatusBadRequest, ErrorResponseGin{
			Error:   "validation_error",
			Message: "Invalid email format",
			Details: gin.H{"field": "email"},
		})
		return
	}

	// Mock update
	user := UserDTO{
		ID:        userID,
		Email:     req.Email,
		Name:      req.Name,
		Age:       req.Age,
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}

	c.JSON(http.StatusOK, user)
}

// Delete handles DELETE /api/users/:id.
func (uc *UserController) Delete(c *gin.Context) {
	userID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponseGin{
			Error:   "validation_error",
			Message: "User ID is required",
		})
		return
	}

	// Mock deletion
	c.Status(http.StatusNoContent)
}

// List handles GET /api/users with pagination.
func (uc *UserController) List(c *gin.Context) {
	// Parse query parameters with validation
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	// Validate pagination params (simplified)
	if page == "" || limit == "" {
		c.JSON(http.StatusBadRequest, ErrorResponseGin{
			Error:   "validation_error",
			Message: "Invalid pagination parameters",
		})
		return
	}

	// Mock user list
	users := []UserDTO{
		{
			ID:        "user_1",
			Email:     "user1@example.com",
			Name:      "User One",
			Age:       25,
			CreatedAt: time.Now().Add(-48 * time.Hour),
		},
		{
			ID:        "user_2",
			Email:     "user2@example.com",
			Name:      "User Two",
			Age:       30,
			CreatedAt: time.Now().Add(-24 * time.Hour),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"page":  page,
		"limit": limit,
		"total": 2,
	})
}

// SetupRoutes configures all user routes.
func SetupUserRoutes(r *gin.Engine) {
	uc := &UserController{}

	// Apply auth middleware to protected routes
	authorized := r.Group("/api")
	authorized.Use(AuthMiddlewareGin())

	authorized.POST("/users", uc.Create)
	authorized.GET("/users/:id", uc.Get)
	authorized.PUT("/users/:id", uc.Update)
	authorized.DELETE("/users/:id", uc.Delete)
	authorized.GET("/users", uc.List)
}

// AuthMiddlewareGin validates JWT tokens.
func AuthMiddlewareGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponseGin{
				Error:   "unauthorized",
				Message: "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Validate Bearer token format
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, ErrorResponseGin{
				Error:   "unauthorized",
				Message: "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := authHeader[7:]
		if len(token) < 10 {
			c.JSON(http.StatusUnauthorized, ErrorResponseGin{
				Error:   "unauthorized",
				Message: "Invalid token",
			})
			c.Abort()
			return
		}

		// Mock token validation
		// In production: parse JWT, validate signature, check expiry, etc.

		c.Next()
	}
}
```

**What makes this PERFECT:**
- Uses Gin's built-in validation with struct tags (required, min, max, gte, lte)
- ShouldBindJSON properly checks errors
- Additional custom validation for email format and password strength
- Proper HTTP status codes (201, 204, 401, 400)
- Structured error responses with details
- Auth middleware properly uses c.Abort() to stop chain
- Consistent response format across all handlers

---

### VIOLATION 1: Missing ShouldBindJSON Error Check

```go
// controllers/product_controller.go — Gin controller with missing validation.

package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProductRequest struct {
	Name  string  `json:"name" binding:"required"`
	Price float64 `json:"price" binding:"required,gt=0"`
	SKU   string  `json:"sku" binding:"required"`
}

type ProductResponse struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// CreateProduct handles POST /products.
// VIOLATION: Ignores ShouldBindJSON error
func CreateProduct(c *gin.Context) {
	var req ProductRequest

	// PROBLEM: Error is ignored, continues with invalid data
	_ = c.ShouldBindJSON(&req)

	// req might be empty or invalid, but we use it anyway
	product := ProductResponse{
		ID:    "prod_123",
		Name:  req.Name,  // Could be empty
		Price: req.Price, // Could be 0 or negative
	}

	c.JSON(http.StatusCreated, product)
}

// UpdateProduct handles PUT /products/:id.
// VIOLATION: Doesn't check if binding failed
func UpdateProduct(c *gin.Context) {
	productID := c.Param("id")
	var req ProductRequest

	// PROBLEM: Calls ShouldBindJSON but doesn't check return value
	c.ShouldBindJSON(&req)

	// Continues regardless of validation failure
	product := ProductResponse{
		ID:    productID,
		Name:  req.Name,
		Price: req.Price,
	}

	c.JSON(http.StatusOK, product)
}

// BulkCreateProducts handles POST /products/bulk.
func BulkCreateProducts(c *gin.Context) {
	var products []ProductRequest

	// PROBLEM: Uses BindJSON (not ShouldBindJSON) which returns 400 automatically
	// but doesn't give us control over error response format
	if err := c.BindJSON(&products); err != nil {
		// This might never run because BindJSON already sent response
		return
	}

	// Process products
	c.JSON(http.StatusCreated, gin.H{"count": len(products)})
}
```

**Issues:**
- ShouldBindJSON errors are ignored
- Continues processing with invalid/empty data
- No validation error responses
- Uses BindJSON instead of ShouldBindJSON (loses control over error format)

---

### VIOLATION 2: Wrong Status Code

```go
// controllers/order_controller.go — Gin controller with wrong status codes.

package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type OrderRequest struct {
	UserID string  `json:"user_id" binding:"required"`
	Total  float64 `json:"total" binding:"required,gt=0"`
}

// CreateOrder handles POST /orders.
// VIOLATION: Returns wrong status code
func CreateOrder(c *gin.Context) {
	var req OrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		// PROBLEM: Returns 200 OK for validation error instead of 400
		c.JSON(http.StatusOK, gin.H{
			"error": "Invalid request",
		})
		return
	}

	order := gin.H{
		"id":      "order_123",
		"user_id": req.UserID,
		"total":   req.Total,
	}

	// PROBLEM: Returns 200 OK for resource creation instead of 201 Created
	c.JSON(http.StatusOK, order)
}

// GetOrder handles GET /orders/:id.
func GetOrder(c *gin.Context) {
	orderID := c.Param("id")

	// Simulate order not found
	if orderID == "missing" {
		// PROBLEM: Returns 400 Bad Request for missing resource instead of 404
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Order not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    orderID,
		"total": 99.99,
	})
}

// CancelOrder handles DELETE /orders/:id.
func CancelOrder(c *gin.Context) {
	orderID := c.Param("id")

	// PROBLEM: Returns 200 OK for deletion instead of 204 No Content
	c.JSON(http.StatusOK, gin.H{
		"message": "Order cancelled",
		"id":      orderID,
	})
}

// UpdateOrder handles PUT /orders/:id.
func UpdateOrder(c *gin.Context) {
	orderID := c.Param("id")
	var req OrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		// PROBLEM: Returns 500 Internal Server Error for validation error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    orderID,
		"total": req.Total,
	})
}
```

**Issues:**
- Returns 200 OK for validation errors (should be 400)
- Returns 200 OK for resource creation (should be 201)
- Returns 400 Bad Request for missing resource (should be 404)
- Returns 200 OK for deletion (should be 204)
- Returns 500 for validation errors (should be 400)

---

### VIOLATION 3: No Auth Middleware

```go
// controllers/admin_controller.go — Gin controller without auth protection.

package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminController handles admin operations.
type AdminController struct{}

// SetupAdminRoutes configures admin routes.
// VIOLATION: No auth middleware on sensitive endpoints
func SetupAdminRoutes(r *gin.Engine) {
	ac := &AdminController{}

	// PROBLEM: Admin routes have no authentication
	admin := r.Group("/admin")

	admin.GET("/users", ac.ListAllUsers)
	admin.DELETE("/users/:id", ac.DeleteUser)
	admin.PUT("/users/:id/ban", ac.BanUser)
	admin.GET("/stats", ac.GetStats)
}

// ListAllUsers returns all users (admin only).
// VIOLATION: No auth check in handler
func (ac *AdminController) ListAllUsers(c *gin.Context) {
	// PROBLEM: Anyone can access this endpoint
	users := []gin.H{
		{"id": "1", "email": "user1@example.com", "role": "user"},
		{"id": "2", "email": "admin@example.com", "role": "admin"},
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// DeleteUser permanently deletes a user (admin only).
func (ac *AdminController) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	// PROBLEM: No auth check, anyone can delete users
	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted",
		"id":      userID,
	})
}

// BanUser bans a user (admin only).
func (ac *AdminController) BanUser(c *gin.Context) {
	userID := c.Param("id")

	// PROBLEM: No auth or permission check
	c.JSON(http.StatusOK, gin.H{
		"message": "User banned",
		"id":      userID,
	})
}

// GetStats returns system statistics (admin only).
func (ac *AdminController) GetStats(c *gin.Context) {
	// PROBLEM: Sensitive data exposed without auth
	stats := gin.H{
		"total_users":    1000,
		"active_users":   750,
		"revenue":        99999.99,
		"server_load":    0.75,
		"database_size":  "500GB",
	}

	c.JSON(http.StatusOK, stats)
}

// BrokenAuthMiddleware is present but doesn't work.
func BrokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// PROBLEM: Gets token but never validates it
		_ = c.GetHeader("X-Admin-Token")

		// PROBLEM: Always calls Next, never aborts
		c.Next()
	}
}
```

**Issues:**
- No authentication middleware on admin routes
- Sensitive endpoints accessible to anyone
- No role/permission checks
- Auth middleware exists but doesn't validate
- Never calls c.Abort() to stop unauthorized requests

---

## Echo Framework

### PERFECT: Echo Handler with Proper Binding

```go
// handlers/echo_handlers.go — Echo handlers with proper validation.

package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// EchoUserRequest represents user creation input.
type EchoUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=8"`
	Age      int    `json:"age" validate:"required,gte=18,lte=120"`
}

// EchoUserResponse represents user data in responses.
type EchoUserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
}

// EchoErrorResponse represents a structured error.
type EchoErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

var emailRegexEcho = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// CreateEchoUser handles POST /api/users.
func CreateEchoUser(c echo.Context) error {
	var req EchoUserRequest

	// Bind request body
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, EchoErrorResponse{
			Error:   "bind_error",
			Message: "Failed to parse request body",
			Details: err.Error(),
		})
	}

	// Validate request (Echo has built-in validator integration)
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, EchoErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request data",
			Details: err.Error(),
		})
	}

	// Additional email format validation
	if !emailRegexEcho.MatchString(req.Email) {
		return c.JSON(http.StatusBadRequest, EchoErrorResponse{
			Error:   "validation_error",
			Message: "Invalid email format",
			Details: map[string]string{"field": "email"},
		})
	}

	// Create user (mock)
	user := EchoUserResponse{
		ID:        fmt.Sprintf("user_%d", time.Now().Unix()),
		Email:     req.Email,
		Name:      req.Name,
		Age:       req.Age,
		CreatedAt: time.Now(),
	}

	return c.JSON(http.StatusCreated, user)
}

// GetEchoUser handles GET /api/users/:id.
func GetEchoUser(c echo.Context) error {
	userID := c.Param("id")

	if userID == "" {
		return c.JSON(http.StatusBadRequest, EchoErrorResponse{
			Error:   "validation_error",
			Message: "User ID is required",
		})
	}

	// Mock user retrieval
	user := EchoUserResponse{
		ID:        userID,
		Email:     "user@example.com",
		Name:      "Test User",
		Age:       25,
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}

	return c.JSON(http.StatusOK, user)
}

// UpdateEchoUser handles PUT /api/users/:id.
func UpdateEchoUser(c echo.Context) error {
	userID := c.Param("id")

	if userID == "" {
		return c.JSON(http.StatusBadRequest, EchoErrorResponse{
			Error:   "validation_error",
			Message: "User ID is required",
		})
	}

	var req EchoUserRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, EchoErrorResponse{
			Error:   "bind_error",
			Message: "Failed to parse request body",
			Details: err.Error(),
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, EchoErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request data",
			Details: err.Error(),
		})
	}

	// Mock update
	user := EchoUserResponse{
		ID:        userID,
		Email:     req.Email,
		Name:      req.Name,
		Age:       req.Age,
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}

	return c.JSON(http.StatusOK, user)
}

// DeleteEchoUser handles DELETE /api/users/:id.
func DeleteEchoUser(c echo.Context) error {
	userID := c.Param("id")

	if userID == "" {
		return c.JSON(http.StatusBadRequest, EchoErrorResponse{
			Error:   "validation_error",
			Message: "User ID is required",
		})
	}

	// Mock deletion
	return c.NoContent(http.StatusNoContent)
}

// ListEchoUsers handles GET /api/users with pagination.
func ListEchoUsers(c echo.Context) error {
	// Parse query parameters
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")

	if page == "" {
		page = "1"
	}
	if limit == "" {
		limit = "10"
	}

	// Mock user list
	users := []EchoUserResponse{
		{
			ID:        "user_1",
			Email:     "user1@example.com",
			Name:      "User One",
			Age:       25,
			CreatedAt: time.Now().Add(-48 * time.Hour),
		},
		{
			ID:        "user_2",
			Email:     "user2@example.com",
			Name:      "User Two",
			Age:       30,
			CreatedAt: time.Now().Add(-24 * time.Hour),
		},
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"users": users,
		"page":  page,
		"limit": limit,
		"total": 2,
	})
}

// SetupEchoRoutes configures all Echo routes.
func SetupEchoRoutes(e *echo.Echo) {
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	// Public routes
	e.POST("/auth/login", EchoLogin)
	e.GET("/health", EchoHealth)

	// Protected routes
	api := e.Group("/api")
	api.Use(EchoAuthMiddleware)

	api.POST("/users", CreateEchoUser)
	api.GET("/users/:id", GetEchoUser)
	api.PUT("/users/:id", UpdateEchoUser)
	api.DELETE("/users/:id", DeleteEchoUser)
	api.GET("/users", ListEchoUsers)
}

// EchoAuthMiddleware validates authentication.
func EchoAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")

		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, EchoErrorResponse{
				Error:   "unauthorized",
				Message: "Authorization header is required",
			})
		}

		// Validate Bearer token format
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			return c.JSON(http.StatusUnauthorized, EchoErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid authorization header format",
			})
		}

		token := authHeader[7:]
		if len(token) < 10 {
			return c.JSON(http.StatusUnauthorized, EchoErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid token",
			})
		}

		// Mock token validation
		// In production: parse JWT, validate signature, check expiry, etc.

		return next(c)
	}
}

// EchoLogin handles POST /auth/login.
func EchoLogin(c echo.Context) error {
	var req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, EchoErrorResponse{
			Error:   "bind_error",
			Message: "Failed to parse request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, EchoErrorResponse{
			Error:   "validation_error",
			Message: "Invalid credentials",
		})
	}

	// Mock authentication
	token := fmt.Sprintf("mock_token_%d", time.Now().Unix())

	return c.JSON(http.StatusOK, map[string]string{
		"token": token,
	})
}

// EchoHealth handles GET /health.
func EchoHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "healthy",
	})
}
```

**What makes this PERFECT:**
- Proper use of c.Bind() with error checking
- Uses c.Validate() for struct tag validation
- Additional custom validation for email format
- Proper HTTP status codes (201, 204, 401, 400)
- Structured error responses
- Auth middleware returns error to stop chain
- Consistent response format across handlers
- Uses c.NoContent() for 204 responses

---

### VIOLATION 1: Missing Bind Error Check

```go
// handlers/echo_product.go — Echo handler with missing bind check.

package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type EchoProductRequest struct {
	Name  string  `json:"name" validate:"required"`
	Price float64 `json:"price" validate:"required,gt=0"`
}

// CreateEchoProduct handles POST /products.
// VIOLATION: Ignores c.Bind() error
func CreateEchoProduct(c echo.Context) error {
	var req EchoProductRequest

	// PROBLEM: Error is ignored
	_ = c.Bind(&req)

	// PROBLEM: No validation call
	// req might be empty or invalid, but we use it anyway

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"id":    "prod_123",
		"name":  req.Name,  // Could be empty
		"price": req.Price, // Could be 0 or negative
	})
}

// UpdateEchoProduct handles PUT /products/:id.
// VIOLATION: Doesn't check if binding failed
func UpdateEchoProduct(c echo.Context) error {
	productID := c.Param("id")
	var req EchoProductRequest

	// PROBLEM: Calls Bind but doesn't check return value
	c.Bind(&req)

	// PROBLEM: Calls Validate but doesn't check return value
	c.Validate(&req)

	// Continues regardless of validation failure
	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":    productID,
		"name":  req.Name,
		"price": req.Price,
	})
}

// BulkCreateProducts handles POST /products/bulk.
func BulkCreateProducts(c echo.Context) error {
	var products []EchoProductRequest

	// PROBLEM: Bind error ignored
	if c.Bind(&products) != nil {
		// Empty if block, error discarded
	}

	// Process potentially invalid data
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"count": len(products),
	})
}
```

**Issues:**
- c.Bind() errors are ignored
- c.Validate() errors are ignored
- Continues processing with invalid/empty data
- No validation error responses
- Silent failures lead to incorrect data processing

---

### VIOLATION 2: Wrong Status Codes

```go
// handlers/echo_order.go — Echo handler with wrong status codes.

package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type EchoOrderRequest struct {
	UserID string  `json:"user_id" validate:"required"`
	Total  float64 `json:"total" validate:"required,gt=0"`
}

// CreateEchoOrder handles POST /orders.
// VIOLATION: Wrong status codes
func CreateEchoOrder(c echo.Context) error {
	var req EchoOrderRequest

	if err := c.Bind(&req); err != nil {
		// PROBLEM: Returns 200 OK for bind error instead of 400
		return c.JSON(http.StatusOK, map[string]string{
			"error": "Invalid request",
		})
	}

	if err := c.Validate(&req); err != nil {
		// PROBLEM: Returns 500 for validation error instead of 400
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	order := map[string]interface{}{
		"id":      "order_123",
		"user_id": req.UserID,
		"total":   req.Total,
	}

	// PROBLEM: Returns 200 OK for creation instead of 201 Created
	return c.JSON(http.StatusOK, order)
}

// GetEchoOrder handles GET /orders/:id.
func GetEchoOrder(c echo.Context) error {
	orderID := c.Param("id")

	// Simulate order not found
	if orderID == "missing" {
		// PROBLEM: Returns 400 for missing resource instead of 404
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Order not found",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":    orderID,
		"total": 99.99,
	})
}

// DeleteEchoOrder handles DELETE /orders/:id.
func DeleteEchoOrder(c echo.Context) error {
	orderID := c.Param("id")

	// PROBLEM: Returns 200 OK with body for deletion
	// Should use 204 No Content with c.NoContent()
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Order deleted",
		"id":      orderID,
	})
}

// CancelEchoOrder handles POST /orders/:id/cancel.
func CancelEchoOrder(c echo.Context) error {
	orderID := c.Param("id")

	// PROBLEM: Returns 201 Created for state change
	// Should be 200 OK
	return c.JSON(http.StatusCreated, map[string]string{
		"status": "cancelled",
		"id":     orderID,
	})
}
```

**Issues:**
- Returns 200 OK for bind/validation errors (should be 400)
- Returns 500 for validation errors (should be 400)
- Returns 200 OK for creation (should be 201)
- Returns 400 for missing resource (should be 404)
- Returns 200 with body for deletion (should be 204 No Content)
- Returns 201 for state changes (should be 200)

---

### VIOLATION 3: No Auth Middleware

```go
// handlers/echo_admin.go — Echo handlers without auth.

package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// SetupEchoAdminRoutes configures admin routes.
// VIOLATION: No authentication middleware
func SetupEchoAdminRoutes(e *echo.Echo) {
	// PROBLEM: Admin routes have no auth middleware
	admin := e.Group("/admin")

	admin.GET("/users", ListAllEchoUsers)
	admin.DELETE("/users/:id", DeleteEchoUser)
	admin.POST("/users/:id/ban", BanEchoUser)
	admin.GET("/stats", GetEchoStats)
}

// ListAllEchoUsers returns all users (admin only).
// VIOLATION: No auth check
func ListAllEchoUsers(c echo.Context) error {
	// PROBLEM: Anyone can access this endpoint
	users := []map[string]interface{}{
		{"id": "1", "email": "user1@example.com", "role": "user"},
		{"id": "2", "email": "admin@example.com", "role": "admin"},
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"users": users,
	})
}

// DeleteEchoUser permanently deletes a user (admin only).
func DeleteEchoUser(c echo.Context) error {
	userID := c.Param("id")

	// PROBLEM: No auth or permission check
	return c.JSON(http.StatusOK, map[string]string{
		"message": "User deleted",
		"id":      userID,
	})
}

// BanEchoUser bans a user (admin only).
func BanEchoUser(c echo.Context) error {
	userID := c.Param("id")

	// PROBLEM: No auth check
	return c.JSON(http.StatusOK, map[string]string{
		"message": "User banned",
		"id":      userID,
	})
}

// GetEchoStats returns system statistics (admin only).
func GetEchoStats(c echo.Context) error {
	// PROBLEM: Sensitive data exposed without auth
	stats := map[string]interface{}{
		"total_users":    1000,
		"active_users":   750,
		"revenue":        99999.99,
		"server_load":    0.75,
		"database_size":  "500GB",
	}

	return c.JSON(http.StatusOK, stats)
}

// BrokenEchoAuthMiddleware exists but doesn't work.
func BrokenEchoAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// PROBLEM: Gets token but never validates it
		_ = c.Request().Header.Get("X-Admin-Token")

		// PROBLEM: Always calls next, never returns error
		return next(c)
	}
}

// AnotherBrokenMiddleware shows more failures.
func AnotherBrokenMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := c.Request().Header.Get("Authorization")

		// PROBLEM: Logic error makes auth useless
		if token != "" || token == "" {
			return next(c)
		}

		// Unreachable code
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}
}
```

**Issues:**
- No authentication middleware on admin routes
- Sensitive endpoints accessible to anyone
- No role/permission checks
- Auth middleware exists but doesn't validate tokens
- Logic errors make auth checks ineffective
- Never returns error to stop unauthorized requests

---

## Summary

This validation set demonstrates route detection across 4 Go HTTP frameworks:

1. **net/http** — Standard library patterns with HandlerFunc and ServeMux
2. **Chi** — Router with middleware chains and route groups
3. **Gin** — ShouldBindJSON, c.JSON, gin.H patterns
4. **Echo** — c.Bind, c.Validate, c.JSON patterns

**Common Violations:**
- Missing input validation (V1)
- Wrong HTTP status codes (V2)
- Missing or broken authentication (V3)

**Perfect Implementations Include:**
- Comprehensive input validation
- Proper error handling and status codes
- Structured error responses
- Authentication middleware
- No panics or unsafe operations

Total lines: ~950
