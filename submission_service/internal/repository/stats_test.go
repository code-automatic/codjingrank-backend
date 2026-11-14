package repository

import (
	"context"
	"innogen-backend/shared/testutil"
	"testing"
	"time"
)

func TestTargetStatsDistinctProgressAndUTCActivity(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.Background()
	repo := New(f.Pool)
	stats, err := repo.GetUserStats(ctx, f.UserID)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Rank.CurrentRank != nil || stats.SolvedCount.Total != 0 || stats.Activity == nil || len(stats.Activity) != 0 {
		t.Fatalf("empty stats: %+v", stats)
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	for i, code := range []string{"a", "a", "b"} {
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
		if e = repo.UpdateSubmissionResult(ctx, sub.ID, a, "Accepted", nil, nil, nil, 1, 1); e != nil {
			t.Fatal(e)
		}
		at := today.Add(time.Minute)
		if i == 0 {
			at = today.Add(-time.Minute)
		}
		if _, e = f.Pool.Exec(ctx, `UPDATE submissions SET judged_at=$2 WHERE id=$1`, sub.ID, at); e != nil {
			t.Fatal(e)
		}
	}
	stats, err = repo.GetUserStats(ctx, f.UserID)
	if err != nil {
		t.Fatal(err)
	}
	if stats.SolvedCount.Total != 1 || stats.SolvedCount.Easy != 1 || stats.Rank.CurrentRank == nil || len(stats.Activity) != 2 || stats.Activity[0].Count != 1 || stats.Activity[1].Count != 2 || stats.Activity[1].Date != today.Format("2006-01-02") || stats.Streak.CurrentStreak != 2 {
		t.Fatalf("stats %+v", stats)
	}
	var count int
	if err = f.Pool.QueryRow(ctx, `SELECT accepted_count FROM user_problem_progress WHERE user_id=$1`, f.UserID).Scan(&count); err != nil || count != 2 {
		t.Fatalf("distinct Accepted count %d %v", count, err)
	}
}

func TestUTCStreakGapsAndFutureDates(t *testing.T) {
	day := func(s string) time.Time {
		d, e := time.Parse("2006-01-02", s)
		if e != nil {
			t.Fatal(e)
		}
		return d
	}
	for _, tc := range []struct {
		dates        []time.Time
		now          string
		current, max int
	}{
		{nil, "2026-11-14", 0, 0},
		{[]time.Time{day("2026-11-10"), day("2026-11-11"), day("2026-11-13")}, "2026-11-14", 1, 2},
		{[]time.Time{day("2026-11-10"), day("2026-11-11")}, "2026-11-14", 0, 2},
		{[]time.Time{day("2026-11-15")}, "2026-11-14", 0, 1},
	} {
		got := utcStreak(tc.dates, day(tc.now))
		if got.CurrentStreak != tc.current || got.MaxStreak != tc.max {
			t.Fatalf("streak %+v", got)
		}
	}
}
