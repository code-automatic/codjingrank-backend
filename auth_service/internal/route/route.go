package route

import (
	"net/http"

	"innogen-backend/auth_service/internal/handler"
	"innogen-backend/shared/middleware"
)

// Register registers all auth service routes on the given ServeMux.
func Register(mux *http.ServeMux, h *handler.Handler) {
	mux.HandleFunc("POST /auth/login", h.Login)
	mux.HandleFunc("POST /auth/refresh", h.Refresh)
	mux.HandleFunc("POST /auth/logout", h.Logout)
	mux.HandleFunc("POST /auth/register", h.Register)
	mux.Handle("GET /auth/me", middleware.XUserID()(http.HandlerFunc(h.CurrentUser)))
	mux.Handle("GET /auth/github/connect", middleware.XUserID()(http.HandlerFunc(h.GithubConnect)))
	mux.HandleFunc("GET /auth/github/callback", h.GithubCallback)
	mux.Handle("GET /auth/github/status", middleware.XUserID()(http.HandlerFunc(h.GithubStatus)))
}
