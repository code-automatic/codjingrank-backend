package route

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strings"

	"innogen-backend/shared/middleware"
)

// ProxySet holds all backend service proxies.
type ProxySet struct {
	AuthPublic *httputil.ReverseProxy // /auth/login (public, only strip X-User-*)
	Auth       *httputil.ReverseProxy // /auth/me (authenticated)
	Run        *httputil.ReverseProxy
	Submission *httputil.ReverseProxy
}

// stripHealthPath rewrites the incoming path to /health before forwarding.
func stripHealthPath(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/health"
		next.ServeHTTP(w, r)
	})
}

// stripUserHeaders removes only X-User-* headers (keeps Authorization intact).
// Used for public routes like /auth/login.
func stripUserHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("X-User-ID")
		r.Header.Del("X-User-Email")
		r.Header.Del("X-User-Role")
		next.ServeHTTP(w, r)
	})
}

// stripAllAuth removes Authorization and X-User-* headers.
// Used for authenticated proxy routes where the gateway validates JWT and injects claims.
func stripAllAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("Authorization")
		r.Header.Del("X-User-ID")
		r.Header.Del("X-User-Email")
		r.Header.Del("X-User-Role")
		next.ServeHTTP(w, r)
	})
}

// RegisterProxyRoutes registers all proxy routes on the mux.
func RegisterProxyRoutes(mux *http.ServeMux, proxies *ProxySet, log *slog.Logger, jwtSecret string) {
	authMW := middleware.Auth(jwtSecret)

	// Public auth route — only strip X-User-* headers, keep Authorization intact
	mux.Handle("POST /auth/login", stripUserHeaders(proxies.AuthPublic))
	mux.Handle("POST /auth/refresh", stripUserHeaders(proxies.AuthPublic))
	mux.Handle("POST /auth/logout", stripUserHeaders(proxies.AuthPublic))
	mux.Handle("POST /auth/register", stripUserHeaders(proxies.AuthPublic))

	// Authenticated routes — validate JWT, strip all auth headers, inject from claims
	mux.Handle("GET /auth/me", authMW(stripAllAuth(proxies.Auth)))

	mux.Handle("POST /run", authMW(stripAllAuth(proxies.Run)))

	mux.Handle("POST /submit", authMW(stripAllAuth(proxies.Submission)))
	mux.Handle("GET /submissions/{id}", authMW(stripAllAuth(proxies.Submission)))
	mux.Handle("GET /me/submissions", authMW(stripAllAuth(proxies.Submission)))
	mux.Handle("GET /me/submissions/{problemId}/latest", authMW(stripAllAuth(proxies.Submission)))
	mux.Handle("GET /me/stats", authMW(stripAllAuth(proxies.Submission)))

	// These upstreams now accept the same UUID request at root and v1.
	stripV1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api/v1")
			next.ServeHTTP(w, r)
		})
	}
	mux.Handle("POST /api/v1/run", authMW(stripAllAuth(stripV1(proxies.Run))))
	mux.Handle("POST /api/v1/submit", authMW(stripAllAuth(stripV1(proxies.Submission))))
	mux.Handle("GET /api/v1/submissions/{id}", authMW(stripAllAuth(stripV1(proxies.Submission))))
	// Health checks proxied to backend services
	mux.Handle("GET /health/auth", stripUserHeaders(stripHealthPath(proxies.AuthPublic)))
	mux.Handle("GET /health/run", stripUserHeaders(stripHealthPath(proxies.Run)))
	mux.Handle("GET /health/submission", stripUserHeaders(stripHealthPath(proxies.Submission)))
}
