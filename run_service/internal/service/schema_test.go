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
	"strings"
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
		if req.RunTimeoutMs == nil || *req.RunTimeoutMs != 1000 {
			t.Errorf("timeout missing: %+v", req)
		}
		if req.RunMemoryLimit == nil || *req.RunMemoryLimit != 128*1024*1024 {
			t.Errorf("memory limit missing: %+v", req)
		}
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

func TestRunHidesProviderDiagnostics(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, f.UserID)
	engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("PRIVATE PROVIDER KEY"))
	}))
	defer engine.Close()
	svc := New(repository.New(f.Pool), piston.NewClient(engine.URL))
	result, err := svc.Run(ctx, dto.RunRequest{ProblemID: f.ProblemID, LanguageID: f.LanguageID, Code: "print(3)"})
	if err != ErrExecutionUnavailable || result != nil {
		t.Fatalf("unsafe provider failure: %+v %v", result, err)
	}
	raw, _ := json.Marshal(err.Error())
	if strings.Contains(string(raw), "PRIVATE PROVIDER KEY") {
		t.Fatalf("provider leaked: %s", raw)
	}
}

func TestRunFunctionHidesWrapperDiagnostics(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.WithValue(context.Background(), middleware.UserIDKey, f.UserID)
	if _, err := f.Pool.Exec(ctx, `UPDATE problems SET execution_mode='function',function_name='solve' WHERE id=$1`, f.ProblemID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Pool.Exec(ctx, `UPDATE code_templates SET wrapper_code='PRIVATE WRAPPER' WHERE problem_id=$1`, f.ProblemID); err != nil {
		t.Fatal(err)
	}
	engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(piston.Response{Compile: &piston.Stage{Code: 1, Stderr: "PRIVATE WRAPPER failed"}})
	}))
	defer engine.Close()
	svc := New(repository.New(f.Pool), piston.NewClient(engine.URL))
	result, err := svc.Run(ctx, dto.RunRequest{ProblemID: f.ProblemID, LanguageID: f.LanguageID, Code: "invalid"})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(result)
	if strings.Contains(string(raw), "PRIVATE WRAPPER") {
		t.Fatalf("wrapper leaked: %s", raw)
	}
}
