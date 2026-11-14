package handler

import (
	"context"
	"encoding/json"
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

func TestHistoryPaginationFiltersAndOwnership(t *testing.T) {
	f := testutil.Catalog(t)
	other := testutil.Catalog(t)
	ctx := context.Background()
	repo := repository.New(f.Pool)
	ids := []string{}
	for _, code := range []string{"print(3)", "print(0)"} {
		sub, e := repo.CreateSubmission(ctx, f.UserID, f.ProblemID, f.LanguageID, code)
		if e != nil {
			t.Fatal(e)
		}
		ids = append(ids, sub.ID)
		if _, e = f.Pool.Exec(ctx, `UPDATE submissions SET created_at='2026-11-01T12:00:00Z' WHERE id=$1`, sub.ID); e != nil {
			t.Fatal(e)
		}
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := New(service.New(repo, nil, log), log)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /me/submissions", h.ListMySubmissions)
	for _, tc := range []struct {
		query, uid     string
		want, total, n int
	}{
		{"?page=1&size=1", f.UserID, 200, 2, 1}, {"?page=2&size=1", f.UserID, 200, 2, 1}, {"?page=3&size=1", f.UserID, 200, 2, 0},
		{"?status=Accepted", f.UserID, 200, 0, 0}, {"?problemId=" + other.ProblemID, f.UserID, 200, 0, 0}, {"?languageId=" + f.LanguageID + "&status=Pending", f.UserID, 200, 2, 2},
		{"", other.UserID, 200, 0, 0}, {"?page=0", f.UserID, 400, 0, 0}, {"?size=101", f.UserID, 400, 0, 0}, {"?status=bogus", f.UserID, 400, 0, 0}, {"?problemId=1", f.UserID, 400, 0, 0},
	} {
		req := httptest.NewRequest("GET", "/me/submissions"+tc.query, nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, tc.uid))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Fatalf("%s: %d %s", tc.query, rec.Code, rec.Body.String())
		}
		if tc.want != 200 {
			continue
		}
		var env struct {
			Data struct {
				Items []map[string]any `json:"items"`
				Total int              `json:"total"`
			} `json:"data"`
		}
		if e := json.Unmarshal(rec.Body.Bytes(), &env); e != nil {
			t.Fatal(e)
		}
		if env.Data.Total != tc.total || len(env.Data.Items) != tc.n || env.Data.Items == nil {
			t.Fatalf("history %s: %s", tc.query, rec.Body.String())
		}
		if tc.query == "?page=1&size=1" && env.Data.Items[0]["id"] != max(ids[0], ids[1]) {
			t.Fatal("unstable tie order")
		}
		if tc.query == "?page=2&size=1" && env.Data.Items[0]["id"] != min(ids[0], ids[1]) {
			t.Fatal("duplicate item across pages")
		}
		for _, item := range env.Data.Items {
			if _, ok := item["code"]; ok {
				t.Fatal("source leaked in history")
			}
		}
		testutil.Capture(t, rec, "GET", "GET /api/v1/me/submissions")
	}
}
