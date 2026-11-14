package repository

import (
	"context"
	"github.com/jackc/pgx/v5"
	"innogen-backend/shared/models"
	"innogen-backend/submission_service/internal/dto"
)

// History returns a consistent count/page without ever loading source code.
func (r *SubmissionRepository) History(ctx context.Context, uid string, q dto.HistoryQuery) (*dto.ListSubmissionsResponse, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	data := &dto.ListSubmissionsResponse{Items: []dto.SubmissionListItem{}, Page: q.Page, Size: q.Size}
	const where = ` FROM submissions s WHERE s.user_id=$1 AND (NULLIF($2,'')::uuid IS NULL OR s.problem_id=NULLIF($2,'')::uuid) AND (NULLIF($3,'')::uuid IS NULL OR s.language_id=NULLIF($3,'')::uuid) AND ($4='' OR s.status::text=$4)`
	if err = tx.QueryRow(ctx, `SELECT count(*)`+where, uid, q.ProblemID, q.LanguageID, q.Status).Scan(&data.Total); err != nil {
		return nil, err
	}
	rows, err := tx.Query(ctx, `SELECT s.id,s.user_id,s.problem_id,s.language_id,s.status,s.runtime_ms,s.memory_kb,s.error_message,s.pass_count,s.total_testcases,s.created_at,s.judged_at,
 (SELECT slug FROM problems WHERE id=s.problem_id),(SELECT title FROM problems WHERE id=s.problem_id),(SELECT slug FROM languages WHERE id=s.language_id)`+where+` ORDER BY s.created_at DESC,s.id DESC LIMIT $5 OFFSET $6`, uid, q.ProblemID, q.LanguageID, q.Status, q.Size, (q.Page-1)*q.Size)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var s models.Submission
		if err = rows.Scan(&s.ID, &s.UserID, &s.ProblemID, &s.LanguageID, &s.Status, &s.RuntimeMs, &s.MemoryKb, &s.ErrorMessage, &s.PassCount, &s.TotalTestcases, &s.CreatedAt, &s.JudgedAt, &s.ProblemSlug, &s.ProblemTitle, &s.LanguageSlug); err != nil {
			rows.Close()
			return nil, err
		}
		data.Items = append(data.Items, dto.ToSubmissionListItem(&s))
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return data, nil
}
