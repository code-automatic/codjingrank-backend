package route

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"
)

func TestProxySurfaceAfterLegacyRemoval(t *testing.T) {
	mux := http.NewServeMux()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	proxy := &httputil.ReverseProxy{}
	RegisterProxyRoutes(mux, &ProxySet{AuthPublic: proxy, Auth: proxy, Run: proxy, Submission: proxy}, log, "test-secret")
	for _, path := range []string{"/repositories", "/repositories/1/commits", "/github/connection", "/github/oauth/callback", "/webhooks/github", "/auth/github/connect", "/auth/github/status", "/auth/github/callback", "/health/repo"} {
		for _, method := range []string{"GET", "POST"} {
			_, pattern := mux.Handler(httptest.NewRequest(method, path, nil))
			if pattern != "" {
				t.Errorf("retired route %s %s still registered as %s", method, path, pattern)
			}
		}
	}
	for _, route := range []struct{ method, path string }{
		{"POST", "/auth/login"}, {"POST", "/auth/register"}, {"POST", "/auth/refresh"}, {"POST", "/auth/logout"},
		{"GET", "/auth/me"}, {"POST", "/run"}, {"POST", "/submit"}, {"GET", "/submissions/id"}, {"GET", "/me/submissions"},
	} {
		r := httptest.NewRequest(route.method, route.path, nil)
		_, pattern := mux.Handler(r)
		if pattern == "" {
			t.Errorf("retained route missing: %s %s", route.method, route.path)
		}
		if route.path == "/run" || route.path == "/submit" || route.path == "/auth/me" {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			if w.Code != http.StatusUnauthorized {
				t.Errorf("%s without auth: got %d", route.path, w.Code)
			}
		}
	}
}
