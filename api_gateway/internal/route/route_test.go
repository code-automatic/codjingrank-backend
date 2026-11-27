package route

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"

	"innogen-backend/api_gateway/internal/proxy"
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
		{"POST", "/auth/forgot-password"}, {"POST", "/auth/reset-password"},
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

func TestRetiredCatalogReturnsGone(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRetiredCatalog(mux)
	for _, path := range []string{"/problems/daily-challenge", "/problems/anything", "/problems/1/test-cases", "/admin/problems", "/admin/languages/1"} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != 410 {
			t.Fatalf("%s got %d", path, rec.Code)
		}
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func signTestJWT(t *testing.T, userID string) string {
	t.Helper()
	claims := jwtlib.MapClaims{
		"userId": userID,
		"email":  "learner@example.com",
		"role":   "student",
		"sub":    userID,
		"exp":    time.Now().Add(time.Hour).Unix(),
	}
	token, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign test token: %v", err)
	}
	return token
}

// TestV1AuthPublicFacade verifies that /api/v1/auth/* reaches Auth with the
// prefix stripped and that client-supplied X-User-* headers are removed.
func TestV1AuthPublicFacade(t *testing.T) {
	var gotPath, gotAuth, gotUserID string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotUserID = r.Header.Get("X-User-ID")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	p, err := proxy.NewProxy(upstream.URL, testLogger())
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterProxyRoutes(mux, &ProxySet{AuthPublic: p, Auth: p, Run: p, Submission: p}, testLogger(), "test-secret")

	for _, tc := range []struct{ path, want string }{
		{"/api/v1/auth/login", "/auth/login"},
		{"/api/v1/auth/register", "/auth/register"},
		{"/api/v1/auth/refresh", "/auth/refresh"},
		{"/api/v1/auth/logout", "/auth/logout"},
		{"/api/v1/auth/forgot-password", "/auth/forgot-password"},
		{"/api/v1/auth/reset-password", "/auth/reset-password"},
	} {
		gotPath, gotAuth, gotUserID = "", "", ""
		req := httptest.NewRequest("POST", tc.path, nil)
		req.Header.Set("Authorization", "Bearer client-token")
		req.Header.Set("X-User-ID", "spoofed-id")
		req.Header.Set("X-User-Email", "spoof@example.com")
		req.Header.Set("X-User-Role", "admin")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("%s: got %d", tc.path, rec.Code)
		}
		if gotPath != tc.want {
			t.Errorf("%s forwarded as %s, want %s", tc.path, gotPath, tc.want)
		}
		if gotUserID != "" {
			t.Errorf("%s leaked X-User-ID: %s", tc.path, gotUserID)
		}
		if gotAuth != "Bearer client-token" {
			t.Errorf("%s Authorization changed: %q", tc.path, gotAuth)
		}
	}
}

// TestV1UsersMeRewritesToAuthMe verifies the v1 profile facade maps to Auth
// GET /auth/me and forwards identity derived from the verified token only.
func TestV1UsersMeRewritesToAuthMe(t *testing.T) {
	const userID = "b1000000-0000-4000-8000-000000000001"
	var gotPath, gotUserID, gotEmail, gotAuth string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotUserID = r.Header.Get("X-User-ID")
		gotEmail = r.Header.Get("X-User-Email")
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer upstream.Close()

	p, err := proxy.NewAuthenticatedProxy(upstream.URL, testLogger())
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterProxyRoutes(mux, &ProxySet{AuthPublic: p, Auth: p, Run: p, Submission: p}, testLogger(), "test-secret")

	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+signTestJWT(t, userID))
	req.Header.Set("X-User-ID", "spoofed-id")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("got %d, want 204", rec.Code)
	}
	if gotPath != "/auth/me" {
		t.Errorf("forwarded as %s, want /auth/me", gotPath)
	}
	if gotUserID != userID {
		t.Errorf("X-User-ID = %q, want %q", gotUserID, userID)
	}
	if gotEmail != "learner@example.com" {
		t.Errorf("X-User-Email = %q", gotEmail)
	}
	if gotAuth != "" {
		t.Errorf("Authorization leaked upstream: %q", gotAuth)
	}
}

func TestV1UsersMeRequiresBearer(t *testing.T) {
	mux := http.NewServeMux()
	p := &httputil.ReverseProxy{}
	RegisterProxyRoutes(mux, &ProxySet{AuthPublic: p, Auth: p, Run: p, Submission: p}, testLogger(), "test-secret")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/users/me", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", rec.Code)
	}
}

func TestV1ExecutionRoutesProtected(t *testing.T) {
	mux := http.NewServeMux()
	p := &httputil.ReverseProxy{}
	RegisterProxyRoutes(mux, &ProxySet{AuthPublic: p, Auth: p, Run: p, Submission: p}, slog.New(slog.NewTextHandler(io.Discard, nil)), "test-secret")
	for _, path := range []string{"/api/v1/run", "/api/v1/submit"} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("POST", path, nil))
		if rec.Code != 401 {
			t.Fatalf("%s got %d", path, rec.Code)
		}
	}
}
