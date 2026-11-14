package repository

import (
	"context"
	"innogen-backend/shared/testutil"
	"slices"
	"testing"
	"time"
)

func TestRecoveryFencesOldAttemptAndBoundsRetries(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.Background()
	repo := New(f.Pool)
	sub, err := repo.CreateSubmission(ctx, f.UserID, f.ProblemID, f.LanguageID, "print(3)")
	if err != nil {
		t.Fatal(err)
	}
	// A committed Pending row is recoverable even if its Redis enqueue was lost.
	ids, err := repo.Recover(ctx, 0, time.Minute, 3)
	if err != nil || !slices.Contains(ids, sub.ID) {
		t.Fatalf("lost enqueue: %v %v", ids, err)
	}
	attempt, err := repo.ClaimSubmission(ctx, sub.ID)
	if err != nil || attempt != 1 {
		t.Fatalf("claim %v %v", attempt, err)
	}
	again, err := repo.ClaimSubmission(ctx, sub.ID)
	if err != nil || again != 0 {
		t.Fatalf("duplicate %v %v", again, err)
	}
	if _, err = f.Pool.Exec(ctx, `UPDATE submissions SET claimed_at=now()-interval '5 minutes' WHERE id=$1`, sub.ID); err != nil {
		t.Fatal(err)
	}
	ids, err = repo.Recover(ctx, 0, time.Minute, 3)
	if err != nil || !slices.Contains(ids, sub.ID) {
		t.Fatalf("stale running: %v %v", ids, err)
	}
	newer, err := repo.ClaimSubmission(ctx, sub.ID)
	if err != nil || newer != 2 {
		t.Fatalf("reclaim %v %v", newer, err)
	}
	if err = repo.UpdateSubmissionResult(ctx, sub.ID, attempt, "Wrong Answer", nil, nil, nil, 0, 1); err == nil {
		t.Fatal("stale attempt overwrote active job")
	}
	if err = repo.UpdateSubmissionResult(ctx, sub.ID, newer, "Accepted", nil, nil, nil, 1, 1); err != nil {
		t.Fatal(err)
	}
	if err = repo.UpdateSubmissionResult(ctx, sub.ID, newer, "Wrong Answer", nil, nil, nil, 0, 1); err == nil {
		t.Fatal("final verdict mutable")
	}
	ids, err = repo.Recover(ctx, 0, time.Minute, 3)
	if err != nil || slices.Contains(ids, sub.ID) {
		t.Fatalf("terminal recovered: %v %v", ids, err)
	}
}

func TestRecoveryExhaustionAndLiveHeartbeat(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.Background()
	repo := New(f.Pool)
	sub, err := repo.CreateSubmission(ctx, f.UserID, f.ProblemID, f.LanguageID, "print(3)")
	if err != nil {
		t.Fatal(err)
	}
	attempt, err := repo.ClaimSubmission(ctx, sub.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = repo.Heartbeat(ctx, sub.ID, attempt); err != nil {
		t.Fatal(err)
	}
	ids, err := repo.Recover(ctx, 0, time.Minute, 3)
	if err != nil || slices.Contains(ids, sub.ID) {
		t.Fatalf("live job recovered: %v %v", ids, err)
	}
	if _, err = f.Pool.Exec(ctx, `UPDATE submissions SET claimed_at=now()-interval '5 minutes',attempt_count=3 WHERE id=$1`, sub.ID); err != nil {
		t.Fatal(err)
	}
	ids, err = repo.Recover(ctx, 0, time.Minute, 3)
	if err != nil || slices.Contains(ids, sub.ID) {
		t.Fatalf("exhausted job requeued: %v %v", ids, err)
	}
	got, err := repo.FindByID(ctx, sub.ID)
	if err != nil || got.Status != "Internal Error" {
		t.Fatalf("exhausted verdict: %+v %v", got, err)
	}
}
