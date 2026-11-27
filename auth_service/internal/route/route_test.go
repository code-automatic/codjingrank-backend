package route

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"innogen-backend/auth_service/internal/handler"
	"innogen-backend/auth_service/internal/jwt"
	"innogen-backend/auth_service/internal/service"
	"innogen-backend/shared/config"
)

func TestPasswordResetRoutesRegistered(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := service.New(nil, jwt.NewService("test-secret"), &config.Config{})
	mux := http.NewServeMux()
	Register(mux, handler.New(svc, log))

	for _, path := range []string{"/auth/forgot-password", "/auth/reset-password"} {
		_, pattern := mux.Handler(httptest.NewRequest(http.MethodPost, path, nil))
		if pattern == "" {
			t.Errorf("route not registered: POST %s", path)
		}
	}
}
