package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"innogen-backend/shared/models"
)

// AdminRepository is the data access layer for admin operations.
type AdminRepository struct {
	pool *pgxpool.Pool
}

// NewAdminRepository creates a new AdminRepository.
func NewAdminRepository(pool *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{pool: pool}
}

// buildSetClause constructs a dynamic SET clause for UPDATE statements.
// updates is a map of column name to value. startIdx is the first parameter
// index to use (e.g. 2 when $1 is reserved for the id).
// Returns the SET clause string (e.g. "name=$2, slug=$3") and the values slice
// in the same order as the clause parameters.
func buildSetClause(updates map[string]interface{}, startIdx int) (string, []interface{}) {
	if len(updates) == 0 {
		return "", nil
	}

	keys := make([]string, 0, len(updates))
	for k := range updates {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	clauses := make([]string, 0, len(keys))
	values := make([]interface{}, 0, len(keys))
	for i, k := range keys {
		clauses = append(clauses, fmt.Sprintf("%s=$%d", k, startIdx+i))
		values = append(values, updates[k])
	}
	return strings.Join(clauses, ", "), values
}

// =========================================================================
//  LANGUAGES
// =========================================================================

// CreateLanguage inserts a new language and returns the full row.
func (r *AdminRepository) CreateLanguage(ctx context.Context, req CreateLanguageRequest) (*models.Language, error) {
	var lang models.Language
	err := r.pool.QueryRow(ctx,
		`INSERT INTO languages (name, piston_alias, piston_version, file_extension, default_file_name, is_active)
		 VALUES ($1, $2, $3, $4, $5, COALESCE($6, TRUE))
		 RETURNING id, name, piston_alias, piston_version, file_extension, default_file_name, is_active, created_at, updated_at`,
		req.Name, req.PistonAlias, req.PistonVersion, req.FileExtension, req.DefaultFileName, req.IsActive,
	).Scan(
		&lang.ID, &lang.Name, &lang.PistonAlias, &lang.PistonVersion,
		&lang.FileExtension, &lang.DefaultFileName, &lang.IsActive,
		&lang.CreatedAt, &lang.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("admin.CreateLanguage: %w", err)
	}
	return &lang, nil
}

// UpdateLanguage updates a language by id. Only non-nil fields are updated.
// Returns nil, nil if the row does not exist.
func (r *AdminRepository) UpdateLanguage(ctx context.Context, id int, req UpdateLanguageRequest) (*models.Language, error) {
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.PistonAlias != nil {
		updates["piston_alias"] = *req.PistonAlias
	}
	if req.PistonVersion != nil {
		updates["piston_version"] = *req.PistonVersion
	}
	if req.FileExtension != nil {
		updates["file_extension"] = *req.FileExtension
	}
	if req.DefaultFileName != nil {
		updates["default_file_name"] = *req.DefaultFileName
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		return r.FindLanguageByID(ctx, id)
	}

	setClause, values := buildSetClause(updates, 2)
	query := fmt.Sprintf(
		`UPDATE languages SET %s WHERE id=$1
		 RETURNING id, name, piston_alias, piston_version, file_extension, default_file_name, is_active, created_at, updated_at`,
		setClause,
	)
	args := append([]interface{}{id}, values...)

	var lang models.Language
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&lang.ID, &lang.Name, &lang.PistonAlias, &lang.PistonVersion,
		&lang.FileExtension, &lang.DefaultFileName, &lang.IsActive,
		&lang.CreatedAt, &lang.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("admin.UpdateLanguage: %w", err)
	}
	return &lang, nil
}

