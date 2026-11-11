package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"innogen-backend/shared/catalog"
	"innogen-backend/shared/models"
)

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool} }
func (r *Repository) GetProblemByID(ctx context.Context, id string) (*models.Problem, error) {
	return catalog.New(r.pool).ExecutionProblem(ctx, id)
}
func (r *Repository) GetLanguageByID(ctx context.Context, id string) (*models.Language, error) {
	return catalog.New(r.pool).ExecutionLanguage(ctx, id)
}
func (r *Repository) GetSampleTestCases(ctx context.Context, id string) ([]models.TestCase, error) {
	return catalog.New(r.pool).ExecutionTests(ctx, id, false)
}
func (r *Repository) CheckExecution(ctx context.Context, uid, pid, lid string) error {
	return catalog.New(r.pool).CheckExecution(ctx, uid, pid, lid)
}
func (r *Repository) LoadTemplate(ctx context.Context, problem *models.Problem, lid string) error {
	return catalog.New(r.pool).ExecutionTemplate(ctx, problem, lid)
}
