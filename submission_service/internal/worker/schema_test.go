package worker

import (
	"context"
	"encoding/json"
	"innogen-backend/shared/piston"
	"innogen-backend/shared/testutil"
	"innogen-backend/submission_service/internal/repository"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWorkerUUIDSchemaAndDuplicateDelivery(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.Background()
	repo := repository.New(f.Pool)
	sub, err := repo.CreateSubmission(ctx, f.UserID, f.ProblemID, f.LanguageID, "print(3)")
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req piston.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
		}
		calls++
		if req.Stdin != "HIDDEN INPUT" || req.Files[0].Name != "main.py" {
			t.Errorf("wrong execution input: %+v", req)
		}
		json.NewEncoder(w).Encode(piston.Response{Run: &piston.Stage{Code: 0, Stdout: "3", CPUTime: 0.001}})
	}))
	defer engine.Close()
	worker := New(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, piston.NewClient(engine.URL), nil)
	if err = worker.processJob(ctx, sub.ID); err != nil {
		t.Fatal(err)
	}
	got, err := repo.FindByID(ctx, sub.ID)
	if err != nil || got.Status != "Accepted" || got.PassCount != 1 {
		t.Fatalf("verdict: %+v %v", got, err)
	}
	if err = worker.processJob(ctx, sub.ID); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("duplicate delivery executed %d times", calls)
	}
}

func TestHiddenDiagnosticsNeverReachSubmission(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.Background()
	repo := repository.New(f.Pool)
	sub, err := repo.CreateSubmission(ctx, f.UserID, f.ProblemID, f.LanguageID, "print(3)")
	if err != nil {
		t.Fatal(err)
	}
	engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(piston.Response{Run: &piston.Stage{Code: 1, Stderr: "PRIVATE WRAPPER: HIDDEN INPUT"}})
	}))
	defer engine.Close()
	worker := New(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, piston.NewClient(engine.URL), nil)
	if err = worker.processJob(ctx, sub.ID); err != nil {
		t.Fatal(err)
	}
	got, err := repo.FindByID(ctx, sub.ID)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(got)
	if strings.Contains(string(raw), "HIDDEN INPUT") || strings.Contains(string(raw), "PRIVATE WRAPPER") {
		t.Fatalf("hidden diagnostics leaked: %s", raw)
	}
}
