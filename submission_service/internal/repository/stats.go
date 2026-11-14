package repository

import (
	"context"
	"github.com/jackc/pgx/v5"
	"innogen-backend/submission_service/internal/dto"
	"time"
)

// GetUserStats reads one snapshot of progress/rank and Accepted activity, never AI scores.
func (r *SubmissionRepository) GetUserStats(ctx context.Context, uid string) (*dto.UserStatsResponse, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	stats := &dto.UserStatsResponse{Activity: []dto.Activity{}}
	err = tx.QueryRow(ctx, `SELECT count(*),count(*) FILTER(WHERE l.difficulty='easy'),count(*) FILTER(WHERE l.difficulty='medium'),count(*) FILTER(WHERE l.difficulty='hard') FROM user_problem_progress upp JOIN problems p ON p.id=upp.problem_id JOIN levels l ON l.id=p.level_id WHERE upp.user_id=$1`, uid).Scan(&stats.SolvedCount.Total, &stats.SolvedCount.Easy, &stats.SolvedCount.Medium, &stats.SolvedCount.Hard)
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, `SELECT (COALESCE(judged_at,created_at) AT TIME ZONE 'UTC')::date,count(*) FROM submissions WHERE user_id=$1 AND status='Accepted' GROUP BY 1 ORDER BY 1`, uid)
	if err != nil {
		return nil, err
	}
	var dates []time.Time
	for rows.Next() {
		var d time.Time
		var count int
		if err = rows.Scan(&d, &count); err != nil {
			rows.Close()
			return nil, err
		}
		dates = append(dates, d)
		stats.Activity = append(stats.Activity, dto.Activity{Date: d.Format("2006-01-02"), Count: count})
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	stats.Streak = utcStreak(dates, time.Now())
	err = tx.QueryRow(ctx, `SELECT (SELECT rank::integer FROM leaderboard WHERE user_id=$1),(SELECT count(*) FROM users WHERE is_active)`, uid).Scan(&stats.Rank.CurrentRank, &stats.Rank.TotalUsers)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return stats, nil
}

func utcStreak(dates []time.Time, now time.Time) dto.Streak {
	result := dto.Streak{}
	if len(dates) == 0 {
		return result
	}
	last := dates[len(dates)-1].Format("2006-01-02")
	result.LastActiveDate = &last
	current := 1
	result.MaxStreak = 1
	for i := 1; i < len(dates); i++ {
		if dates[i].Sub(dates[i-1]) == 24*time.Hour {
			current++
		} else {
			current = 1
		}
		if current > result.MaxStreak {
			result.MaxStreak = current
		}
	}
	age := now.UTC().Truncate(24 * time.Hour).Sub(dates[len(dates)-1])
	if age >= 0 && age <= 24*time.Hour {
		result.CurrentStreak = current
	}
	return result
}
