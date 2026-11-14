package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"innogen-backend/shared/catalog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"innogen-backend/shared/models"
)

// ErrSpamCooldown is returned when the anti-spam trigger prevents a submission.
var ErrSpamCooldown = errors.New("Please wait 10 seconds before submitting again")

// SubmissionRepository handles database queries for the submissions table.
type SubmissionRepository struct {
	pool *pgxpool.Pool
}

// New creates a new SubmissionRepository.
func New(pool *pgxpool.Pool) *SubmissionRepository {
	return &SubmissionRepository{pool: pool}
}

const submissionColumnList = `id, user_id, problem_id, language_id, source_code, status, runtime_ms, memory_kb, error_message, pass_count, total_testcases, created_at, judged_at`

const selectAllSubmissionColumns = `SELECT ` + submissionColumnList + ` FROM submissions`

const submissionColumnListNoCode = `id, user_id, problem_id, language_id, status, runtime_ms, memory_kb, error_message, pass_count, total_testcases, created_at, judged_at`

const selectSubmissionColumnsNoCode = `SELECT ` + submissionColumnListNoCode + ` FROM submissions`

// CreateSubmission inserts a new Pending submission and returns it.
func (r *SubmissionRepository) CreateSubmission(ctx context.Context, userID string, problemID, languageID string, code string) (*models.Submission, error) {
	s := &models.Submission{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO submissions (user_id, problem_id, language_id, source_code,source_hash)
         VALUES ($1, $2, $3, $4,$5)
         RETURNING `+submissionColumnList,
		userID, problemID, languageID, code, sourceHash(code),
	).Scan(
		&s.ID, &s.UserID, &s.ProblemID, &s.LanguageID, &s.Code,
		&s.Status, &s.RuntimeMs, &s.MemoryKb, &s.ErrorMessage,
		&s.PassCount, &s.TotalTestcases,
		&s.CreatedAt, &s.JudgedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "Please wait 10 seconds") {
			return nil, ErrSpamCooldown
		}
		return nil, fmt.Errorf("repository.CreateSubmission: %w", err)
	}
	if err = r.EnrichSubmission(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

// FindByID looks up a submission by its UUID.
// Returns nil, nil if not found.
func (r *SubmissionRepository) FindByID(ctx context.Context, id string) (*models.Submission, error) {
	s := &models.Submission{}
	err := r.pool.QueryRow(ctx,
		`SELECT s.id, s.user_id, s.problem_id, s.language_id, s.source_code, s.status, s.runtime_ms, s.memory_kb, s.error_message, s.pass_count, s.total_testcases,
			s.created_at, s.judged_at
		FROM submissions s
		WHERE s.id = $1`,
		id,
	).Scan(
		&s.ID, &s.UserID, &s.ProblemID, &s.LanguageID, &s.Code,
		&s.Status, &s.RuntimeMs, &s.MemoryKb, &s.ErrorMessage,
		&s.PassCount, &s.TotalTestcases,
		&s.CreatedAt, &s.JudgedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.FindByID: %w", err)
	}
	if err = r.EnrichSubmission(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

// FindByUserID returns all submissions for a user, newest first, without code.
func (r *SubmissionRepository) FindByUserID(ctx context.Context, userID string) ([]models.Submission, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT s.id, s.user_id, s.problem_id, s.language_id, s.status, s.runtime_ms, s.memory_kb, s.error_message, s.pass_count, s.total_testcases,
			s.created_at, s.judged_at
		FROM submissions s
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
		LIMIT 50`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("repository.FindByUserID: %w", err)
	}
	defer rows.Close()

	var submissions []models.Submission
	for rows.Next() {
		var s models.Submission
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.ProblemID, &s.LanguageID,
			&s.Status, &s.RuntimeMs, &s.MemoryKb, &s.ErrorMessage,
			&s.PassCount, &s.TotalTestcases,
			&s.CreatedAt, &s.JudgedAt,
		); err != nil {
			return nil, fmt.Errorf("repository.FindByUserID: %w", err)
		}
		submissions = append(submissions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository.FindByUserID: %w", err)
	}
	return submissions, nil
}

