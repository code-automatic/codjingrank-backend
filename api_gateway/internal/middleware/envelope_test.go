package middleware

import (
	"encoding/json"
	sharedMW "innogen-backend/shared/middleware"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddlewareEnvelope(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("unexpected downstream call") })
	limiter := &RateLimiter{visitors: map[string][]time.Time{}, limit: 0, window: time.Minute}
	for _, tc := range []struct {
		name    string
		handler http.Handler
		code    int
		machine string
	}{
		{"auth", sharedMW.Auth("test")(next), 401, "UNAUTHORIZED"},
		{"rate limit", RateLimit(limiter)(next), 429, "RATE_LIMITED"},
		{"panic", Recover(log, http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("private error") })), 500, "INTERNAL_ERROR"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tc.handler.ServeHTTP(w, httptest.NewRequest("POST", "/run", nil))
			var env map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
				t.Fatal(err)
			}
			if w.Code != tc.code || env["status"] != "error" || env["code"] != float64(tc.code) || env["error"] != tc.machine || len(env) != 5 {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
			if env["details"] != nil {
				t.Fatal("details not null")
			}
		})
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("OPTIONS", "/run", nil)
	r.Header.Set("Origin", "http://localhost:3000")
	CORS("http://localhost:3000")(next).ServeHTTP(w, r)
	if w.Code != 204 || w.Body.Len() != 0 {
		t.Fatalf("preflight %d %s", w.Code, w.Body.String())
	}
}