// FindAllLanguages returns all languages ordered by id.
func (r *AdminRepository) FindAllLanguages(ctx context.Context) ([]models.Language, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, piston_alias, piston_version, file_extension, default_file_name, is_active, created_at, updated_at
		 FROM languages ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("admin.FindAllLanguages: %w", err)
	}
	defer rows.Close()

	items := make([]models.Language, 0)
	for rows.Next() {
		var lang models.Language
		err := rows.Scan(
			&lang.ID, &lang.Name, &lang.PistonAlias, &lang.PistonVersion,
			&lang.FileExtension, &lang.DefaultFileName, &lang.IsActive,
			&lang.CreatedAt, &lang.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("admin.FindAllLanguages: %w", err)
		}
		items = append(items, lang)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("admin.FindAllLanguages: %w", err)
	}
	return items, nil
}

// FindLanguageByID returns a language by id, or nil if not found.
func (r *AdminRepository) FindLanguageByID(ctx context.Context, id int) (*models.Language, error) {
	var lang models.Language
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, piston_alias, piston_version, file_extension, default_file_name, is_active, created_at, updated_at
		 FROM languages WHERE id=$1`, id,
	).Scan(
		&lang.ID, &lang.Name, &lang.PistonAlias, &lang.PistonVersion,
		&lang.FileExtension, &lang.DefaultFileName, &lang.IsActive,
		&lang.CreatedAt, &lang.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("admin.FindLanguageByID: %w", err)
	}
	return &lang, nil
}

// =========================================================================
//  PROBLEMS
// =========================================================================

// CreateProblem inserts a new problem and returns the full row.
func (r *AdminRepository) CreateProblem(ctx context.Context, req CreateProblemRequest) (*models.Problem, error) {
	var prob models.Problem
	var sampleTCBytes []byte

	err := r.pool.QueryRow(ctx,
		`INSERT INTO problems (slug, title, difficulty, problem_md, time_limit_ms, memory_limit_mb, is_published, execution_mode, function_name, initial_code, driver_code, solution_file_name, sample_test_cases)
		 VALUES ($1, $2, $3, $4, COALESCE($5, 1000), COALESCE($6, 128), COALESCE($7, FALSE), COALESCE($8, 'function'), $9, $10, $11, $12, $13)
		 RETURNING id, slug, title, difficulty, problem_md, time_limit_ms, memory_limit_mb, acceptance_rate, is_published, execution_mode, function_name, initial_code, driver_code, solution_file_name, sample_test_cases, created_at, updated_at`,
		req.Slug, req.Title, req.Difficulty, req.ProblemMD,
		req.TimeLimitMs, req.MemoryLimitMb, req.IsPublished,
		req.ExecutionMode, req.FunctionName, req.InitialCode, req.DriverCode, req.SolutionFileName, req.SampleTestCases,
	).Scan(
		&prob.ID, &prob.Slug, &prob.Title, &prob.Difficulty, &prob.ProblemMD,
		&prob.TimeLimitMs, &prob.MemoryLimitMb, &prob.AcceptanceRate, &prob.IsPublished,
		&prob.ExecutionMode, &prob.FunctionName, &prob.InitialCode, &prob.DriverCode, &prob.SolutionFileName,
		&sampleTCBytes, &prob.CreatedAt, &prob.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("admin.CreateProblem: %w", err)
	}
	prob.SampleTestCases = json.RawMessage(sampleTCBytes)
	return &prob, nil
}

// UpdateProblem updates a problem by id. Only non-nil fields are updated.
// Returns nil, nil if the row does not exist.
func (r *AdminRepository) UpdateProblem(ctx context.Context, id int, req UpdateProblemRequest) (*models.Problem, error) {
	updates := make(map[string]interface{})
	if req.Slug != nil {
		updates["slug"] = *req.Slug
	}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Difficulty != nil {
		updates["difficulty"] = *req.Difficulty
	}
	if req.ProblemMD != nil {
		updates["problem_md"] = *req.ProblemMD
	}
	if req.TimeLimitMs != nil {
		updates["time_limit_ms"] = *req.TimeLimitMs
	}
	if req.MemoryLimitMb != nil {
		updates["memory_limit_mb"] = *req.MemoryLimitMb
	}
	if req.IsPublished != nil {
		updates["is_published"] = *req.IsPublished
	}
	if req.ExecutionMode != nil {
		updates["execution_mode"] = *req.ExecutionMode
	}
	if req.FunctionName != nil {
		updates["function_name"] = *req.FunctionName
	}
	if req.InitialCode != nil {
		updates["initial_code"] = *req.InitialCode
	}
	if req.DriverCode != nil {
		updates["driver_code"] = *req.DriverCode
	}
	if req.SolutionFileName != nil {
		updates["solution_file_name"] = *req.SolutionFileName
	}
	if req.SampleTestCases != nil {
		updates["sample_test_cases"] = req.SampleTestCases
	}

	if len(updates) == 0 {
		return r.FindProblemByID(ctx, id)
	}

	setClause, values := buildSetClause(updates, 2)
	query := fmt.Sprintf(
		`UPDATE problems SET %s WHERE id=$1
		 RETURNING id, slug, title, difficulty, problem_md, time_limit_ms, memory_limit_mb, acceptance_rate, is_published, execution_mode, function_name, initial_code, driver_code, solution_file_name, sample_test_cases, created_at, updated_at`,
		setClause,
	)
	args := append([]interface{}{id}, values...)

	var prob models.Problem
	var sampleTCBytes []byte
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&prob.ID, &prob.Slug, &prob.Title, &prob.Difficulty, &prob.ProblemMD,
		&prob.TimeLimitMs, &prob.MemoryLimitMb, &prob.AcceptanceRate, &prob.IsPublished,
		&prob.ExecutionMode, &prob.FunctionName, &prob.InitialCode, &prob.DriverCode, &prob.SolutionFileName,
		&sampleTCBytes, &prob.CreatedAt, &prob.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("admin.UpdateProblem: %w", err)
	}
	prob.SampleTestCases = json.RawMessage(sampleTCBytes)
	return &prob, nil
}

// DeleteProblem deletes a problem by id. Returns true if a row was deleted.
func (r *AdminRepository) DeleteProblem(ctx context.Context, id int) (bool, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM problems WHERE id=$1`, id)
	if err != nil {
		return false, fmt.Errorf("admin.DeleteProblem: %w", err)
	}
	return ct.RowsAffected() > 0, nil
}

