package curriculum

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"innogen-backend/shared/models"
)

// CurriculumRepository provides database access for curriculum entities.
type CurriculumRepository struct {
	pool *pgxpool.Pool
}

// NewCurriculumRepository creates a new CurriculumRepository.
func NewCurriculumRepository(pool *pgxpool.Pool) *CurriculumRepository {
	return &CurriculumRepository{pool: pool}
}

// FindAllSubjects returns all published subjects ordered by id.
func (r *CurriculumRepository) FindAllSubjects(ctx context.Context) ([]models.Subject, error) {
	query := `SELECT id, title, slug, color, language_id FROM subjects WHERE is_published = true ORDER BY id`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find all subjects: %w", err)
	}
	defer rows.Close()

	var subjects []models.Subject
	for rows.Next() {
		var s models.Subject
		if err := rows.Scan(&s.ID, &s.Title, &s.Slug, &s.Color, &s.LanguageID); err != nil {
			return nil, fmt.Errorf("scan subject row: %w", err)
		}
		subjects = append(subjects, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subject rows: %w", err)
	}

	return subjects, nil
}

// FindSubjectBySlug returns a single published subject by slug.
// Returns nil, nil if not found.
func (r *CurriculumRepository) FindSubjectBySlug(ctx context.Context, slug string) (*models.Subject, error) {
	query := `SELECT id, title, slug, color, is_published, language_id, created_at, updated_at FROM subjects WHERE slug = $1 AND is_published = true`

	var s models.Subject
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&s.ID, &s.Title, &s.Slug, &s.Color, &s.IsPublished, &s.LanguageID, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find subject by slug %q: %w", slug, err)
	}

	return &s, nil
}

// FindSessionsBySubjectID returns sessions for a given subject ordered by order_index.
func (r *CurriculumRepository) FindSessionsBySubjectID(ctx context.Context, subjectID int) ([]models.SubjectSession, error) {
	query := `SELECT id, subject_id, title, description, order_index FROM subject_sessions WHERE subject_id = $1 ORDER BY order_index ASC`

	rows, err := r.pool.Query(ctx, query, subjectID)
	if err != nil {
		return nil, fmt.Errorf("find sessions by subject id %d: %w", subjectID, err)
	}
	defer rows.Close()

	var sessions []models.SubjectSession
	for rows.Next() {
		var s models.SubjectSession
		if err := rows.Scan(&s.ID, &s.SubjectID, &s.Title, &s.Description, &s.OrderIndex); err != nil {
			return nil, fmt.Errorf("scan session row: %w", err)
		}
		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate session rows: %w", err)
	}

	return sessions, nil
}

// FindLessonsBySessionID returns lessons for a given session ordered by order_index.
func (r *CurriculumRepository) FindLessonsBySessionID(ctx context.Context, sessionID int) ([]models.Lesson, error) {
	query := `SELECT id, subject_session_id, title, content_md, order_index FROM lessons WHERE subject_session_id = $1 ORDER BY order_index ASC`

	rows, err := r.pool.Query(ctx, query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("find lessons by session id %d: %w", sessionID, err)
	}
	defer rows.Close()

	var lessons []models.Lesson
	for rows.Next() {
		var l models.Lesson
		if err := rows.Scan(&l.ID, &l.SubjectSessionID, &l.Title, &l.ContentMD, &l.OrderIndex); err != nil {
			return nil, fmt.Errorf("scan lesson row: %w", err)
		}
		lessons = append(lessons, l)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate lesson rows: %w", err)
	}

	return lessons, nil
}

// FindLessonByID returns a single lesson by id.
// Returns nil, nil if not found.
func (r *CurriculumRepository) FindLessonByID(ctx context.Context, id int) (*models.Lesson, error) {
	query := `SELECT id, subject_session_id, title, content_md, order_index, created_at, updated_at FROM lessons WHERE id = $1`

	var l models.Lesson
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&l.ID, &l.SubjectSessionID, &l.Title, &l.ContentMD, &l.OrderIndex, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find lesson by id %d: %w", id, err)
	}

	return &l, nil
}

// FindProblemsByLessonID returns published problems associated with a lesson.
func (r *CurriculumRepository) FindProblemsByLessonID(ctx context.Context, lessonID int) ([]ProblemListItem, error) {
	query := `
		SELECT p.id, p.slug, p.title, p.difficulty, p.acceptance_rate, lp.order_index
		FROM problems p
		JOIN lesson_problems lp ON p.id = lp.problem_id
		WHERE lp.lesson_id = $1 AND p.is_published = true
		ORDER BY lp.order_index ASC
	`

	rows, err := r.pool.Query(ctx, query, lessonID)
	if err != nil {
		return nil, fmt.Errorf("find problems by lesson id %d: %w", lessonID, err)
	}
	defer rows.Close()

	var items []ProblemListItem
	for rows.Next() {
		var item ProblemListItem
		if err := rows.Scan(&item.ID, &item.Slug, &item.Title, &item.Difficulty, &item.AcceptanceRate, &item.OrderIndex); err != nil {
			return nil, fmt.Errorf("scan problem list item row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate problem list item rows: %w", err)
	}

	return items, nil
}
