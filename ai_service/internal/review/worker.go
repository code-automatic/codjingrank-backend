package review

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func reconcile(ctx context.Context, pool *pgxpool.Pool, since time.Time) error {
	_, err := pool.Exec(ctx, `INSERT INTO ai_reviews(submission_id) SELECT id FROM submissions s WHERE s.status='Accepted' AND s.judged_at >= $1 AND NOT EXISTS (SELECT 1 FROM ai_reviews r WHERE r.submission_id=s.id) ON CONFLICT(submission_id) DO NOTHING`, since)
	return err
}

func Run(ctx context.Context, pool *pgxpool.Pool, since time.Time, url, token, model string, log *slog.Logger) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		if err := reconcile(ctx, pool, since); err != nil {
			log.Error("review reconciliation failed")
		}
		if err := process(ctx, pool, url, token, model); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			log.Error("review processing failed")
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func process(ctx context.Context, pool *pgxpool.Pool, url, token, model string) error {
	_, err := pool.Exec(ctx, `UPDATE ai_reviews SET status='failed',completed_at=now(),error_message='Review is temporarily unavailable' WHERE status='running' AND started_at < now()-interval '2 minutes' AND attempt_count>=3`)
	if err != nil {
		return err
	}
	var id string
	var attempt int
	var input providerInput
	err = pool.QueryRow(ctx, `UPDATE ai_reviews SET status='running',started_at=now(),attempt_count=attempt_count+1 WHERE id=(SELECT id FROM ai_reviews WHERE attempt_count < 3 AND (status='queued' OR (status='running' AND started_at < now()-interval '2 minutes')) ORDER BY created_at LIMIT 1 FOR UPDATE SKIP LOCKED) RETURNING id,attempt_count`).Scan(&id, &attempt)
	if err != nil {
		return err
	}
	err = pool.QueryRow(ctx, `SELECT s.source_code,l.name,p.statement_md FROM ai_reviews r JOIN submissions s ON s.id=r.submission_id JOIN languages l ON l.id=s.language_id JOIN problems p ON p.id=s.problem_id WHERE r.id=$1`, id).Scan(&input.Source, &input.Language, &input.Problem)
	var result providerResult
	if err == nil {
		result, err = callProvider(ctx, url, token, model, input)
	}
	if err != nil {
		var permanent permanentProviderError
		terminal := errors.As(err, &permanent) || attempt >= 3
		_, updateErr := pool.Exec(ctx, `UPDATE ai_reviews SET status=CASE WHEN $3 THEN 'failed'::ai_review_status ELSE 'queued'::ai_review_status END,completed_at=CASE WHEN $3 THEN now() ELSE NULL END,error_message=CASE WHEN $3 THEN 'Review is temporarily unavailable' ELSE NULL END WHERE id=$1 AND status='running' AND attempt_count=$2`, id, attempt, terminal)
		return updateErr
	}
	dimensions, err := json.Marshal(result.Dimensions)
	if err != nil {
		return err
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `UPDATE ai_reviews SET status='completed',score=$3,summary=$4,dimensions=$5,model=$6,prompt_version='v1',completed_at=now() WHERE id=$1 AND status='running' AND attempt_count=$2`, id, attempt, *result.Score, result.Summary, dimensions, model)
	if err != nil || tag.RowsAffected() == 0 {
		return err
	}
	for i, f := range result.Findings {
		_, err = tx.Exec(ctx, `INSERT INTO ai_review_findings(review_id,severity,line_start,line_end,title,explanation,suggestion,display_order) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, id, f.Severity, f.LineStart, f.LineEnd, f.Title, f.Explanation, f.Suggestion, i+1)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
