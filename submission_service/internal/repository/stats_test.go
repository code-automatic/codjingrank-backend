package repository

import (
	"context"
	"testing"
	"time"

	"innogen-backend/shared/database"
)

func TestGetUserStats(t *testing.T) {
	ctx := context.Background()
	pool, err := database.NewPostgresPool(ctx, "postgres://innogen:innogen@localhost:5432/innogen?sslmode=disable")
	if err != nil {
		t.Skip("Database not available")
	}
	defer pool.Close()

	repo := New(pool)

	// Use an isolated user ID for testing
	userID := 99999

	// Ensure the user exists
	_, err = pool.Exec(ctx, `INSERT INTO users (id, username, email, password) VALUES ($1, 'testuserstats', 'stats@test.com', 'hash') ON CONFLICT (id) DO NOTHING`, userID)
	if err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}

	// Create some problems
	_, err = pool.Exec(ctx, `
		INSERT INTO problems (id, slug, title, difficulty, problem_md) VALUES 
		(9001, 'prob-9001', 'P1', 'Easy', 'md'),
		(9002, 'prob-9002', 'P2', 'Medium', 'md'),
		(9003, 'prob-9003', 'P3', 'Hard', 'md')
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		t.Fatalf("failed to insert test problems: %v", err)
	}

	// Ensure clean state for this user
	pool.Exec(ctx, `DELETE FROM submissions WHERE user_id = $1`, userID)

	// 1. user with no submissions returns zeros and empty activity.
	stats, err := repo.GetUserStats(ctx, userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if stats.SolvedCount.Total != 0 {
		t.Errorf("expected 0 total solved, got %d", stats.SolvedCount.Total)
	}
	if len(stats.Activity) != 0 {
		t.Errorf("expected 0 activity, got %d", len(stats.Activity))
	}
	if stats.Streak.CurrentStreak != 0 || stats.Streak.MaxStreak != 0 {
		t.Errorf("expected 0 streaks, got %d/%d", stats.Streak.CurrentStreak, stats.Streak.MaxStreak)
	}

	// 2. duplicate accepted submissions for same problem count once.
	// 3. wrong answer submissions do not count solved.
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	// Insert WA
	_, err = pool.Exec(ctx, `INSERT INTO submissions (user_id, problem_id, language_id, status, code, created_at) VALUES ($1, $2, $3, 'Wrong Answer', 'x', $4)`, userID, 9001, 1, yesterday)

	// Insert AC (duplicate for 9001)
	_, err = pool.Exec(ctx, `INSERT INTO submissions (user_id, problem_id, language_id, status, code, created_at) VALUES ($1, $2, $3, 'Accepted', 'x', $4)`, userID, 9001, 1, yesterday)
	_, err = pool.Exec(ctx, `INSERT INTO submissions (user_id, problem_id, language_id, status, code, created_at) VALUES ($1, $2, $3, 'Accepted', 'x', $4)`, userID, 9001, 1, now)

	// Insert AC for medium
	_, err = pool.Exec(ctx, `INSERT INTO submissions (user_id, problem_id, language_id, status, code, created_at) VALUES ($1, $2, $3, 'Accepted', 'x', $4)`, userID, 9002, 1, now)

	if err != nil {
		t.Fatalf("failed to insert submissions: %v", err)
	}

	stats, err = repo.GetUserStats(ctx, userID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if stats.SolvedCount.Total != 2 {
		t.Errorf("expected 2 total solved, got %d", stats.SolvedCount.Total)
	}
	if stats.SolvedCount.Easy != 1 {
		t.Errorf("expected 1 easy solved, got %d", stats.SolvedCount.Easy)
	}
	if stats.SolvedCount.Medium != 1 {
		t.Errorf("expected 1 medium solved, got %d", stats.SolvedCount.Medium)
	}

	// 4. activity groups by date.
	if len(stats.Activity) != 2 {
		t.Errorf("expected 2 activity days, got %d", len(stats.Activity))
	} else {
		// yesterday should have count 1 (only 1 accepted)
		// today should have count 2 (2 accepted)
		if stats.Activity[0].Count != 1 {
			t.Errorf("expected yesterday activity count 1, got %d", stats.Activity[0].Count)
		}
		if stats.Activity[1].Count != 2 {
			t.Errorf("expected today activity count 2, got %d", stats.Activity[1].Count)
		}
	}

	// 5. streak calculation handles gaps.
	if stats.Streak.CurrentStreak != 2 {
		t.Errorf("expected current streak 2, got %d", stats.Streak.CurrentStreak)
	}
	if stats.Streak.MaxStreak != 2 {
		t.Errorf("expected max streak 2, got %d", stats.Streak.MaxStreak)
	}
}