// FindLatestByUserAndProblem returns the most recent submission for a user+problem pair.
// Returns nil, nil if none found.
func (r *SubmissionRepository) FindLatestByUserAndProblem(ctx context.Context, userID string, problemID string) (*models.Submission, error) {
	s := &models.Submission{}
	err := r.pool.QueryRow(ctx,
		`SELECT s.id, s.user_id, s.problem_id, s.language_id, s.source_code, s.status, s.runtime_ms, s.memory_kb, s.error_message, s.pass_count, s.total_testcases,
			s.created_at, s.judged_at
		FROM submissions s
		WHERE s.user_id = $1 AND s.problem_id = $2
		ORDER BY s.created_at DESC
		LIMIT 1`,
		userID, problemID,
	).Scan(
		&s.ID, &s.UserID, &s.ProblemID, &s.LanguageID, &s.Code,
		&s.Status, &s.RuntimeMs, &s.MemoryKb, &s.ErrorMessage,
		&s.PassCount, &s.TotalTestcases,
		&s.CreatedAt, &s.JudgedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.FindLatestByUserAndProblem: %w", err)
	}
	if err = r.EnrichSubmission(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

// ProblemExists checks whether a problem with the given ID exists.
func (r *SubmissionRepository) ProblemExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM problems WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("repository.ProblemExists: %w", err)
	}
	return exists, nil
}

// LanguageExists checks whether an active language with the given ID exists.
func (r *SubmissionRepository) LanguageExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM languages WHERE id = $1 AND is_active = true)`, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("repository.LanguageExists: %w", err)
	}
	return exists, nil
}

// GetProblemByID retrieves a full problem by ID.
// Returns nil, nil if not found.
func (r *SubmissionRepository) GetProblemByID(ctx context.Context, id string) (*models.Problem, error) {
	return catalog.New(r.pool).ExecutionProblem(ctx, id)
}
func (r *SubmissionRepository) GetLanguageByID(ctx context.Context, id string) (*models.Language, error) {
	return catalog.New(r.pool).ExecutionLanguage(ctx, id)
}
func (r *SubmissionRepository) GetTestCasesByProblemID(ctx context.Context, id string) ([]models.TestCase, error) {
	return catalog.New(r.pool).ExecutionTests(ctx, id, true)
}
func (r *SubmissionRepository) CheckExecution(ctx context.Context, uid, pid, lid string) error {
	return catalog.New(r.pool).CheckExecution(ctx, uid, pid, lid)
}
func (r *SubmissionRepository) LoadTemplate(ctx context.Context, p *models.Problem, lid string) error {
	return catalog.New(r.pool).ExecutionTemplate(ctx, p, lid)
}
func sourceHash(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

// UpdateSubmissionStatus updates the status of a submission.
func (r *SubmissionRepository) UpdateSubmissionStatus(ctx context.Context, id string, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE submissions SET status = $2,claimed_at=CASE WHEN $2::text='Running' THEN now() ELSE claimed_at END WHERE id = $1`, id, status)
	if err != nil {
		return fmt.Errorf("repository.UpdateSubmissionStatus: %w", err)
	}
	return nil
}

// UpdateSubmissionResult updates the final result of a judged submission.
func (r *SubmissionRepository) UpdateSubmissionResult(ctx context.Context, id string, attempt int, status string, runtimeMs *int, memoryKb *int, errorMessage *string, passCount int, totalTestcases int) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE submissions
         SET status = $2, runtime_ms = $3, memory_kb = $4, error_message = $5,
             pass_count = $6, total_testcases = $7, judged_at = CURRENT_TIMESTAMP
         WHERE id = $1 AND status='Running' AND attempt_count=$8`,
		id, status, runtimeMs, memoryKb, errorMessage, passCount, totalTestcases, attempt,
	)
	if err != nil {
		return fmt.Errorf("repository.UpdateSubmissionResult: %w", err)
	}
	if result.RowsAffected() != 1 {
		return errors.New("submission attempt no longer owns job")
	}
	return nil
}

// ClaimSubmission permits one worker to execute a Pending submission.
func (r *SubmissionRepository) ClaimSubmission(ctx context.Context, id string) (int, error) {
	var attempt int
	err := r.pool.QueryRow(ctx, `UPDATE submissions SET status='Running',claimed_at=now(),attempt_count=attempt_count+1 WHERE id=$1 AND status='Pending' RETURNING attempt_count`, id).Scan(&attempt)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return attempt, err
}

func (r *SubmissionRepository) EnrichSubmission(ctx context.Context, s *models.Submission) error {
	return r.pool.QueryRow(ctx, `SELECT p.slug,p.title,l.slug FROM problems p CROSS JOIN languages l WHERE p.id=$1 AND l.id=$2`, s.ProblemID, s.LanguageID).Scan(&s.ProblemSlug, &s.ProblemTitle, &s.LanguageSlug)
}
