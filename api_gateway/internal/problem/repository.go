package problem

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"innogen-backend/shared/models"
)

// ProblemRepository provides database access for problem entities.
type ProblemRepository struct {
	pool *pgxpool.Pool
}

// NewProblemRepository creates a new ProblemRepository.
func NewProblemRepository(pool *pgxpool.Pool) *ProblemRepository {
	return &ProblemRepository{pool: pool}
}

// FindBySlug returns a single published problem by slug.
// Returns nil, nil if not found.
func (r *ProblemRepository) FindBySlug(ctx context.Context, slug string) (*models.Problem, error) {
	query := `
		SELECT id, slug, title, difficulty, problem_md, time_limit_ms, memory_limit_mb,
		       acceptance_rate, is_published, execution_mode, function_name, initial_code, driver_code, solution_file_name, sample_test_cases, created_at, updated_at
		FROM problems
		WHERE slug = $1 AND is_published = true
	`

	var p models.Problem
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&p.ID, &p.Slug, &p.Title, &p.Difficulty, &p.ProblemMD,
		&p.TimeLimitMs, &p.MemoryLimitMb, &p.AcceptanceRate,
		&p.IsPublished, &p.ExecutionMode, &p.FunctionName, &p.InitialCode, &p.DriverCode, &p.SolutionFileName, &p.SampleTestCases, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find problem by slug %q: %w", slug, err)
	}

	return &p, nil
}

// FindByID returns a single published problem by id.
// Returns nil, nil if not found.
func (r *ProblemRepository) FindByID(ctx context.Context, id int) (*models.Problem, error) {
	query := `
		SELECT id, slug, title, difficulty, problem_md, time_limit_ms, memory_limit_mb,
		       acceptance_rate, is_published, execution_mode, function_name, initial_code, driver_code, solution_file_name, sample_test_cases, created_at, updated_at
		FROM problems
		WHERE id = $1 AND is_published = true
	`

	var p models.Problem
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Slug, &p.Title, &p.Difficulty, &p.ProblemMD,
		&p.TimeLimitMs, &p.MemoryLimitMb, &p.AcceptanceRate,
		&p.IsPublished, &p.ExecutionMode, &p.FunctionName, &p.InitialCode, &p.DriverCode, &p.SolutionFileName, &p.SampleTestCases, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find problem by id %d: %w", id, err)
	}

	return &p, nil
}

// FindTestCasesByProblemID returns test cases for a given problem filtered by visibility.
// For public API usage, visibility should be "sample".
// Hidden test cases are never returned unless visibility="hidden" is explicitly passed.
func (r *ProblemRepository) FindTestCasesByProblemID(ctx context.Context, problemID int, visibility string) ([]models.TestCase, error) {
	query := `
		SELECT id, problem_id, visibility, input_data, expected_output, order_index
		FROM test_cases
		WHERE problem_id = $1 AND visibility = $2
		ORDER BY order_index ASC
	`

	rows, err := r.pool.Query(ctx, query, problemID, visibility)
	if err != nil {
		return nil, fmt.Errorf("find test cases by problem id %d: %w", problemID, err)
	}
	defer rows.Close()

	var cases []models.TestCase
	for rows.Next() {
		var tc models.TestCase
		if err := rows.Scan(
			&tc.ID, &tc.ProblemID, &tc.Visibility, &tc.InputData, &tc.ExpectedOutput, &tc.OrderIndex,
		); err != nil {
			return nil, fmt.Errorf("scan test case row: %w", err)
		}
		cases = append(cases, tc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate test case rows: %w", err)
	}

	return cases, nil
}

// GetDailyChallenge returns today's daily challenge problem.
// It automatically selects a new problem if one hasn't been chosen yet for today.
func (r *ProblemRepository) GetDailyChallenge(ctx context.Context) (*models.Problem, error) {
	// Try to insert an unused problem for today's challenge.
	insertUnusedQuery := `
		INSERT INTO daily_challenges (problem_id, challenge_date)
		SELECT id, CURRENT_DATE
		FROM problems
		WHERE is_published = true
		  AND id NOT IN (SELECT problem_id FROM daily_challenges)
		ORDER BY RANDOM()
		LIMIT 1
		ON CONFLICT (challenge_date) DO NOTHING
	`
	if _, err := r.pool.Exec(ctx, insertUnusedQuery); err != nil {
		return nil, fmt.Errorf("insert unused daily challenge: %w", err)
	}

	// If all problems have been used, the above query might not insert anything.
	// Try a fallback insert that allows reused problems.
	// If a challenge for today already exists, this will just DO NOTHING.
	insertFallbackQuery := `
		INSERT INTO daily_challenges (problem_id, challenge_date)
		SELECT id, CURRENT_DATE
		FROM problems
		WHERE is_published = true
		ORDER BY RANDOM()
		LIMIT 1
		ON CONFLICT (challenge_date) DO NOTHING
	`
	if _, err := r.pool.Exec(ctx, insertFallbackQuery); err != nil {
		return nil, fmt.Errorf("insert fallback daily challenge: %w", err)
	}

	query := `
		SELECT p.id, p.slug, p.title, p.difficulty, p.problem_md, p.time_limit_ms, p.memory_limit_mb,
		       p.acceptance_rate, p.is_published, p.execution_mode, p.function_name, p.initial_code, p.driver_code, p.solution_file_name, p.sample_test_cases, p.created_at, p.updated_at
		FROM problems p
		JOIN daily_challenges dc ON p.id = dc.problem_id
		WHERE dc.challenge_date = CURRENT_DATE
	`

	var p models.Problem
	err := r.pool.QueryRow(ctx, query).Scan(
		&p.ID, &p.Slug, &p.Title, &p.Difficulty, &p.ProblemMD,
		&p.TimeLimitMs, &p.MemoryLimitMb, &p.AcceptanceRate,
		&p.IsPublished, &p.ExecutionMode, &p.FunctionName, &p.InitialCode, &p.DriverCode, &p.SolutionFileName, &p.SampleTestCases, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No published problems available
		}
		return nil, fmt.Errorf("get daily challenge: %w", err)
	}

	return &p, nil
}
