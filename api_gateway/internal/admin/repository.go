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
//  SUBJECTS
// =========================================================================

// CreateSubject inserts a new subject and returns the full row.
func (r *AdminRepository) CreateSubject(ctx context.Context, req CreateSubjectRequest) (*models.Subject, error) {
	var sub models.Subject
	err := r.pool.QueryRow(ctx,
		`INSERT INTO subjects (title, slug, color, is_published, language_id)
		 VALUES ($1, $2, $3, COALESCE($4, FALSE), $5)
		 RETURNING id, title, slug, color, is_published, language_id, created_at, updated_at`,
		req.Title, req.Slug, req.Color, req.IsPublished, req.LanguageID,
	).Scan(
		&sub.ID, &sub.Title, &sub.Slug,
		&sub.Color, &sub.IsPublished, &sub.LanguageID,
		&sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("admin.CreateSubject: %w", err)
	}
	return &sub, nil
}

// UpdateSubject updates a subject by id. Only non-nil fields are updated.
// Returns nil, nil if the row does not exist.
func (r *AdminRepository) UpdateSubject(ctx context.Context, id int, req UpdateSubjectRequest) (*models.Subject, error) {
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Slug != nil {
		updates["slug"] = *req.Slug
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	if req.IsPublished != nil {
		updates["is_published"] = *req.IsPublished
	}
	if req.LanguageID != nil {
		updates["language_id"] = *req.LanguageID
	}

	if len(updates) == 0 {
		return r.FindSubjectByID(ctx, id)
	}

	setClause, values := buildSetClause(updates, 2)
	query := fmt.Sprintf(
		`UPDATE subjects SET %s WHERE id=$1
		 RETURNING id, title, slug, color, is_published, language_id, created_at, updated_at`,
		setClause,
	)
	args := append([]interface{}{id}, values...)

	var sub models.Subject
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&sub.ID, &sub.Title, &sub.Slug,
		&sub.Color, &sub.IsPublished, &sub.LanguageID,
		&sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("admin.UpdateSubject: %w", err)
	}
	return &sub, nil
}

// DeleteSubject deletes a subject by id. Returns true if a row was deleted.
func (r *AdminRepository) DeleteSubject(ctx context.Context, id int) (bool, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM subjects WHERE id=$1`, id)
	if err != nil {
		return false, fmt.Errorf("admin.DeleteSubject: %w", err)
	}
	return ct.RowsAffected() > 0, nil
}

// FindAllSubjects returns all subjects ordered by id.
func (r *AdminRepository) FindAllSubjects(ctx context.Context) ([]models.Subject, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, title, slug, color, is_published, language_id, created_at, updated_at
		 FROM subjects ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("admin.FindAllSubjects: %w", err)
	}
	defer rows.Close()

	items := make([]models.Subject, 0)
	for rows.Next() {
		var sub models.Subject
		err := rows.Scan(
			&sub.ID, &sub.Title, &sub.Slug,
			&sub.Color, &sub.IsPublished, &sub.LanguageID,
			&sub.CreatedAt, &sub.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("admin.FindAllSubjects: %w", err)
		}
		items = append(items, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("admin.FindAllSubjects: %w", err)
	}
	return items, nil
}

// FindSubjectByID returns a subject by id, or nil if not found.
func (r *AdminRepository) FindSubjectByID(ctx context.Context, id int) (*models.Subject, error) {
	var sub models.Subject
	err := r.pool.QueryRow(ctx,
		`SELECT id, title, slug, color, is_published, language_id, created_at, updated_at
		 FROM subjects WHERE id=$1`, id,
	).Scan(
		&sub.ID, &sub.Title, &sub.Slug,
		&sub.Color, &sub.IsPublished, &sub.LanguageID,
		&sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("admin.FindSubjectByID: %w", err)
	}
	return &sub, nil
}

// =========================================================================
//  SESSIONS
// =========================================================================

// CreateSession inserts a new session for the given subject and returns the full row.
func (r *AdminRepository) CreateSession(ctx context.Context, subjectID int, req CreateSessionRequest) (*models.SubjectSession, error) {
	var sess models.SubjectSession
	err := r.pool.QueryRow(ctx,
		`INSERT INTO subject_sessions (subject_id, title, description, order_index)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, subject_id, title, description, order_index, created_at, updated_at`,
		subjectID, req.Title, req.Description, req.OrderIndex,
	).Scan(
		&sess.ID, &sess.SubjectID, &sess.Title, &sess.Description,
		&sess.OrderIndex, &sess.CreatedAt, &sess.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("admin.CreateSession: %w", err)
	}
	return &sess, nil
}

// UpdateSession updates a session by id. Only non-nil fields are updated.
// Returns nil, nil if the row does not exist.
func (r *AdminRepository) UpdateSession(ctx context.Context, id int, req UpdateSessionRequest) (*models.SubjectSession, error) {
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.OrderIndex != nil {
		updates["order_index"] = *req.OrderIndex
	}

	if len(updates) == 0 {
		return r.FindSessionByID(ctx, id)
	}

	setClause, values := buildSetClause(updates, 2)
	query := fmt.Sprintf(
		`UPDATE subject_sessions SET %s WHERE id=$1
		 RETURNING id, subject_id, title, description, order_index, created_at, updated_at`,
		setClause,
	)
	args := append([]interface{}{id}, values...)

	var sess models.SubjectSession
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&sess.ID, &sess.SubjectID, &sess.Title, &sess.Description,
		&sess.OrderIndex, &sess.CreatedAt, &sess.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("admin.UpdateSession: %w", err)
	}
	return &sess, nil
}

// DeleteSession deletes a session by id. Returns true if a row was deleted.
func (r *AdminRepository) DeleteSession(ctx context.Context, id int) (bool, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM subject_sessions WHERE id=$1`, id)
	if err != nil {
		return false, fmt.Errorf("admin.DeleteSession: %w", err)
	}
	return ct.RowsAffected() > 0, nil
}

