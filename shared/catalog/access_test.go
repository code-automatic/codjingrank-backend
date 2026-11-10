package catalog

import (
	"context"
	"errors"
	"testing"
)

func TestExecutionPolicyRejectsUnpublishedAndInactive(t *testing.T) {
	pool := testPool(t)
	repo := New(pool)
	ctx := context.Background()
	var uid string
	if err := pool.QueryRow(ctx, `INSERT INTO users(email,username,password_hash) VALUES('access-test@example.org','access-test','hash') RETURNING id`).Scan(&uid); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, uid) })
	if err := repo.CheckExecution(ctx, uid, "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000000"); !errors.Is(err, ErrProblemNotFound) {
		t.Fatalf("missing: %v", err)
	}
	pool.Exec(ctx, `UPDATE users SET is_active=false WHERE id=$1`, uid)
	if err := repo.CheckExecution(ctx, uid, "00000000-0000-0000-0000-000000000000", "00000000-0000-0000-0000-000000000000"); !errors.Is(err, ErrAccountInactive) {
		t.Fatalf("inactive: %v", err)
	}
}
