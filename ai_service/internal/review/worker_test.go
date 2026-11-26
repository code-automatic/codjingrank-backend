package review

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"innogen-backend/shared/testutil"
)

func TestReconcileQueuesOnlyEligibleSubmissions(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.Background()
	var id string
	err := f.Pool.QueryRow(ctx, `INSERT INTO submissions(user_id,problem_id,language_id,source_code,source_hash,status,judged_at) VALUES($1,$2,$3,'source',repeat('a',64),'Accepted',now()) RETURNING id`, f.UserID, f.ProblemID, f.LanguageID).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := f.Pool.Exec(context.Background(), `DELETE FROM ai_reviews WHERE submission_id=$1`, id); err != nil {
			t.Error(err)
		}
	})
	other := testutil.Catalog(t)
	var wrongID string
	if err := other.Pool.QueryRow(ctx, `INSERT INTO submissions(user_id,problem_id,language_id,source_code,source_hash,status,judged_at)
		VALUES($1,$2,$3,'source',repeat('b',64),'Wrong Answer',now()) RETURNING id`, other.UserID, other.ProblemID, other.LanguageID).Scan(&wrongID); err != nil {
		t.Fatal(err)
	}
	assertCount := func(submission string, want int) {
		t.Helper()
		var count int
		err := f.Pool.QueryRow(ctx, `SELECT count(*) FROM ai_reviews WHERE submission_id=$1`, submission).Scan(&count)
		if err != nil || count != want {
			t.Fatalf("review count=%d want=%d err=%v", count, want, err)
		}
	}
	if err = reconcile(ctx, f.Pool, time.Now().Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	assertCount(id, 0)
	if err = reconcile(ctx, f.Pool, time.Now().Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err = reconcile(ctx, f.Pool, time.Now().Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	assertCount(id, 1)
	assertCount(wrongID, 0)
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(openAIEnvelope(`{"score":90,"summary":"Clear","dimensions":{"readability":90},"findings":[]}`)))
	}))
	defer provider.Close()
	if err := process(ctx, f.Pool, provider.URL, "", "test-model"); err != nil {
		t.Fatal(err)
	}
	var status string
	if err := f.Pool.QueryRow(ctx, `SELECT status FROM ai_reviews WHERE submission_id=$1`, id).Scan(&status); err != nil || status != "completed" {
		t.Fatalf("status=%s err=%v", status, err)
	}
	mux := http.NewServeMux()
	Register(mux, f.Pool, "secret", "internal")
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": f.UserID, "userId": f.UserID, "exp": time.Now().Add(time.Hour).Unix()}).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/api/v1/submissions/"+id+"/ai-review", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	testutil.Capture(t, w, req.Method, req.Pattern)
	var otherID string
	if err := f.Pool.QueryRow(ctx, `INSERT INTO users(email,username,password_hash) VALUES(gen_random_uuid()::text || '@example.org',gen_random_uuid()::text,'hash') RETURNING id`).Scan(&otherID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := f.Pool.Exec(context.Background(), `DELETE FROM users WHERE id=$1`, otherID); err != nil {
			t.Error(err)
		}
	})
	otherToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": otherID, "userId": otherID, "exp": time.Now().Add(time.Hour).Unix()}).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	req = httptest.NewRequest("GET", "/api/v1/submissions/"+id+"/ai-review", nil)
	req.Header.Set("Authorization", "Bearer "+otherToken)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != 404 {
		t.Fatalf("outsider status=%d body=%s", w.Code, w.Body.String())
	}

	if _, err := f.Pool.Exec(ctx, `DELETE FROM ai_reviews WHERE submission_id=$1`, id); err != nil {
		t.Fatal(err)
	}
	if err := reconcile(ctx, f.Pool, time.Now().Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	unavailable := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer unavailable.Close()
	for i, want := range []string{"queued", "queued", "failed"} {
		if err := process(ctx, f.Pool, unavailable.URL, "", "test-model"); err != nil {
			t.Fatal(err)
		}
		var attempts int
		var verdict string
		err := f.Pool.QueryRow(ctx, `SELECT r.status,r.attempt_count,s.status FROM ai_reviews r
			JOIN submissions s ON s.id=r.submission_id WHERE s.id=$1`, id).Scan(&status, &attempts, &verdict)
		if err != nil || status != want || attempts != i+1 || verdict != "Accepted" {
			t.Fatalf("status=%s attempts=%d verdict=%s err=%v", status, attempts, verdict, err)
		}
	}

	if _, err := f.Pool.Exec(ctx, `UPDATE ai_reviews SET status='queued', attempt_count=0,
		started_at=NULL, completed_at=NULL, error_message=NULL WHERE submission_id=$1`, id); err != nil {
		t.Fatal(err)
	}
	started, release := make(chan struct{}), make(chan struct{})
	oldProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		w.Write([]byte(openAIEnvelope(`{"score":12,"summary":"Stale","dimensions":{}}`)))
	}))
	defer oldProvider.Close()
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	oldDone := make(chan error, 1)
	go func() { oldDone <- process(ctx, f.Pool, oldProvider.URL, "", "old-model") }()
	select {
	case <-started:
	case err := <-oldDone:
		t.Fatalf("old attempt exited before provider call: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("old attempt did not reach provider")
	}
	if err := process(ctx, f.Pool, provider.URL, "", "test-model"); err != pgx.ErrNoRows {
		t.Fatalf("another attempt claimed an active review: %v", err)
	}
	if _, err := f.Pool.Exec(ctx, `UPDATE ai_reviews SET started_at=now()-interval '3 minutes' WHERE submission_id=$1`, id); err != nil {
		t.Fatal(err)
	}
	if err := process(ctx, f.Pool, provider.URL, "", "test-model"); err != nil {
		t.Fatal(err)
	}
	close(release)
	if err := <-oldDone; err != nil {
		t.Fatal(err)
	}
	var summary string
	var attempts int
	err = f.Pool.QueryRow(ctx, `SELECT status,summary,attempt_count FROM ai_reviews WHERE submission_id=$1`, id).Scan(&status, &summary, &attempts)
	if err != nil || status != "completed" || summary != "Clear" || attempts != 2 {
		t.Fatalf("replacement attempt lost: status=%s summary=%s attempts=%d err=%v", status, summary, attempts, err)
	}
	if _, err := f.Pool.Exec(ctx, `DELETE FROM ai_reviews WHERE submission_id=$1`, id); err != nil {
		t.Fatal(err)
	}
	if err := reconcile(ctx, f.Pool, time.Now().Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	badRequest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer badRequest.Close()
	if err := process(ctx, f.Pool, badRequest.URL, "", "test-model"); err != nil {
		t.Fatal(err)
	}
	err = f.Pool.QueryRow(ctx, `SELECT status,attempt_count FROM ai_reviews WHERE submission_id=$1`, id).Scan(&status, &attempts)
	if err != nil || status != "failed" || attempts != 1 {
		t.Fatalf("HTTP 400 should fail without retry: status=%s attempts=%d err=%v", status, attempts, err)
	}
}