// FindAllProblems returns a paginated list of problems with total count.
func (r *AdminRepository) FindAllProblems(ctx context.Context, page, limit int) ([]models.Problem, int, error) {
	// Count total
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM problems`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("admin.FindAllProblems: %w", err)
	}

	// Fetch page
	offset := (page - 1) * limit
	rows, err := r.pool.Query(ctx,
		`SELECT id, slug, title, difficulty, problem_md, time_limit_ms, memory_limit_mb, acceptance_rate, is_published, execution_mode, function_name, initial_code, driver_code, solution_file_name, sample_test_cases, created_at, updated_at
		 FROM problems ORDER BY id LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("admin.FindAllProblems: %w", err)
	}
	defer rows.Close()

	items := make([]models.Problem, 0)
	for rows.Next() {
		var prob models.Problem
		var sampleTCBytes []byte
		err := rows.Scan(
			&prob.ID, &prob.Slug, &prob.Title, &prob.Difficulty, &prob.ProblemMD,
			&prob.TimeLimitMs, &prob.MemoryLimitMb, &prob.AcceptanceRate, &prob.IsPublished,
			&prob.ExecutionMode, &prob.FunctionName, &prob.InitialCode, &prob.DriverCode, &prob.SolutionFileName,
			&sampleTCBytes, &prob.CreatedAt, &prob.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("admin.FindAllProblems: %w", err)
		}
		prob.SampleTestCases = json.RawMessage(sampleTCBytes)
		items = append(items, prob)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("admin.FindAllProblems: %w", err)
	}
	return items, total, nil
}

