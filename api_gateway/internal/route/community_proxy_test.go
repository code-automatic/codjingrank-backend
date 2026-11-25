package route

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCommunityProxyPreservesBearerAndDoesNotExposeInternalEnqueue(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" || r.Header.Get("X-User-ID") != "" ||
			r.URL.Path != "/api/v1/submissions/test-id/ai-review" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()
	mux := http.NewServeMux()
	if err := RegisterCommunityProxies(mux, slog.Default(), upstream.URL, upstream.URL); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/api/v1/submissions/test-id/ai-review", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("X-User-ID", "client-supplied")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 204 {
		t.Fatalf("proxy returned %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("POST", "/internal/reviews/test-id", nil))
	if rec.Code != 404 {
		t.Fatalf("internal enqueue unexpectedly routed: %d", rec.Code)
	}
}
