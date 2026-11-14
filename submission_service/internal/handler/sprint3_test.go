package handler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"innogen-backend/shared/middleware"
	"innogen-backend/shared/testutil"
	"innogen-backend/submission_service/internal/repository"
	"innogen-backend/submission_service/internal/service"
)

func TestSubmissionAdminAndInactiveReadPolicy(t *testing.T) {
	f := testutil.Catalog(t)
	other := testutil.Catalog(t)
	ctx := context.Background()
	repo := repository.New(f.Pool)
	sub, err := repo.CreateSubmission(ctx, f.UserID, f.ProblemID, f.LanguageID, "print(3)")
	if err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := New(service.New(repo, nil, log), log)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /submissions/{id}", h.GetByID)
	request := func(uid string, want int) {
		t.Helper()
		req := httptest.NewRequest("GET", "/submissions/"+sub.ID, nil)
		req.Header.Set("X-User-Role", "admin")
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, uid))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("want %d got %d %s", want, rec.Code, rec.Body.String())
		}
		if want == 200 {
			testutil.Capture(t, rec, "GET", "/api/v1/submissions/{id}")
		}
	}
	request(other.UserID, 404)
	if _, err = f.Pool.Exec(ctx, `UPDATE users SET role='admin' WHERE id=$1`, other.UserID); err != nil {
		t.Fatal(err)
	}
	request(other.UserID, 200)
	if _, err = f.Pool.Exec(ctx, `UPDATE users SET is_active=false WHERE id=$1`, other.UserID); err != nil {
		t.Fatal(err)
	}
	request(other.UserID, 403)
	request(f.UserID, 200)
}

func TestLatestAbsentIsNull(t *testing.T) {
	f := testutil.Catalog(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := New(service.New(repository.New(f.Pool), nil, log), log)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /me/submissions/{problemId}/latest", h.GetLatestForProblem)
	req := httptest.NewRequest("GET", "/me/submissions/"+f.ProblemID+"/latest", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, f.UserID))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var env struct {
		Data struct {
			Submission any `json:"submission"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if rec.Code != 200 || env.Data.Submission != nil {
		t.Fatalf("latest absence: %d %s", rec.Code, rec.Body.String())
	}
	testutil.Capture(t, rec, "GET", "/api/v1/me/submissions/{problemId}/latest")
}
