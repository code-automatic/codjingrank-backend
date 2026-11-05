package route

import (
	"net/http"

	"innogen-backend/run_service/internal/handler"
	"innogen-backend/shared/middleware"
)

// Register registers all run service routes on the given ServeMux.
func Register(mux *http.ServeMux, h *handler.Handler) {
	mux.Handle("POST /run", middleware.XUserID()(http.HandlerFunc(h.Run)))
}
