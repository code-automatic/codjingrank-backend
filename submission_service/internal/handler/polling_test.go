package handler

import (
	"context"
	"encoding/json"
	"innogen-backend/shared/middleware"
	"innogen-backend/shared/models"
	"innogen-backend/shared/testutil"
	"innogen-backend/submission_service/internal/repository"
	"innogen-backend/submission_service/internal/service"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPollingLatestAndStatsContract(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.Background()
	repo := repository.New(f.Pool)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := New(service.New(repo, nil, log), log)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /submissions/{id}", h.GetByID)
	mux.HandleFunc("GET /me/submissions/{problemId}/latest", h.GetLatestForProblem)
	mux.HandleFunc("GET /me/stats", h.GetStats)
	read := func(path, pattern string, want int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest("GET", path, nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, f.UserID))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("%s %d %s", path, rec.Code, rec.Body.String())
		}
		testutil.Capture(t, rec, "GET", "GET "+pattern)
		return rec
	}
	read("/me/stats", "/api/v1/me/stats", 200)
	sub, e := repo.CreateSubmission(ctx, f.UserID, f.ProblemID, f.LanguageID, "print(3)")
	if e != nil {
		t.Fatal(e)
	}
	poll := func(want string) {
		t.Helper()
		rec := read("/submissions/"+sub.ID, "/api/v1/submissions/{id}", 200)
		var env struct {
			Data struct {
				Submission models.Submission `json:"submission"`
			} `json:"data"`
		}
		if e := json.Unmarshal(rec.Body.Bytes(), &env); e != nil {
			t.Fatal(e)
		}
		if env.Data.Submission.Status != want || env.Data.Submission.Code != "print(3)" {
			t.Fatalf("poll %+v", env)
		}
	}
	poll("Pending")
	a, e := repo.ClaimSubmission(ctx, sub.ID)
	if e != nil {
		t.Fatal(e)
	}
	poll("Running")
	if e = repo.UpdateSubmissionResult(ctx, sub.ID, a, "Accepted", nil, nil, nil, 1, 1); e != nil {
		t.Fatal(e)
	}
	poll("Accepted")
	read("/me/submissions/"+f.ProblemID+"/latest?languageId="+f.LanguageID, "/api/v1/me/submissions/{problemId}/latest", 200)
	read("/me/submissions/"+f.ProblemID+"/latest?languageId=00000000-0000-0000-0000-000000000001", "/api/v1/me/submissions/{problemId}/latest", 200)
	read("/me/submissions/"+f.ProblemID+"/latest?languageId=1", "/api/v1/me/submissions/{problemId}/latest", 400)
	read("/me/stats", "/api/v1/me/stats", 200)
}
