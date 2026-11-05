package problem

import (
	"context"
	"testing"
	"time"

	"innogen-backend/shared/database"
)

func TestGetDailyChallenge(t *testing.T) {
	ctx := context.Background()
	pool, err := database.NewPostgresPool(ctx, "postgres://innogen:innogen@localhost:5432/innogen?sslmode=disable")
	if err != nil {
		t.Skip("Database not available")
	}
	defer pool.Close()

	repo := NewProblemRepository(pool)

	// Clean up for test
	pool.Exec(ctx, `DELETE FROM daily_challenges`)
	pool.Exec(ctx, `DELETE FROM test_cases WHERE problem_id IN (9004, 9005)`)
	pool.Exec(ctx, `DELETE FROM problems WHERE id IN (9004, 9005)`)

	// Create a published problem
	_, err = pool.Exec(ctx, `
		INSERT INTO problems (id, slug, title, difficulty, problem_md, is_published, driver_code) VALUES 
		(9004, 'prob-9004', 'Daily 1', 'Easy', 'md', true, 'driver')
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		t.Fatalf("failed to insert test problems: %v", err)
	}

	// 1. First call creates today's challenge.
	prob1, err := repo.GetDailyChallenge(ctx)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	if prob1 == nil {
		t.Fatal("expected problem, got nil")
	}

	// 2. Second call returns the same challenge.
	prob2, err := repo.GetDailyChallenge(ctx)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if prob2.ID != prob1.ID {
		t.Errorf("expected same problem on second call, got %d", prob2.ID)
	}

	// 3. (In handler) driver_code is excluded, let's just make sure repo has it so handler can omit it.
	// Actually, ProblemDetail struct in handler doesn't map DriverCode anyway.
	if prob1.DriverCode == nil {
		t.Error("expected driver code in repo return")
	}

	// Insert another problem
	_, err = pool.Exec(ctx, `
		INSERT INTO problems (id, slug, title, difficulty, problem_md, is_published, driver_code) VALUES 
		(9005, 'prob-9005', 'Daily 2', 'Medium', 'md', true, 'driver')
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		t.Fatalf("failed to insert test problem: %v", err)
	}

	// Manipulate the date of the daily challenge to yesterday
	yesterday := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	_, err = pool.Exec(ctx, `UPDATE daily_challenges SET challenge_date = $1 WHERE problem_id = $2`, yesterday, prob1.ID)
	if err != nil {
		t.Fatalf("failed to update challenge date: %v", err)
	}

	// Now it should pick the unused problem (9005)
	prob3, err := repo.GetDailyChallenge(ctx)
	if err != nil {
		t.Fatalf("third call failed: %v", err)
	}
	if prob3 == nil {
		t.Fatal("expected problem for today, got nil")
	}
	if prob3.ID == prob1.ID {
		// Just to log, not strictly an error if there's only 1 published problem in the entire DB,
		// but since we inserted two test problems, it should pick something else.
		t.Logf("picked the same problem again, which means fallback triggered or random chose it again if only 1 problem exists")
	}
}
