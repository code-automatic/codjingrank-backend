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

const selectUserColumns = `SELECT id, email, password, username, full_name, role, is_active, created_at, updated_at FROM users`

// FindByEmail looks up a user by their email address.
// Returns nil, nil if no user is found.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	u := &models.User{}
	err := r.pool.QueryRow(ctx, selectUserColumns+" WHERE email = $1", email).Scan(
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
func (r *UserRepository) FindByID(ctx context.Context, id int) (*models.User, error) {
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

// UpsertGithubAccount creates or updates a github_accounts row for a user with the given installation.
func (r *UserRepository) UpsertGithubAccount(ctx context.Context, userID int, installationID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO github_accounts (user_id, installation_id, github_owner, github_owner_type, status)
		 VALUES ($1, $2, '', 'User', 'pending')
		 ON CONFLICT (user_id) DO UPDATE SET installation_id = EXCLUDED.installation_id, updated_at = CURRENT_TIMESTAMP`,
		userID, installationID,
	)
	if err != nil {
		return fmt.Errorf("repository.UpsertGithubAccount: %w", err)
	}
	return nil
}

// UpdateGithubAccountOwner updates the owner info and status on a github_accounts row.
func (r *UserRepository) UpdateGithubAccountOwner(ctx context.Context, userID int, owner, ownerType, status string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE github_accounts SET github_owner = $2, github_owner_type = $3, status = $4, updated_at = CURRENT_TIMESTAMP WHERE user_id = $1`,
		userID, owner, ownerType, status,
	)
	if err != nil {
		return fmt.Errorf("repository.UpdateGithubAccountOwner: %w", err)
	}
	return nil
}

// GetGithubInstallationOwner retrieves owner info from github_installations for backfill.
func (r *UserRepository) GetGithubInstallationOwner(ctx context.Context, installationID string) (owner, ownerType string, err error) {
	err = r.pool.QueryRow(ctx,
		`SELECT github_owner, github_owner_type FROM github_installations WHERE installation_id = $1`,
		installationID,
	).Scan(&owner, &ownerType)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", nil
	}
	if err != nil {
		return "", "", fmt.Errorf("repository.GetGithubInstallationOwner: %w", err)
	}
	return owner, ownerType, nil
}

// CreateUser inserts a new user with bcrypt hashed password.
func (r *UserRepository) CreateUser(ctx context.Context, email, passwordHash, username, fullName string) (*models.User, error) {
	u := &models.User{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password, username, full_name, role, is_active)
         VALUES ($1, $2, $3, $4, 'student', true)
         RETURNING id, email, username, full_name, role, is_active, created_at, updated_at`,
		email, passwordHash, username, fullName,
	).Scan(&u.ID, &u.Email, &u.Username, &u.FullName, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("repository.CreateUser: %w", err)
	}
	return u, nil
}

// GetGithubAccountByUserID retrieves the github_account for a user (for status check).
func (r *UserRepository) CreateRefreshToken(ctx context.Context, userID int, tokenHash, userAgent, ipAddress string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, user_agent, ip_address, expires_at) VALUES ($1,$2,$3,$4,$5)`,
		userID, tokenHash, userAgent, ipAddress, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("repository.CreateRefreshToken: %w", err)
	}
	return nil
}

func (r *UserRepository) FindRefreshToken(ctx context.Context, tokenHash string) (id string, userID int, expiresAt time.Time, revokedAt *time.Time, err error) {
	err = r.pool.QueryRow(ctx,
		`SELECT id, user_id, expires_at, revoked_at FROM refresh_tokens WHERE token_hash=$1`, tokenHash,
	).Scan(&id, &userID, &expiresAt, &revokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", 0, time.Time{}, nil, nil
	}
	if err != nil {
		return "", 0, time.Time{}, nil, fmt.Errorf("repository.FindRefreshToken: %w", err)
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

func (r *UserRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID int) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked_at=CURRENT_TIMESTAMP WHERE user_id=$1 AND revoked_at IS NULL`, userID)
	if err != nil {
		return fmt.Errorf("repository.RevokeAllUserRefreshTokens: %w", err)
	}
	return nil
}

func (r *UserRepository) GetGithubAccountByUserID(ctx context.Context, userID int) (installationID, owner, ownerType, status string, err error) {
	err = r.pool.QueryRow(ctx,
		`SELECT installation_id, github_owner, github_owner_type, status FROM github_accounts WHERE user_id = $1`,
		userID,
	).Scan(&installationID, &owner, &ownerType, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", "", "", nil
	}
	if err != nil {
		return "", "", "", "", fmt.Errorf("repository.GetGithubAccountByUserID: %w", err)
	}
	return installationID, owner, ownerType, status, nil
}
