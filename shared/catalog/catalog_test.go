package catalog

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("CATALOG_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set CATALOG_TEST_DATABASE_URL to a disposable bootstrapped database")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestPackageCreationAndDynamicProgress(t *testing.T) {
	ctx := context.Background()
	pool := testPool(t)
	repo := New(pool)
	pkg, err := repo.CreatePackage(ctx, "sprint-two", "Sprint two", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM levels WHERE package_id=$1", pkg.ID)
		pool.Exec(ctx, "DELETE FROM packages WHERE id=$1", pkg.ID)
	})
	levels, err := repo.Levels(ctx, pkg.ID, "00000000-0000-0000-0000-000000000001")
	if err != nil {
		t.Fatal(err)
	}
	if len(levels) != 3 {
		t.Fatalf("want three levels, got %d", len(levels))
	}
	for i, l := range levels {
		if l.LevelNo != i+1 || !l.Unlocked || l.Total != 0 || l.Solved != 0 || !l.Completed {
			t.Fatalf("empty level must not block: %+v", l)
		}
	}
	// A published unsolved problem blocks every later level, including across an empty level.
	var problemID, userID string
	if err = pool.QueryRow(ctx, `INSERT INTO users(email,username,password_hash) VALUES('progress-test@example.org','progress-test','hash') RETURNING id`).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM user_problem_progress WHERE user_id=$1", userID)
		pool.Exec(ctx, "DELETE FROM problems WHERE id=$1", problemID)
		pool.Exec(ctx, "DELETE FROM users WHERE id=$1", userID)
	})
	if err = pool.QueryRow(ctx, `INSERT INTO problems(level_id,slug,title,statement_md) VALUES($1,'progress-test','Progress','Statement') RETURNING id`, levels[0].ID).Scan(&problemID); err != nil {
		t.Fatal(err)
	}
	for _, sql := range []string{
		`INSERT INTO problem_languages SELECT $1,id FROM languages LIMIT 1`,
		`INSERT INTO problem_samples(problem_id,expected_output) VALUES($1,'1')`,
		`INSERT INTO problem_test_cases(problem_id,expected_output) VALUES($1,'1')`,
		`UPDATE problems SET status='published',published_at=now() WHERE id=$1`,
	} {
		if _, err = pool.Exec(ctx, sql, problemID); err != nil {
			t.Fatal(err)
		}
	}
	levels, err = repo.Levels(ctx, pkg.ID, userID)
	if err != nil {
		t.Fatal(err)
	}
	if levels[0].Completed || levels[1].Unlocked || levels[2].Unlocked {
		t.Fatalf("unsolved prerequisite leaked unlock: %+v", levels)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO user_problem_progress(user_id,problem_id,first_solved_at,last_solved_at) VALUES($1,$2,now(),now())`, userID, problemID); err != nil {
		t.Fatal(err)
	}
	levels, err = repo.Levels(ctx, pkg.ID, userID)
	if err != nil {
		t.Fatal(err)
	}
	if levels[0].Solved != 1 || !levels[2].Unlocked {
		t.Fatalf("persisted progress did not unlock: %+v", levels)
	}
	// Cleanup published fixtures requires deleting their dependent rows first.
	t.Cleanup(func() {
		pool.Exec(ctx, "UPDATE problems SET status='draft',published_at=NULL WHERE id=$1", problemID)
		pool.Exec(ctx, "DELETE FROM problem_test_cases WHERE problem_id=$1", problemID)
		pool.Exec(ctx, "DELETE FROM problem_samples WHERE problem_id=$1", problemID)
		pool.Exec(ctx, "DELETE FROM problem_languages WHERE problem_id=$1", problemID)
	})
	if _, err = repo.CreatePackage(ctx, "sprint-two", "Duplicate", nil, 0); err == nil {
		t.Fatal("duplicate slug accepted")
	}
	var count int
	if err = pool.QueryRow(ctx, "SELECT count(*) FROM levels WHERE package_id=$1", pkg.ID).Scan(&count); err != nil || count != 3 {
		t.Fatalf("failed create changed levels: %d %v", count, err)
	}
}
