package handler

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"innogen-backend/auth_service/internal/jwt"
	"innogen-backend/auth_service/internal/service"
	"innogen-backend/shared/config"
	"innogen-backend/shared/models"
)

type stubStore struct{}

func (stubStore) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}

func (stubStore) FindByID(ctx context.Context, id string) (*models.User, error) {
	return nil, nil
}

func (stubStore) CreateUser(ctx context.Context, email, passwordHash, username, fullName string) (*models.User, error) {
	return nil, nil
}

func (stubStore) CreateRefreshToken(ctx context.Context, userID string, tokenHash, userAgent, ipAddress string, expiresAt time.Time) error {
	return nil
}

func (stubStore) FindRefreshToken(ctx context.Context, tokenHash string) (string, string, time.Time, *time.Time, error) {
	return "", "", time.Time{}, nil, nil
}

func (stubStore) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	return nil
}

func (stubStore) RotateRefreshToken(ctx context.Context, tokenID, userID, hash, userAgent, ip string, expires time.Time) (bool, error) {
	return false, nil
}

func newTestHandler() *Handler {
	cfg := &config.Config{
		RefreshCookieName:  "refreshToken",
		RefreshCookiePath:  "/",
		CORSAllowedOrigins: "http://app.example.com",
	}
	svc := service.New(stubStore{}, jwt.NewService("test-secret"), cfg)
	return New(svc, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestRefreshAndLogoutRejectForbiddenOrigin(t *testing.T) {
	h := newTestHandler()

	for _, tc := range []struct {
		name    string
		handler http.HandlerFunc
	}{
		{name: "refresh", handler: h.Refresh},
		{name: "logout", handler: h.Logout},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/auth/"+tc.name, nil)
			req.Header.Set("Origin", "http://evil.test")
			rec := httptest.NewRecorder()

			tc.handler(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403", rec.Code)
			}
			if !strings.Contains(rec.Body.String(), `"FORBIDDEN"`) {
				t.Fatalf("body = %s, want FORBIDDEN code", rec.Body.String())
			}
		})
	}
}

func TestRefreshAndLogoutAllowConfiguredOrAbsentOrigin(t *testing.T) {
	h := newTestHandler()

	for _, tc := range []struct {
		name       string
		handler    http.HandlerFunc
		origin     string
		wantStatus int
	}{
		{name: "refresh allowed origin", handler: h.Refresh, origin: "http://app.example.com", wantStatus: http.StatusUnauthorized},
		{name: "refresh no origin", handler: h.Refresh, wantStatus: http.StatusUnauthorized},
		{name: "logout allowed origin", handler: h.Logout, origin: "http://app.example.com", wantStatus: http.StatusOK},
		{name: "logout no origin", handler: h.Logout, wantStatus: http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			rec := httptest.NewRecorder()

			tc.handler(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
		})
	}
}
