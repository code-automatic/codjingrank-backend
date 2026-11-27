package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

	user      *models.User
	reset     *models.PasswordReset
	completed bool
}

func (s *stubStore) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.user, nil
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

func (s *stubStore) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	return nil
}

func (s *stubStore) CreatePasswordReset(ctx context.Context, userID, otpHash string, expiresAt time.Time) error {
	return nil
}

func (s *stubStore) FindLatestPasswordReset(ctx context.Context, userID string) (*models.PasswordReset, error) {
	return s.reset, nil
}

func (s *stubStore) LatestPasswordResetCreatedAt(ctx context.Context, userID string) (time.Time, bool, error) {
	return time.Time{}, false, nil
}

func (s *stubStore) IncrementPasswordResetAttempts(ctx context.Context, id string) (int, error) {
	return 0, nil
}

func (s *stubStore) InvalidatePasswordReset(ctx context.Context, id string) error {
	return nil
}

func (s *stubStore) CompletePasswordReset(ctx context.Context, resetID, userID, passwordHash string) (bool, error) {
	s.completed = true
	return true, nil
}

func newTestHandler(store *stubStore) *Handler {
	cfg := &config.Config{
		JWTSecret:                "test-secret",
		OTPSecret:                "otp-test-secret",
		RefreshCookieName:        "refreshToken",
		RefreshCookiePath:        "/",
		CORSAllowedOrigins:       "http://app.example.com",
		PasswordResetTTLMinutes:  30,
		PasswordResetMaxAttempts: 5,
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

// otpHash mirrors the service's keyed OTP digest for the test config.
func otpHash(otp string) string {
	mac := hmac.New(sha256.New, []byte("otp-test-secret"))
	mac.Write([]byte(otp))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestForgotPasswordHandlerAlwaysReturns202(t *testing.T) {
	for _, tc := range []struct {
		name string
		user *models.User
	}{
		{name: "existing email", user: &models.User{ID: "user-1", Email: "user@example.com"}},
		{name: "unknown email"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestHandler(&stubStore{user: tc.user})
			req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", strings.NewReader(`{"email":"user@example.com"}`))
			rec := httptest.NewRecorder()

			h.ForgotPassword(rec, req)

			if rec.Code != http.StatusAccepted {
				t.Fatalf("status = %d, want 202", rec.Code)
			}
			if !strings.Contains(rec.Body.String(), "If the email exists, an OTP has been sent") {
				t.Fatalf("body = %s", rec.Body.String())
			}
		})
	}
}

func TestResetPasswordHandlerSuccessIs204(t *testing.T) {
	user := &models.User{ID: "user-1", Email: "user@example.com"}
	store := &stubStore{user: user, reset: &models.PasswordReset{
		ID: "reset-1", UserID: user.ID, OTPHash: otpHash("123456"),
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}}
	h := newTestHandler(store)
	req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", strings.NewReader(`{"email":"user@example.com","otp":"123456","newPassword":"newpassword1"}`))
	rec := httptest.NewRecorder()

	h.ResetPassword(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("204 response has body: %s", rec.Body.String())
	}
	if !store.completed {
		t.Fatal("password reset not applied")
	}
}

func TestResetPasswordHandlerInvalidOTPIs400(t *testing.T) {
	user := &models.User{ID: "user-1", Email: "user@example.com"}
	store := &stubStore{user: user, reset: &models.PasswordReset{
		ID: "reset-1", UserID: user.ID, OTPHash: otpHash("123456"),
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}}
	h := newTestHandler(store)
	req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", strings.NewReader(`{"email":"user@example.com","otp":"000000","newPassword":"newpassword1"}`))
	rec := httptest.NewRecorder()

	h.ResetPassword(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "RESET_TOKEN_INVALID") {
		t.Fatalf("body = %s, want RESET_TOKEN_INVALID", rec.Body.String())
	}
	if store.completed {
		t.Fatal("password changed with invalid OTP")
	}
}

func TestResetPasswordHandlerShortPasswordIs400(t *testing.T) {
	h := newTestHandler(&stubStore{})
	req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", strings.NewReader(`{"email":"user@example.com","otp":"123456","newPassword":"12345"}`))
	rec := httptest.NewRecorder()

	h.ResetPassword(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestResetPasswordHandlerMalformedBodyIs400(t *testing.T) {
	h := newTestHandler(&stubStore{})
	req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", strings.NewReader(`{"email":`))
	rec := httptest.NewRecorder()

	h.ResetPassword(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
