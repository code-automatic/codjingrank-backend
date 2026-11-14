package repository

import (
	"context"
	"innogen-backend/shared/catalog"
	"innogen-backend/shared/testutil"
	"sync"
	"testing"
)

func TestAcceptedProgressDeduplicatesAndUnlocks(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.Background()
	repo := New(f.Pool)
	levels, err := catalog.New(f.Pool).Levels(ctx, f.PackageID, f.UserID)
	if err != nil || levels[1].Unlocked {
		t.Fatalf("premature unlock: %+v %v", levels, err)
	}
	for i, code := range []string{"print(3)", "print(3)", "print(1+2)", "print(0)"} {
		sub, e := repo.CreateSubmission(ctx, f.UserID, f.ProblemID, f.LanguageID, code)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = f.Pool.Exec(ctx, `UPDATE submissions SET created_at=now()-interval '11 seconds' WHERE id=$1`, sub.ID); e != nil {
			t.Fatal(e)
		}
		attempt, e := repo.ClaimSubmission(ctx, sub.ID)
		if e != nil {
			t.Fatal(e)
		}
		status, passed := "Accepted", 1
		if i == 3 {
			status, passed = "Wrong Answer", 0
		}
		runtime := 12 - i
		if e = repo.UpdateSubmissionResult(ctx, sub.ID, attempt, status, &runtime, nil, nil, passed, 1); e != nil {
			t.Fatal(e)
		}
	}
	var sources, count, best int
	if err = f.Pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM accepted_solutions WHERE user_id=$1),accepted_count,best_runtime_ms FROM user_problem_progress WHERE user_id=$1 AND problem_id=$2`, f.UserID, f.ProblemID).Scan(&sources, &count, &best); err != nil {
		t.Fatal(err)
	}
	if sources != 2 || count != 2 || best != 10 {
		t.Fatalf("distinct sources=%d progress=%d best=%d", sources, count, best)
	}
	levels, err = catalog.New(f.Pool).Levels(ctx, f.PackageID, f.UserID)
	if err != nil || !levels[1].Unlocked || levels[0].Solved != 1 {
		t.Fatalf("Accepted unlock: %+v %v", levels, err)
	}
}

func TestConcurrentAcceptedAndTransactionalRollback(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.Background()
	repo := New(f.Pool)
	ids := []string{}
	attempts := []int{}
	for _, code := range []string{"a", "b", "c"} {
		sub, e := repo.CreateSubmission(ctx, f.UserID, f.ProblemID, f.LanguageID, code)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = f.Pool.Exec(ctx, `UPDATE submissions SET created_at=now()-interval '11 seconds' WHERE id=$1`, sub.ID); e != nil {
			t.Fatal(e)
		}
		a, e := repo.ClaimSubmission(ctx, sub.ID)
		if e != nil {
			t.Fatal(e)
		}
		ids = append(ids, sub.ID)
		attempts = append(attempts, a)
	}
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if e := repo.UpdateSubmissionResult(ctx, ids[i], attempts[i], "Accepted", nil, nil, nil, 1, 1); e != nil {
				t.Error(e)
			}
		}(i)
	}
	wg.Wait()
	var count int
	if e := f.Pool.QueryRow(ctx, `SELECT accepted_count FROM user_problem_progress WHERE user_id=$1`, f.UserID).Scan(&count); e != nil || count != 2 {
		t.Fatalf("concurrent count %d %v", count, e)
	}
	if _, err := f.Pool.Exec(ctx, `UPDATE user_problem_progress SET accepted_count=2147483647 WHERE user_id=$1`, f.UserID); err != nil {
		t.Fatal(err)
	}
	if err := repo.UpdateSubmissionResult(ctx, ids[2], attempts[2], "Accepted", nil, nil, nil, 1, 1); err == nil {
		t.Fatal("expected progress overflow")
	}
	sub, e := repo.FindByID(ctx, ids[2])
	if e != nil || sub.Status != "Running" {
		t.Fatalf("partial verdict committed: %+v %v", sub, e)
	}
	if e = f.Pool.QueryRow(ctx, `SELECT count(*) FROM accepted_solutions WHERE submission_id=$1`, ids[2]).Scan(&count); e != nil || count != 0 {
		t.Fatalf("partial solution committed: %d %v", count, e)
	}
}

func TestConcurrentIdenticalAcceptedCountsOnce(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.Background()
	repo := New(f.Pool)
	ids := []string{}
	attempts := []int{}
	for range 2 {
		sub, err := repo.CreateSubmission(ctx, f.UserID, f.ProblemID, f.LanguageID, "print(3)")
		if err != nil {
			t.Fatal(err)
		}
		if _, err = f.Pool.Exec(ctx, `UPDATE submissions SET created_at=now()-interval '11 seconds' WHERE id=$1`, sub.ID); err != nil {
			t.Fatal(err)
		}
		attempt, err := repo.ClaimSubmission(ctx, sub.ID)
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, sub.ID)
		attempts = append(attempts, attempt)
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range ids {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			if err := repo.UpdateSubmissionResult(ctx, ids[i], attempts[i], "Accepted", nil, nil, nil, 1, 1); err != nil {
				t.Error(err)
			}
		}(i)
	}
	close(start)
	wg.Wait()
	var sources, count, verdicts int
	if err := f.Pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM accepted_solutions WHERE user_id=$1),accepted_count,(SELECT count(*) FROM submissions WHERE user_id=$1 AND status='Accepted') FROM user_problem_progress WHERE user_id=$1 AND problem_id=$2`, f.UserID, f.ProblemID).Scan(&sources, &count, &verdicts); err != nil {
		t.Fatal(err)
	}
	if sources != 1 || count != 1 || verdicts != 2 {
		t.Fatalf("sources=%d progress=%d Accepted verdicts=%d", sources, count, verdicts)
	}
}