// FindSessionByID returns a session by id, or nil if not found.
func (r *AdminRepository) FindSessionByID(ctx context.Context, id int) (*models.SubjectSession, error) {
	var sess models.SubjectSession
	err := r.pool.QueryRow(ctx,
		`SELECT id, subject_id, title, description, order_index, created_at, updated_at
		 FROM subject_sessions WHERE id=$1`, id,
	).Scan(
		&sess.ID, &sess.SubjectID, &sess.Title, &sess.Description,
		&sess.OrderIndex, &sess.CreatedAt, &sess.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("admin.FindSessionByID: %w", err)
	}
	return &sess, nil
}

// =========================================================================
//  LESSONS
// =========================================================================

// CreateLesson inserts a new lesson for the given session and returns the full row.
func (r *AdminRepository) CreateLesson(ctx context.Context, sessionID int, req CreateLessonRequest) (*models.Lesson, error) {
	var lesson models.Lesson
	err := r.pool.QueryRow(ctx,
		`INSERT INTO lessons (subject_session_id, title, content_md, order_index)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, subject_session_id, title, content_md, order_index, created_at, updated_at`,
		sessionID, req.Title, req.ContentMD, req.OrderIndex,
	).Scan(
		&lesson.ID, &lesson.SubjectSessionID, &lesson.Title,
		&lesson.ContentMD, &lesson.OrderIndex, &lesson.CreatedAt, &lesson.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("admin.CreateLesson: %w", err)
	}
	return &lesson, nil
}

// UpdateLesson updates a lesson by id. Only non-nil fields are updated.
// Returns nil, nil if the row does not exist.
func (r *AdminRepository) UpdateLesson(ctx context.Context, id int, req UpdateLessonRequest) (*models.Lesson, error) {
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.ContentMD != nil {
		updates["content_md"] = *req.ContentMD
	}
	if req.OrderIndex != nil {
		updates["order_index"] = *req.OrderIndex
	}

	if len(updates) == 0 {
		return r.FindLessonByID(ctx, id)
	}

	setClause, values := buildSetClause(updates, 2)
	query := fmt.Sprintf(
		`UPDATE lessons SET %s WHERE id=$1
		 RETURNING id, subject_session_id, title, content_md, order_index, created_at, updated_at`,
		setClause,
	)
	args := append([]interface{}{id}, values...)

	var lesson models.Lesson
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&lesson.ID, &lesson.SubjectSessionID, &lesson.Title,
		&lesson.ContentMD, &lesson.OrderIndex, &lesson.CreatedAt, &lesson.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("admin.UpdateLesson: %w", err)
	}
	return &lesson, nil
}

// DeleteLesson deletes a lesson by id. Returns true if a row was deleted.
func (r *AdminRepository) DeleteLesson(ctx context.Context, id int) (bool, error) {
	ct, err := r.pool.Exec(ctx, `DELETE FROM lessons WHERE id=$1`, id)
	if err != nil {
		return false, fmt.Errorf("admin.DeleteLesson: %w", err)
	}
	return ct.RowsAffected() > 0, nil
}

// FindLessonByID returns a lesson by id, or nil if not found.
func (r *AdminRepository) FindLessonByID(ctx context.Context, id int) (*models.Lesson, error) {
	var lesson models.Lesson
	err := r.pool.QueryRow(ctx,
		`SELECT id, subject_session_id, title, content_md, order_index, created_at, updated_at
		 FROM lessons WHERE id=$1`, id,
	).Scan(
		&lesson.ID, &lesson.SubjectSessionID, &lesson.Title,
		&lesson.ContentMD, &lesson.OrderIndex, &lesson.CreatedAt, &lesson.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("admin.FindLessonByID: %w", err)
	}
	return &lesson, nil
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
//  LESSON-PROBLEMS
// =========================================================================

// CreateLessonProblem links a problem to a lesson with an order index.
func (r *AdminRepository) CreateLessonProblem(ctx context.Context, lessonID int, req CreateLessonProblemRequest) (*models.LessonProblem, error) {
	var lp models.LessonProblem
	err := r.pool.QueryRow(ctx,
		`INSERT INTO lesson_problems (lesson_id, problem_id, order_index)
		 VALUES ($1, $2, $3)
		 RETURNING lesson_id, problem_id, order_index`,
		lessonID, req.ProblemID, req.OrderIndex,
	).Scan(&lp.LessonID, &lp.ProblemID, &lp.OrderIndex)
	if err != nil {
		return nil, fmt.Errorf("admin.CreateLessonProblem: %w", err)
	}
	return &lp, nil
}

// DeleteLessonProblem removes a problem from a lesson. Returns true if a row was deleted.
func (r *AdminRepository) DeleteLessonProblem(ctx context.Context, lessonID, problemID int) (bool, error) {
	ct, err := r.pool.Exec(ctx,
		`DELETE FROM lesson_problems WHERE lesson_id=$1 AND problem_id=$2`,
		lessonID, problemID,
	)
	if err != nil {
		return false, fmt.Errorf("admin.DeleteLessonProblem: %w", err)
	}
	return ct.RowsAffected() > 0, nil
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
