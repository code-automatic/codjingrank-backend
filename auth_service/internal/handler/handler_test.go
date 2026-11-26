package handler

import (
	"context"
	"errors"
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

type stubStore struct {
	findTokenID string
	revokeErr   error
	revoked     []string
}

func (s *stubStore) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}

func (s *stubStore) FindByID(ctx context.Context, id string) (*models.User, error) {
	return nil, nil
}

func (s *stubStore) CreateUser(ctx context.Context, email, passwordHash, username, fullName string) (*models.User, error) {
	return nil, nil
}

func (s *stubStore) CreateRefreshToken(ctx context.Context, userID string, tokenHash, userAgent, ipAddress string, expiresAt time.Time) error {
	return nil
}

func (s *stubStore) FindRefreshToken(ctx context.Context, tokenHash string) (string, string, time.Time, *time.Time, error) {
	return s.findTokenID, "", time.Time{}, nil, nil
}

func (s *stubStore) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	if s.revokeErr != nil {
		return s.revokeErr
	}
	s.revoked = append(s.revoked, tokenID)
	return nil
}

func (s *stubStore) RotateRefreshToken(ctx context.Context, tokenID, userID, hash, userAgent, ip string, expires time.Time) (bool, error) {
	return false, nil
}

func newTestHandler(store *stubStore) *Handler {
	cfg := &config.Config{
		RefreshCookieName:  "refreshToken",
		RefreshCookiePath:  "/",
		CORSAllowedOrigins: "http://app.example.com",
	}
	svc := service.New(store, jwt.NewService("test-secret"), cfg)
	return New(svc, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestRefreshAndLogoutRejectForbiddenOrigin(t *testing.T) {
	h := newTestHandler(&stubStore{})

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
	h := newTestHandler(&stubStore{})

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

func TestLogoutSucceedsWhenRevokeFails(t *testing.T) {
	h := newTestHandler(&stubStore{findTokenID: "token-1", revokeErr: errors.New("db down")})
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Origin", "http://app.example.com")
	req.AddCookie(&http.Cookie{Name: "refreshToken", Value: "raw-token"})
	rec := httptest.NewRecorder()

	h.Logout(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge >= 0 {
		t.Fatalf("cookie not cleared: %+v", cookies)
	}
}
