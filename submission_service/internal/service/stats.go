package service

import (
	"context"

	"innogen-backend/submission_service/internal/dto"
)

// GetUserStats retrieves dashboard statistics for the specified user.
func (s *SubmissionService) GetUserStats(ctx context.Context, userID int) (*dto.UserStatsResponse, error) {
	return s.repo.GetUserStats(ctx, userID)
}
