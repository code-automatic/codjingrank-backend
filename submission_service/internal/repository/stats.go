package repository

import (
	"context"
	"time"

	"innogen-backend/submission_service/internal/dto"
)

// GetUserStats calculates dashboard statistics for a user.
func (r *SubmissionRepository) GetUserStats(ctx context.Context, userID int) (*dto.UserStatsResponse, error) {
	stats := &dto.UserStatsResponse{
		Activity: []dto.Activity{},
	}

	// 1. Solved Count by difficulty (distinct accepted problems)
	solvedQuery := `
		SELECT p.difficulty, COUNT(DISTINCT s.problem_id)
		FROM submissions s
		JOIN problems p ON s.problem_id = p.id
		WHERE s.user_id = $1 AND s.status = 'Accepted'
		GROUP BY p.difficulty
	`
	rows, err := r.pool.Query(ctx, solvedQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var difficulty string
		var count int
		if err := rows.Scan(&difficulty, &count); err != nil {
			return nil, err
		}
		stats.SolvedCount.Total += count
		switch difficulty {
		case "Easy":
			stats.SolvedCount.Easy = count
		case "Medium":
			stats.SolvedCount.Medium = count
		case "Hard":
			stats.SolvedCount.Hard = count
		}
	}

	// 2. Activity (accepted submissions grouped by date)
	activityQuery := `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM submissions
		WHERE user_id = $1 AND status = 'Accepted'
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`
	actRows, err := r.pool.Query(ctx, activityQuery, userID)
	if err != nil {
		return nil, err
	}
	defer actRows.Close()

	var dates []time.Time
	for actRows.Next() {
		var d time.Time
		var c int
		if err := actRows.Scan(&d, &c); err != nil {
			return nil, err
		}
		dates = append(dates, d)
		stats.Activity = append(stats.Activity, dto.Activity{
			Date:  d.Format("2006-01-02"),
			Count: c,
		})
	}

	// 3. Streak Calculation
	currentStreak := 0
	maxStreak := 0

	if len(dates) > 0 {
		lastDate := dates[len(dates)-1].Format("2006-01-02")
		stats.Streak.LastActiveDate = &lastDate

		curr := 1
		maxStreak = 1
		for i := 1; i < len(dates); i++ {
			diff := int(dates[i].Sub(dates[i-1]).Hours() / 24)
			if diff == 1 {
				curr++
			} else if diff > 1 {
				curr = 1
			}
			if curr > maxStreak {
				maxStreak = curr
			}
		}

		// If the last active date is today or yesterday, they have an active streak
		today := time.Now().Truncate(24 * time.Hour)
		daysSinceLastActive := int(today.Sub(dates[len(dates)-1]).Hours() / 24)
		if daysSinceLastActive <= 1 {
			currentStreak = curr
		} else {
			currentStreak = 0
		}
	}
	stats.Streak.CurrentStreak = currentStreak
	stats.Streak.MaxStreak = maxStreak

	// 4. Rank (MVP: Rank based on total distinct problems solved compared to others)
	// We count users who have more distinct accepted problems than this user.
	// Total users simply counts distinct user_ids in submissions table.
	rankQuery := `
		WITH UserSolvedCounts AS (
			SELECT user_id, COUNT(DISTINCT problem_id) as solved_count
			FROM submissions
			WHERE status = 'Accepted'
			GROUP BY user_id
		)
		SELECT 
			(SELECT COUNT(*) + 1 FROM UserSolvedCounts WHERE solved_count > $1) as current_rank,
			(SELECT COUNT(DISTINCT id) FROM users) as total_users
	`
	var currentRank, totalUsers int
	if err := r.pool.QueryRow(ctx, rankQuery, stats.SolvedCount.Total).Scan(&currentRank, &totalUsers); err != nil {
		// totalUsers might fail if we query the users table?
		// Wait, submission_service has access to users table (shared database).
		// Let's handle it safely.
		return nil, err
	}

	stats.Rank.CurrentRank = currentRank
	stats.Rank.TotalUsers = totalUsers
	stats.Rank.Rating = 1200 + (stats.SolvedCount.Total * 10) // Simple MVP rating

	return stats, nil
}
