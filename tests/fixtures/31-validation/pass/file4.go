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