// FindProblemByID returns a problem by id, or nil if not found.
func (r *AdminRepository) FindProblemByID(ctx context.Context, id int) (*models.Problem, error) {
	var prob models.Problem
	var sampleTCBytes []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id, slug, title, difficulty, problem_md, time_limit_ms, memory_limit_mb, acceptance_rate, is_published, execution_mode, function_name, initial_code, driver_code, solution_file_name, sample_test_cases, created_at, updated_at
		 FROM problems WHERE id=$1`, id,
	).Scan(
		&prob.ID, &prob.Slug, &prob.Title, &prob.Difficulty, &prob.ProblemMD,
		&prob.TimeLimitMs, &prob.MemoryLimitMb, &prob.AcceptanceRate, &prob.IsPublished,
		&prob.ExecutionMode, &prob.FunctionName, &prob.InitialCode, &prob.DriverCode, &prob.SolutionFileName,
		&sampleTCBytes, &prob.CreatedAt, &prob.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("admin.FindProblemByID: %w", err)
	}
	prob.SampleTestCases = json.RawMessage(sampleTCBytes)
	return &prob, nil
}

// =========================================================================
//  TEST CASES
// =========================================================================

// CreateTestCase inserts a new test case for the given problem and returns the full row.
func (r *AdminRepository) CreateTestCase(ctx context.Context, problemID int, req CreateTestCaseRequest) (*models.TestCase, error) {
	var tc models.TestCase
	err := r.pool.QueryRow(ctx,
		`INSERT INTO test_cases (problem_id, visibility, input_data, expected_output, execute_code, order_index)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, problem_id, visibility, input_data, expected_output, execute_code, order_index, created_at`,
		problemID, req.Visibility, req.InputData, req.ExpectedOutput, req.ExecuteCode, req.OrderIndex,
	).Scan(
		&tc.ID, &tc.ProblemID, &tc.Visibility,
		&tc.InputData, &tc.ExpectedOutput, &tc.ExecuteCode,
		&tc.OrderIndex, &tc.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("admin.CreateTestCase: %w", err)
	}
	return &tc, nil
}

// UpdateTestCase updates a test case by id. Only non-nil fields are updated.
// Returns nil, nil if the row does not exist.
func (r *AdminRepository) UpdateTestCase(ctx context.Context, id int, req UpdateTestCaseRequest) (*models.TestCase, error) {
	updates := make(map[string]interface{})
	if req.Visibility != nil {
		updates["visibility"] = *req.Visibility
	}
	if req.InputData != nil {
		updates["input_data"] = *req.InputData
	}
	if req.ExpectedOutput != nil {
		updates["expected_output"] = *req.ExpectedOutput
	}
	if req.ExecuteCode != nil {
		updates["execute_code"] = *req.ExecuteCode
	}
	if req.OrderIndex != nil {
		updates["order_index"] = *req.OrderIndex
	}

	if len(updates) == 0 {
		return r.FindTestCaseByID(ctx, id)
	}

	setClause, values := buildSetClause(updates, 2)
	query := fmt.Sprintf(
		`UPDATE test_cases SET %s WHERE id=$1
		 RETURNING id, problem_id, visibility, input_data, expected_output, execute_code, order_index, created_at`,
		setClause,
	)
	args := append([]interface{}{id}, values...)

	var tc models.TestCase
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&tc.ID, &tc.ProblemID, &tc.Visibility,
		&tc.InputData, &tc.ExpectedOutput, &tc.ExecuteCode,
		&tc.OrderIndex, &tc.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("admin.UpdateTestCase: %w", err)
	}
	return &tc, nil
}

// DeleteTestCase deletes a test case by id. Returns true if a row was deleted.
func (r *AdminRepository) DeleteTestCase(ctx context.Context, id int) (bool, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM test_cases WHERE id=$1`, id)
	if err != nil {
		return false, fmt.Errorf("admin.DeleteTestCase: %w", err)
	}
	return ct.RowsAffected() > 0, nil
}

