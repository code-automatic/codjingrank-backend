package review

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"innogen-backend/shared/middleware"
	"innogen-backend/shared/response"
)

func Register(mux *http.ServeMux, pool *pgxpool.Pool, jwtSecret, internalSecret string) {
	mux.Handle("GET /api/v1/submissions/{id}/ai-review", middleware.Auth(jwtSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, uid := r.PathValue("id"), ""
		uid, _ = middleware.GetUserID(r)
		if !middleware.ValidUUID(id) {
			response.ErrorValidation(w, "Invalid submission ID", "id")
			return
		}
		var active bool
		var role string
		err := pool.QueryRow(r.Context(), `SELECT is_active,role FROM users WHERE id=$1`, uid).Scan(&active, &role)
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && !active) {
			response.ErrorSimple(w, 403, "Account is inactive")
			return
		}
		if err != nil {
			response.ErrorSimple(w, 500, "Internal server error")
			return
		}
		var owner, status string
		err = pool.QueryRow(r.Context(), `SELECT user_id,status FROM submissions WHERE id=$1`, id).Scan(&owner, &status)
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && owner != uid && role != "admin") {
			response.ErrorSimple(w, 404, "Submission not found")
			return
		}
		if err != nil {
			response.ErrorSimple(w, 500, "Internal server error")
			return
		}
		if status != "Accepted" {
			response.Error(w, 409, "AI review requires Accepted", "AI_REVIEW_NOT_ELIGIBLE")
			return
		}
		var data []byte
		err = pool.QueryRow(r.Context(), `SELECT jsonb_build_object('id',r.id,'submissionId',r.submission_id,'status',r.status,'score',r.score,'summary',r.summary,'dimensions',r.dimensions,'errorMessage',CASE WHEN r.status='failed' THEN 'Review is temporarily unavailable' ELSE NULL END,'createdAt',r.created_at,'startedAt',r.started_at,'completedAt',r.completed_at,'findings',COALESCE((SELECT jsonb_agg(jsonb_build_object('id',f.id,'severity',f.severity,'lineStart',f.line_start,'lineEnd',f.line_end,'title',f.title,'explanation',f.explanation,'suggestion',f.suggestion,'displayOrder',f.display_order) ORDER BY f.display_order) FROM ai_review_findings f WHERE f.review_id=r.id),'[]'::jsonb)) FROM ai_reviews r WHERE r.submission_id=$1`, id).Scan(&data)
		if errors.Is(err, pgx.ErrNoRows) {
			response.Success(w, 200, map[string]any{"review": nil}, "AI review retrieved")
			return
		}
		if err != nil {
			response.ErrorSimple(w, 500, "Internal server error")
			return
		}
		response.Success(w, 200, map[string]any{"review": json.RawMessage(data)}, "AI review retrieved")
	})))
	mux.HandleFunc("POST /internal/reviews/{id}", func(w http.ResponseWriter, r *http.Request) {
		if internalSecret == "" || subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+internalSecret)) != 1 {
			response.ErrorSimple(w, 401, "Unauthorized")
			return
		}
		id := r.PathValue("id")
		if !middleware.ValidUUID(id) {
			response.ErrorValidation(w, "Invalid submission ID", "id")
			return
		}
		since, err := time.Parse(time.RFC3339, os.Getenv("AI_ENABLED_SINCE"))
		if err != nil || os.Getenv("AI_PROVIDER_URL") == "" || os.Getenv("AI_PROVIDER_MODEL") == "" {
			response.ErrorSimple(w, http.StatusServiceUnavailable, "AI review is disabled")
			return
		}
		var status string
		var judgedAt *time.Time
		err = pool.QueryRow(r.Context(), `SELECT status,judged_at FROM submissions WHERE id=$1`, id).Scan(&status, &judgedAt)
		if errors.Is(err, pgx.ErrNoRows) {
			response.ErrorSimple(w, 404, "Submission not found")
			return
		}
		if err != nil {
			response.ErrorSimple(w, 500, "Internal server error")
			return
		}
		if status != "Accepted" {
			response.Error(w, 409, "AI review requires Accepted", "AI_REVIEW_NOT_ELIGIBLE")
			return
		}
		if judgedAt == nil || judgedAt.Before(since) {
			response.ErrorSimple(w, 404, "Submission not eligible")
			return
		}
		_, err = pool.Exec(r.Context(), `INSERT INTO ai_reviews(submission_id) VALUES($1) ON CONFLICT(submission_id) DO NOTHING`, id)
		if err != nil {
			response.ErrorSimple(w, 500, "Internal server error")
			return
		}
		response.Success(w, 200, nil, "Review queued")
	})
}
