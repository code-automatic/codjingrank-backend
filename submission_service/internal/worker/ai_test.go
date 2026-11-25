package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"innogen-backend/shared/piston"
	"innogen-backend/shared/testutil"
	"innogen-backend/submission_service/internal/repository"
)

func TestReviewNotificationFollowsCommittedAccepted(t *testing.T) {
	for _, tc := range []struct {
		verdict, output string
		wantCalls       int32
	}{{"Accepted", "3", 1}, {"Wrong Answer", "4", 0}} {
		t.Run(tc.verdict, func(t *testing.T) {
			f := testutil.Catalog(t)
			ctx := context.Background()
			repo := repository.New(f.Pool)
			sub, err := repo.CreateSubmission(ctx, f.UserID, f.ProblemID, f.LanguageID, "source")
			if err != nil {
				t.Fatal(err)
			}
			var notifications atomic.Int32
			ai := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				notifications.Add(1)
				got, err := repo.FindByID(r.Context(), sub.ID)
				if err != nil || got == nil || got.Status != "Accepted" {
					t.Errorf("notification arrived before Accepted commit: %+v, %v", got, err)
				}
				if r.Method != "POST" || r.URL.Path != "/internal/reviews/"+sub.ID || r.Header.Get("Authorization") != "Bearer worker-test" {
					t.Error("notification violated the internal enqueue contract")
				}
				w.WriteHeader(503)
			}))
			defer ai.Close()
			t.Setenv("AI_SERVICE_URL", ai.URL)
			t.Setenv("AI_INTERNAL_SECRET", "worker-test")
			engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				json.NewEncoder(w).Encode(piston.Response{Run: &piston.Stage{Stdout: tc.output}})
			}))
			defer engine.Close()
			w := New(slog.Default(), repo, piston.NewClient(engine.URL), nil)
			for i := 0; i < 2; i++ {
				if err := w.processJob(ctx, sub.ID); err != nil {
					t.Fatal(err)
				}
			}
			got, err := repo.FindByID(ctx, sub.ID)
			if err != nil || got == nil || got.Status != tc.verdict || notifications.Load() != tc.wantCalls {
				t.Fatalf("submission=%+v notification count=%d err=%v", got, notifications.Load(), err)
			}
		})
	}
}
