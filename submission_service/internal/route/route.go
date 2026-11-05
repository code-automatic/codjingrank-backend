package route

import (
	"net/http"

	"innogen-backend/shared/middleware"
	"innogen-backend/submission_service/internal/handler"
)

// Register registers all submission service routes on the given ServeMux.
func Register(mux *http.ServeMux, h *handler.Handler) {
	mux.Handle("POST /submit", middleware.XUserID()(http.HandlerFunc(h.Submit)))
	mux.Handle("GET /submissions/{id}", middleware.XUserID()(http.HandlerFunc(h.GetByID)))
	mux.Handle("GET /me/submissions", middleware.XUserID()(http.HandlerFunc(h.ListMySubmissions)))
	mux.Handle("GET /me/submissions/{problemId}/latest", middleware.XUserID()(http.HandlerFunc(h.GetLatestForProblem)))
	mux.Handle("GET /me/stats", middleware.XUserID()(http.HandlerFunc(h.GetStats)))
}
