package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"innogen-backend/run_service/internal/dto"
	"innogen-backend/run_service/internal/repository"
	"innogen-backend/shared/middleware"
	"innogen-backend/shared/piston"
	"innogen-backend/shared/testutil"
)

func TestPistonRuntimeIsMilliseconds(t *testing.T) {
	f := testutil.Catalog(t)
	engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(piston.Response{Run: &piston.Stage{Stdout: "3", CPUTime: 12}})
	}))
	defer engine.Close()
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, f.UserID)
	result, err := New(repository.New(f.Pool), piston.NewClient(engine.URL)).Run(ctx, dto.RunRequest{ProblemID: f.ProblemID, LanguageID: f.LanguageID, Code: "print(3)"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Results[0].RuntimeMs == nil || *result.Results[0].RuntimeMs != 12 {
		t.Fatalf("Piston cpu_time is milliseconds: %+v", result.Results[0])
	}
}

func TestPistonFailureIsNotCompletedRun(t *testing.T) {
	f := testutil.Catalog(t)
	engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "PRIVATE PROVIDER", 500) }))
	defer engine.Close()
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, f.UserID)
	result, err := New(repository.New(f.Pool), piston.NewClient(engine.URL)).Run(ctx, dto.RunRequest{ProblemID: f.ProblemID, LanguageID: f.LanguageID, Code: "print(3)"})
	if result != nil || err == nil {
		t.Fatalf("provider failure must be an error: %+v %v", result, err)
	}
	if errors.Is(err, context.Canceled) {
		t.Fatal("wrong error")
	}
}
