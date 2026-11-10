package service

import (
	"context"
	"encoding/json"
	"innogen-backend/run_service/internal/dto"
	"innogen-backend/run_service/internal/repository"
	"innogen-backend/shared/middleware"
	"innogen-backend/shared/piston"
	"innogen-backend/shared/testutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRunUsesOnlySamplesSynchronously(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, f.UserID)
	calls := 0
	engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req piston.Request
		json.NewDecoder(r.Body).Decode(&req)
		calls++
		if req.Stdin != "1 2" {
			t.Errorf("not a sample: %+v", req)
		}
		json.NewEncoder(w).Encode(piston.Response{Run: &piston.Stage{Code: 0, Stdout: "3"}})
	}))
	defer engine.Close()
	svc := New(repository.New(f.Pool), piston.NewClient(engine.URL))
	result, err := svc.Run(ctx, dto.RunRequest{ProblemID: f.ProblemID, LanguageID: f.LanguageID, Code: "print(3)"})
	if err != nil || result.Status != "Accepted" || calls != 1 {
		t.Fatalf("result: %+v calls=%d err=%v", result, calls, err)
	}
	var count int
	if err = f.Pool.QueryRow(ctx, `SELECT count(*) FROM submissions WHERE user_id=$1`, f.UserID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("Run created submission: %d %v", count, err)
	}
}
