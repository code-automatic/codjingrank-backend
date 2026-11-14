package handler

import (
	"context"
	"encoding/json"
	"innogen-backend/run_service/internal/repository"
	"innogen-backend/run_service/internal/service"
	"innogen-backend/shared/middleware"
	"innogen-backend/shared/piston"
	"innogen-backend/shared/testutil"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRunHTTPContractAndProviderFailure(t *testing.T) {
	f := testutil.Catalog(t)
	for _, tc := range []struct {
		name, body       string
		provider, status int
	}{
		{"accepted", `{"run":{"code":0,"stdout":"3","cpu_time":12}}`, 200, 200},
		{"wrong answer", `{"run":{"code":0,"stdout":"0"}}`, 200, 200},
		{"malformed provider", `{}`, 200, 502},
		{"provider failure", `PRIVATE PROVIDER KEY`, 500, 502},
		{"provider timeout", "", 0, 504},
	} {
		t.Run(tc.name, func(t *testing.T) {
			engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.provider == 0 {
					select {
					case <-r.Context().Done():
					case <-time.After(2 * time.Second):
					}
					return
				}
				w.WriteHeader(tc.provider)
				w.Write([]byte(tc.body))
			}))
			defer engine.Close()
			h := New(service.New(repository.New(f.Pool), piston.NewClient(engine.URL)), slog.New(slog.NewTextHandler(io.Discard, nil)))
			raw, _ := json.Marshal(map[string]string{"problemId": f.ProblemID, "languageId": f.LanguageID, "code": "print(3)"})
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			ctx = context.WithValue(ctx, middleware.UserIDKey, f.UserID)
			req := httptest.NewRequest("POST", "/run", strings.NewReader(string(raw))).WithContext(ctx)
			rec := httptest.NewRecorder()
			h.Run(rec, req)
			if rec.Code != tc.status || strings.Contains(rec.Body.String(), "PRIVATE") {
				t.Fatalf("%d %s", rec.Code, rec.Body.String())
			}
			testutil.Capture(t, rec, "POST", "POST /api/v1/run")
		})
	}
}
