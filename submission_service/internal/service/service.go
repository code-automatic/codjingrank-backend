package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"innogen-backend/shared/models"

	"innogen-backend/submission_service/internal/dto"
	"innogen-backend/submission_service/internal/queue"
	"innogen-backend/submission_service/internal/repository"
)

// Sentinel errors for the service layer.
var (
	ErrInvalidInput           = errors.New("invalid input")
	ErrProblemNotFound        = errors.New("problem not found")
	ErrLanguageNotFound       = errors.New("language not found")
	ErrSubmissionNotFound     = errors.New("submission not found")
	ErrNoSubmissionForProblem = errors.New("no submission found for this problem")
)

// SubmissionService handles submission business logic.
type SubmissionService struct {
	repo  *repository.SubmissionRepository
	queue *queue.Queue
	log   *slog.Logger
}

// New creates a new SubmissionService.
func New(repo *repository.SubmissionRepository, q *queue.Queue, log *slog.Logger) *SubmissionService {
	return &SubmissionService{repo: repo, queue: q, log: log}
}

// CreateSubmission validates inputs and creates a new Pending submission.
func (s *SubmissionService) CreateSubmission(ctx context.Context, userID int, req dto.SubmitRequest) (*models.Submission, error) {
	if _, msg := req.Validate(); msg != "" {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, msg)
	}

	exists, err := s.repo.ProblemExists(ctx, req.ProblemID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrProblemNotFound
	}

	exists, err = s.repo.LanguageExists(ctx, req.LanguageID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrLanguageNotFound
	}

	submission, err := s.repo.CreateSubmission(ctx, userID, req.ProblemID, req.LanguageID, req.Code)
	if err != nil {
		return nil, err
	}

	// Enqueue for async judging (best effort — if enqueue fails, submission stays Pending)
	if err := s.queue.Enqueue(ctx, submission.ID); err != nil {
		s.log.Error("failed to enqueue submission", slog.String("submissionId", submission.ID), slog.String("error", err.Error()))
		// Continue — submission is created with Pending status
	}

	return submission, nil
}

// GetByID retrieves a submission by its UUID.
func (s *SubmissionService) GetByID(ctx context.Context, id string) (*models.Submission, error) {
	sub, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, ErrSubmissionNotFound
	}
	return sub, nil
}

// ListByUserID returns all submissions for a user, newest first.
func (s *SubmissionService) ListByUserID(ctx context.Context, userID int) ([]dto.SubmissionListItem, error) {
	subs, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]dto.SubmissionListItem, 0, len(subs))
	for i := range subs {
		items = append(items, dto.ToSubmissionListItem(&subs[i]))
	}
	return items, nil
}

// GetLatestByUserAndProblem returns the most recent submission for a user+problem pair.
func (s *SubmissionService) GetLatestByUserAndProblem(ctx context.Context, userID, problemID int) (*models.Submission, error) {
	sub, err := s.repo.FindLatestByUserAndProblem(ctx, userID, problemID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return nil, ErrNoSubmissionForProblem
	}
	return sub, nil
}
