package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"innogen-backend/shared/models"
)

// UserRepository handles database queries for the users table.
type UserRepository struct {
	pool *pgxpool.Pool
}

// New creates a new UserRepository with the given connection pool.
func New(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

const selectUserColumns = `SELECT id, email, password_hash, username, full_name, role, is_active, created_at, updated_at FROM users`

// FindByEmail looks up a user by their email address.
// Returns nil, nil if no user is found.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	u := &models.User{}
	err := r.pool.QueryRow(ctx, selectUserColumns+" WHERE lower(email) = lower($1)", email).Scan(
		&u.ID,
		&u.Email,
		&u.Password,
		&u.Username,
		&u.FullName,
		&u.Role,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.FindByEmail: %w", err)
	}
	return u, nil
}

// FindByID looks up a user by their ID.
// Returns nil, nil if no user is found.
func (r *UserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	u := &models.User{}
	err := r.pool.QueryRow(ctx, selectUserColumns+" WHERE id = $1", id).Scan(
		&u.ID,
		&u.Email,
		&u.Password,
		&u.Username,
		&u.FullName,
		&u.Role,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.FindByID: %w", err)
	}
	return u, nil
}

// CreateUser inserts a new user with bcrypt hashed password.
func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash, username, fullName string) (*models.User, error) {
	u := &models.User{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, username, full_name, role, is_active)
         VALUES ($1, $2, $3, $4, 'student', true)
         RETURNING id, email, username, full_name, role, is_active, created_at, updated_at`,
		email, passwordHash, username, fullName,
	).Scan(&u.ID, &u.Email, &u.Username, &u.FullName, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("repository.CreateUser: %w", err)
	}
	return u, nil
}

// CreateRefreshToken stores a hashed refresh token.
func (r *UserRepository) CreateRefreshToken(ctx context.Context, userID string, tokenHash, userAgent, ipAddress string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, user_agent, ip_address, expires_at) VALUES ($1,$2,$3,NULLIF($4,'')::inet,$5)`,
		userID, tokenHash, userAgent, ipAddress, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("repository.CreateRefreshToken: %w", err)
	}
	return nil
}

func (r *UserRepository) FindRefreshToken(ctx context.Context, tokenHash string) (id string, userID string, expiresAt time.Time, revokedAt *time.Time, err error) {
	err = r.pool.QueryRow(ctx,
		`SELECT id, user_id, expires_at, revoked_at FROM refresh_tokens WHERE token_hash=$1`, tokenHash,
	).Scan(&id, &userID, &expiresAt, &revokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", time.Time{}, nil, nil
	}
	if err != nil {
		return "", "", time.Time{}, nil, fmt.Errorf("repository.FindRefreshToken: %w", err)
	}
	return
}

func (r *UserRepository) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked_at=CURRENT_TIMESTAMP WHERE id=$1`, tokenID)
	if err != nil {
		return fmt.Errorf("repository.RevokeRefreshToken: %w", err)
	}
	return nil
}

func (r *UserRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked_at=CURRENT_TIMESTAMP WHERE user_id=$1 AND revoked_at IS NULL`, userID)
	if err != nil {
		return fmt.Errorf("repository.RevokeAllUserRefreshTokens: %w", err)
	}
	return nil
}

// RotateRefreshToken atomically consumes one active token and persists its successor.
func (r *UserRepository) RotateRefreshToken(ctx context.Context, tokenID, userID, hash, userAgent, ip string, expires time.Time) (bool, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)
	result, err := tx.Exec(ctx, `UPDATE refresh_tokens SET revoked_at=now() WHERE id=$1 AND user_id=$2 AND revoked_at IS NULL AND expires_at>now()`, tokenID, userID)
	if err != nil {
		return false, err
	}
	if result.RowsAffected() != 1 {
		return false, nil
	}
	_, err = tx.Exec(ctx, `INSERT INTO refresh_tokens(user_id,token_hash,rotated_from,user_agent,ip_address,expires_at) VALUES($1,$2,$3,$4,NULLIF($5,'')::inet,$6)`, userID, hash, tokenID, userAgent, ip, expires)
	if err != nil {
		return false, err
	}
	if err = tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}
