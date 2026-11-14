package repository

import (
	"context"
	"time"
)

// Heartbeat refreshes only the current attempt's lease.
func (r *SubmissionRepository) Heartbeat(ctx context.Context, id string, attempt int) error {
	_, err := r.pool.Exec(ctx, `UPDATE submissions SET claimed_at=now() WHERE id=$1 AND status='Running' AND attempt_count=$2`, id, attempt)
	return err
}

// Recover bounds each scan and atomically fences stale workers before enqueue.
func (r *SubmissionRepository) Recover(ctx context.Context, pendingAge, runningAge time.Duration, maxAttempts int) ([]string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, `WITH stale AS (
 SELECT id FROM submissions WHERE (status='Pending' AND updated_at<now()-$1*interval '1 second') OR (status='Running' AND claimed_at<now()-$2*interval '1 second')
 ORDER BY updated_at,id LIMIT 100 FOR UPDATE SKIP LOCKED)
 UPDATE submissions s SET status=CASE WHEN attempt_count >= $3 THEN 'Internal Error'::submission_status ELSE 'Pending'::submission_status END,
 claimed_at=NULL, judged_at=CASE WHEN attempt_count >= $3 THEN now() ELSE NULL END,
 error_message=CASE WHEN attempt_count >= $3 THEN 'Judging retries exhausted' ELSE NULL END
 FROM stale WHERE s.id=stale.id RETURNING s.id,s.status`, pendingAge.Seconds(), runningAge.Seconds(), maxAttempts)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for rows.Next() {
		var id, status string
		if err = rows.Scan(&id, &status); err != nil {
			rows.Close()
			return nil, err
		}
		if status == "Pending" {
			ids = append(ids, id)
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return ids, nil
}
