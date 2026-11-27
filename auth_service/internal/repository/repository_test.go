package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"innogen-backend/auth_service/internal/repository"
	"innogen-backend/shared/testutil"
)

// TestCompletePasswordResetConsumesOnceAndRevokesTokens verifies the reset is
// atomic from the caller's perspective: the code is consumed, the password hash
// changes, every active refresh token is revoked, and replay returns false.
func TestCompletePasswordResetConsumesOnceAndRevokesTokens(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.Background()
	repo := repository.New(f.Pool)

	var resetID string
	if err := f.Pool.QueryRow(ctx,
		`INSERT INTO password_resets (user_id, otp_hash, expires_at) VALUES ($1,'otp-hash',now()+interval '10 minutes') RETURNING id`,
		f.UserID,
	).Scan(&resetID); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, err := f.Pool.Exec(ctx,
			`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1,$2,now()+interval '1 hour')`,
			f.UserID, fmt.Sprintf("token-%s-%d", resetID, i),
		); err != nil {
			t.Fatal(err)
		}
	}

	ok, err := repo.CompletePasswordReset(ctx, resetID, f.UserID, "new-password-hash")
	if err != nil {
		t.Fatalf("complete password reset: %v", err)
	}
	if !ok {
		t.Fatal("first completion returned false")
	}

	var passwordHash string
	if err = f.Pool.QueryRow(ctx, `SELECT password_hash FROM users WHERE id=$1`, f.UserID).Scan(&passwordHash); err != nil {
		t.Fatal(err)
	}
	if passwordHash != "new-password-hash" {
		t.Fatalf("password_hash = %q, want new-password-hash", passwordHash)
	}

	var usedAt *time.Time
	if err = f.Pool.QueryRow(ctx, `SELECT used_at FROM password_resets WHERE id=$1`, resetID).Scan(&usedAt); err != nil {
		t.Fatal(err)
	}
	if usedAt == nil {
		t.Fatal("reset code was not marked used")
	}

	var active int
	if err = f.Pool.QueryRow(ctx,
		`SELECT count(*) FROM refresh_tokens WHERE user_id=$1 AND revoked_at IS NULL`, f.UserID,
	).Scan(&active); err != nil {
		t.Fatal(err)
	}
	if active != 0 {
		t.Fatalf("active refresh tokens = %d, want 0", active)
	}

	ok, err = repo.CompletePasswordReset(ctx, resetID, f.UserID, "another-hash")
	if err != nil {
		t.Fatalf("second completion: %v", err)
	}
	if ok {
		t.Fatal("replayed reset code was accepted")
	}
}

// TestCreatePasswordResetInvalidatesPreviousCodes verifies a new OTP marks the
// user's outstanding unused codes as used so only the freshest one works.
func TestCreatePasswordResetInvalidatesPreviousCodes(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.Background()
	repo := repository.New(f.Pool)

	var oldID string
	if err := f.Pool.QueryRow(ctx,
		`INSERT INTO password_resets (user_id, otp_hash, expires_at) VALUES ($1,'old-hash',now()+interval '10 minutes') RETURNING id`,
		f.UserID,
	).Scan(&oldID); err != nil {
		t.Fatal(err)
	}

	if err := repo.CreatePasswordReset(ctx, f.UserID, "new-hash", time.Now().Add(10*time.Minute)); err != nil {
		t.Fatalf("create password reset: %v", err)
	}

	var usedAt *time.Time
	if err := f.Pool.QueryRow(ctx, `SELECT used_at FROM password_resets WHERE id=$1`, oldID).Scan(&usedAt); err != nil {
		t.Fatal(err)
	}
	if usedAt == nil {
		t.Fatal("previous reset code was not invalidated")
	}

	var unused int
	if err := f.Pool.QueryRow(ctx,
		`SELECT count(*) FROM password_resets WHERE user_id=$1 AND used_at IS NULL`, f.UserID,
	).Scan(&unused); err != nil {
		t.Fatal(err)
	}
	if unused != 1 {
		t.Fatalf("unused reset codes = %d, want 1", unused)
	}
}
