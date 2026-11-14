package worker

import (
	"context"
	"encoding/json"
	"innogen-backend/shared/piston"
	"innogen-backend/shared/testutil"
	"innogen-backend/submission_service/internal/queue"
	"innogen-backend/submission_service/internal/repository"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// Run separately with dedicated Redis to avoid consuming another package's jobs.
func TestRedisRecoveryAndDuplicateDelivery(t *testing.T) {
	addr := os.Getenv("RECOVERY_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("set dedicated RECOVERY_TEST_REDIS_ADDR; run this integration test separately")
	}
	f := testutil.Catalog(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	repo := repository.New(f.Pool)
	q, err := queue.New(addr)
	if err != nil {
		t.Fatal(err)
	}
	defer q.Close()
	sub, err := repo.CreateSubmission(ctx, f.UserID, f.ProblemID, f.LanguageID, "print(3)")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = repo.ClaimSubmission(ctx, sub.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = f.Pool.Exec(ctx, `UPDATE submissions SET claimed_at=now()-interval '5 minutes' WHERE id=$1`, sub.ID); err != nil {
		t.Fatal(err)
	}
	calls := 0
	engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		json.NewEncoder(w).Encode(piston.Response{Run: &piston.Stage{Stdout: "3", CPUTime: 12}})
	}))
	defer engine.Close()
	worker := New(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, piston.NewClient(engine.URL), q)
	worker.reconcileStaleSubmissions(ctx)
	if err = worker.processOne(ctx); err != nil {
		t.Fatal(err)
	}
	got, err := repo.FindByID(ctx, sub.ID)
	if err != nil || got.Status != "Accepted" || got.RuntimeMs == nil || *got.RuntimeMs != 12 {
		t.Fatalf("recovered verdict: %+v %v", got, err)
	}
	if err = q.Enqueue(ctx, sub.ID); err != nil {
		t.Fatal(err)
	}
	if err = worker.processOne(ctx); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("duplicate execution count=%d", calls)
	}
}
