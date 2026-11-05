package proxy

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	"innogen-backend/shared/middleware"
	"innogen-backend/shared/response"
)

// NewProxy creates a reverse proxy to the given target URL.
// No user headers are injected (for public routes like /auth/login).
func NewProxy(targetURL string, log *slog.Logger) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = &http.Transport{
		ResponseHeaderTimeout: 30 * time.Second,
	}
	proxy.ErrorHandler = proxyErrorHandler(log)
	return proxy, nil
}

// NewAuthenticatedProxy creates a reverse proxy that injects X-User-* headers
// from JWT claims stored in the request context.
func NewAuthenticatedProxy(targetURL string, log *slog.Logger) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = &http.Transport{
		ResponseHeaderTimeout: 30 * time.Second,
	}

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		if userID, ok := middleware.GetUserID(req); ok {
			req.Header.Set("X-User-ID", strconv.Itoa(userID))
		}
		if email, ok := middleware.GetUserEmail(req); ok {
			req.Header.Set("X-User-Email", email)
		}
		if role, ok := middleware.GetUserRole(req); ok {
			req.Header.Set("X-User-Role", role)
		}
	}

	proxy.ErrorHandler = proxyErrorHandler(log)
	return proxy, nil
}

func proxyErrorHandler(log *slog.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		log.Error("proxy error",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("error", err.Error()),
		)
		response.ErrorSimple(w, http.StatusBadGateway, "Bad gateway")
	}
}