// FindTestCasesByProblemID returns all test cases for a problem ordered by id.
func (r *AdminRepository) FindTestCasesByProblemID(ctx context.Context, problemID int) ([]models.TestCase, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, problem_id, visibility, input_data, expected_output, execute_code, order_index, created_at
		 FROM test_cases WHERE problem_id=$1 ORDER BY id`,
		problemID,
	)
	if err != nil {
		return nil, fmt.Errorf("admin.FindTestCasesByProblemID: %w", err)
	}
	defer rows.Close()

	items := make([]models.TestCase, 0)
	for rows.Next() {
		var tc models.TestCase
		err := rows.Scan(
			&tc.ID, &tc.ProblemID, &tc.Visibility,
			&tc.InputData, &tc.ExpectedOutput, &tc.ExecuteCode,
			&tc.OrderIndex, &tc.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("admin.FindTestCasesByProblemID: %w", err)
		}
		items = append(items, tc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("admin.FindTestCasesByProblemID: %w", err)
	}
	return items, nil
}

// FindTestCaseByID returns a test case by id, or nil if not found.
func (r *AdminRepository) FindTestCaseByID(ctx context.Context, id int) (*models.TestCase, error) {
	var tc models.TestCase
	err := r.pool.QueryRow(ctx,
		`SELECT id, problem_id, visibility, input_data, expected_output, execute_code, order_index, created_at
		 FROM test_cases WHERE id=$1`, id,
	).Scan(
		&tc.ID, &tc.ProblemID, &tc.Visibility,
		&tc.InputData, &tc.ExpectedOutput, &tc.ExecuteCode,
		&tc.OrderIndex, &tc.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("admin.FindTestCaseByID: %w", err)
	}
	return &tc, nil
}

// =========================================================================
//  TAGS
// =========================================================================

// CreateTag inserts a new tag and returns the full row.
func (r *AdminRepository) CreateTag(ctx context.Context, req CreateTagRequest) (*models.Tag, error) {
	var tag models.Tag
	err := r.pool.QueryRow(ctx,
		`INSERT INTO tags (name) VALUES ($1)
		 RETURNING id, name, created_at`,
		req.Name,
	).Scan(&tag.ID, &tag.Name, &tag.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("admin.CreateTag: %w", err)
	}
	return &tag, nil
}

// FindAllTags returns all tags ordered by id.
func (r *AdminRepository) FindAllTags(ctx context.Context) ([]models.Tag, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, created_at FROM tags ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("admin.FindAllTags: %w", err)
	}
	defer rows.Close()

	items := make([]models.Tag, 0)
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.ID, &tag.Name, &tag.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("admin.FindAllTags: %w", err)
		}
		items = append(items, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("admin.FindAllTags: %w", err)
	}
	return items, nil
}

// FindTagByID returns a tag by id, or nil if not found.
func (r *AdminRepository) FindTagByID(ctx context.Context, id int) (*models.Tag, error) {
	var tag models.Tag
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, created_at FROM tags WHERE id=$1`, id,
	).Scan(&tag.ID, &tag.Name, &tag.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("admin.FindTagByID: %w", err)
	}
	return &tag, nil
}

// =========================================================================
//  PROBLEM-TAGS
// =========================================================================

// CreateProblemTag links a tag to a problem. Returns an error on duplicate or FK violation.
func (r *AdminRepository) CreateProblemTag(ctx context.Context, problemID int, req CreateProblemTagRequest) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO problem_tags (problem_id, tag_id) VALUES ($1, $2)`,
		problemID, req.TagID,
	)
	if err != nil {
		return fmt.Errorf("admin.CreateProblemTag: %w", err)
	}
	return nil
}

// DeleteProblemTag removes a tag from a problem. Returns true if a row was deleted.
func (r *AdminRepository) DeleteProblemTag(ctx context.Context, problemID, tagID int) (bool, error) {
	ct, err := r.pool.Exec(ctx,
		`DELETE FROM problem_tags WHERE problem_id=$1 AND tag_id=$2`,
		problemID, tagID,
	)
	if err != nil {
		return false, fmt.Errorf("admin.DeleteProblemTag: %w", err)
	}
	return ct.RowsAffected() > 0, nil
}
