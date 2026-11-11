// Package testutil provides isolated target-schema fixtures for integration tests.
package testutil

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"innogen-backend/shared/catalog"
	"os"
	"testing"
)

type Fixture struct {
	Pool                                     *pgxpool.Pool
	UserID, PackageID, ProblemID, LanguageID string
}

func Catalog(t *testing.T) Fixture {
	t.Helper()
	url := os.Getenv("CATALOG_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set CATALOG_TEST_DATABASE_URL to disposable PostgreSQL")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	f := Fixture{Pool: pool}
	var suffix string
	if err = pool.QueryRow(ctx, `SELECT gen_random_uuid()::text`).Scan(&suffix); err != nil {
		t.Fatal(err)
	}
	p, err := catalog.New(pool).CreatePackage(ctx, "fixture-"+suffix, "Fixture", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	f.PackageID = p.ID
	t.Cleanup(func() {
		for _, q := range []string{
			`DELETE FROM submission_test_results WHERE submission_id IN(SELECT id FROM submissions WHERE user_id=$1)`,
			`DELETE FROM user_problem_progress WHERE user_id=$1`, `DELETE FROM accepted_solutions WHERE user_id=$1`, `DELETE FROM submissions WHERE user_id=$1`, `DELETE FROM users WHERE id=$1`,
		} {
			if _, e := pool.Exec(ctx, q, f.UserID); e != nil {
				t.Error(e)
			}
		}
		for _, q := range []string{`UPDATE problems SET status='draft',published_at=NULL WHERE level_id IN(SELECT id FROM levels WHERE package_id=$1)`, `DELETE FROM problem_test_cases WHERE problem_id IN(SELECT id FROM problems WHERE level_id IN(SELECT id FROM levels WHERE package_id=$1))`, `DELETE FROM problems WHERE level_id IN(SELECT id FROM levels WHERE package_id=$1)`, `DELETE FROM levels WHERE package_id=$1`, `DELETE FROM packages WHERE id=$1`} {
			if _, e := pool.Exec(ctx, q, p.ID); e != nil {
				t.Error(e)
			}
		}
	})
	if err = pool.QueryRow(ctx, `INSERT INTO users(email,username,password_hash) VALUES($1,$2,'hash') RETURNING id`, suffix+"@example.org", suffix).Scan(&f.UserID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT id FROM languages WHERE is_active AND slug='python'`).Scan(&f.LanguageID); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `INSERT INTO problems(level_id,slug,title,statement_md,time_limit_ms) SELECT id,$2,'Fixture','Statement',1000 FROM levels WHERE package_id=$1 AND level_no=1 RETURNING id`, p.ID, "problem-"+suffix).Scan(&f.ProblemID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO problem_languages(problem_id,language_id) VALUES($1,$2)`, f.ProblemID, f.LanguageID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO code_templates(problem_id,language_id,starter_code,file_name) VALUES($1,$2,'print(3)','main.py')`, f.ProblemID, f.LanguageID); err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{`INSERT INTO problem_samples(problem_id,input_data,expected_output) VALUES($1,'1 2','3')`, `INSERT INTO problem_test_cases(problem_id,input_data,expected_output) VALUES($1,'HIDDEN INPUT','3')`, `UPDATE problems SET status='published',published_at=now() WHERE id=$1`} {
		if _, err = pool.Exec(ctx, q, f.ProblemID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = pool.Exec(ctx, `UPDATE packages SET status='published',published_at=now() WHERE id=$1`, p.ID); err != nil {
		t.Fatal(err)
	}
	return f
}
