package catalog

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"innogen-backend/shared/models"
)

// Execution projections are internal and must never be serialized as learner DTOs.
func (r *Repository) ExecutionProblem(ctx context.Context, id string) (*models.Problem, error) {
	p := &models.Problem{}
	err := r.pool.QueryRow(ctx, `SELECT p.id,p.slug,p.title,l.difficulty,p.statement_md,p.execution_mode,p.function_name,p.time_limit_ms,p.memory_limit_mb,p.status='published' FROM problems p JOIN levels l ON l.id=p.level_id WHERE p.id=$1`, id).Scan(&p.ID, &p.Slug, &p.Title, &p.Difficulty, &p.ProblemMD, &p.ExecutionMode, &p.FunctionName, &p.TimeLimitMs, &p.MemoryLimitMb, &p.IsPublished)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return p, err
}
func (r *Repository) ExecutionTemplate(ctx context.Context, p *models.Problem, languageID string) error {
	err := r.pool.QueryRow(ctx, `SELECT wrapper_code,file_name FROM code_templates WHERE problem_id=$1 AND language_id=$2`, p.ID, languageID).Scan(&p.DriverCode, &p.SolutionFileName)
	if errors.Is(err, pgx.ErrNoRows) && p.ExecutionMode == "stdin_stdout" {
		return nil
	}
	if err != nil {
		return err
	}
	if p.ExecutionMode == "function" && (p.DriverCode == nil || *p.DriverCode == "") {
		return errors.New("function template wrapper missing")
	}
	return nil
}
func (r *Repository) ExecutionLanguage(ctx context.Context, id string) (*models.Language, error) {
	l := &models.Language{}
	err := r.pool.QueryRow(ctx, `SELECT id,name,piston_alias,piston_version,file_extension,default_file_name,is_active,created_at,updated_at FROM languages WHERE id=$1 AND is_active`, id).Scan(&l.ID, &l.Name, &l.PistonAlias, &l.PistonVersion, &l.FileExtension, &l.DefaultFileName, &l.IsActive, &l.CreatedAt, &l.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return l, err
}
func (r *Repository) ExecutionTests(ctx context.Context, id string, hidden bool) ([]models.TestCase, error) {
	query := `SELECT id,problem_id,'sample',input_data,expected_output,display_order FROM problem_samples WHERE problem_id=$1 ORDER BY display_order,id`
	if hidden {
		query = `SELECT id,problem_id,'hidden',input_data,expected_output,display_order FROM problem_test_cases WHERE problem_id=$1 AND is_active ORDER BY display_order,id`
	}
	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []models.TestCase{}
	for rows.Next() {
		var tc models.TestCase
		if err = rows.Scan(&tc.ID, &tc.ProblemID, &tc.Visibility, &tc.InputData, &tc.ExpectedOutput, &tc.OrderIndex); err != nil {
			return nil, err
		}
		items = append(items, tc)
	}
	return items, rows.Err()
}
