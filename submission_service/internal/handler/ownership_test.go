package handler

import (
	"context"
	"innogen-backend/shared/middleware"
	"innogen-backend/shared/testutil"
	"innogen-backend/submission_service/internal/repository"
	"innogen-backend/submission_service/internal/service"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSubmissionOwnerBoundary(t *testing.T) {
	f := testutil.Catalog(t)
	other := testutil.Catalog(t)
	repo := repository.New(f.Pool)
	sub, err := repo.CreateSubmission(context.Background(), f.UserID, f.ProblemID, f.LanguageID, "print(3)")
	if err != nil {
		t.Fatal(err)
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := New(service.New(repo, nil, log), log)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /submissions/{id}", h.GetByID)
	for _, tc := range []struct {
		uid  string
		want int
	}{{f.UserID, 200}, {other.UserID, 404}} {
		req := httptest.NewRequest("GET", "/submissions/"+sub.ID, nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tc.uid))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Fatalf("identity=%s got %d %s", tc.uid, rec.Code, rec.Body.String())
		}
	}
}
