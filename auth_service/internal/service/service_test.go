package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"innogen-backend/auth_service/internal/jwt"
	"innogen-backend/shared/config"
	"innogen-backend/shared/models"
)

// fakeStore is a minimal UserStore test double.
type fakeStore struct {
	findTokenID string
	findErr     error
	revoked     []string
	revokeErr   error
}

func (f *fakeStore) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}

func (f *fakeStore) FindByID(ctx context.Context, id string) (*models.User, error) {
	return nil, nil
}

func (f *fakeStore) CreateUser(ctx context.Context, email, passwordHash, username, fullName string) (*models.User, error) {
	return nil, nil
}

func (f *fakeStore) CreateRefreshToken(ctx context.Context, userID string, tokenHash, userAgent, ipAddress string, expiresAt time.Time) error {
	return nil
}

func (f *fakeStore) FindRefreshToken(ctx context.Context, tokenHash string) (string, string, time.Time, *time.Time, error) {
	return f.findTokenID, "", time.Now().Add(time.Hour), nil, f.findErr
}

func (f *fakeStore) RevokeRefreshToken(ctx context.Context, tokenID string) error {
	if f.revokeErr != nil {
		return f.revokeErr
	}
	f.revoked = append(f.revoked, tokenID)
	return nil
}

func (f *fakeStore) RotateRefreshToken(ctx context.Context, tokenID, userID, hash, userAgent, ip string, expires time.Time) (bool, error) {
	return false, nil
}

func newTestService(t *testing.T, store UserStore, cfg *config.Config) *AuthService {
	t.Helper()
	return New(store, jwt.NewService("test-secret"), cfg)
}

func TestValidateRequestOrigin(t *testing.T) {
	svc := newTestService(t, &fakeStore{}, &config.Config{
		CORSAllowedOrigins: "http://app.example.com, https://admin.example.com:8443",
	})

	tests := []struct {
		name    string
		origin  string
		referer string
		wantErr bool
	}{
		{name: "allowed origin", origin: "http://app.example.com"},
		{name: "allowed origin with port", origin: "https://admin.example.com:8443"},
		{name: "origin scheme mismatch", origin: "https://app.example.com", wantErr: true},
		{name: "origin prefix lookalike", origin: "http://app.example.com.evil.test", wantErr: true},
		{name: "origin not allowlisted", origin: "http://evil.test", wantErr: true},
		{name: "origin null", origin: "null", wantErr: true},
		{name: "referer allowed", referer: "http://app.example.com/settings/profile"},
		{name: "referer not allowlisted", referer: "http://evil.test/attack", wantErr: true},
		{name: "origin takes precedence over referer", origin: "http://evil.test", referer: "http://app.example.com/", wantErr: true},
		{name: "no browser headers"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.referer != "" {
				req.Header.Set("Referer", tt.referer)
			}
			err := svc.validateRequestOrigin(req)
			if tt.wantErr && !errors.Is(err, ErrForbiddenOrigin) {
				t.Fatalf("got %v, want ErrForbiddenOrigin", err)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("got %v, want nil", err)
			}
		})
	}
}

func TestLogoutRevokesRefreshToken(t *testing.T) {
	store := &fakeStore{findTokenID: "token-1"}
	svc := newTestService(t, store, &config.Config{
		RefreshCookieName: "refreshToken",
		RefreshCookiePath: "/",
	})

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "refreshToken", Value: "raw-token"})
	rec := httptest.NewRecorder()

	if err := svc.Logout(context.Background(), rec, req); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if len(store.revoked) != 1 || store.revoked[0] != "token-1" {
		t.Fatalf("revoked = %v, want [token-1]", store.revoked)
	}
}

func TestLogoutClearsCookieWhenRevokeFails(t *testing.T) {
	store := &fakeStore{findTokenID: "token-1", revokeErr: errors.New("db down")}
	svc := newTestService(t, store, &config.Config{
		RefreshCookieName: "refreshToken",
		RefreshCookiePath: "/",
	})

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "refreshToken", Value: "raw-token"})
	rec := httptest.NewRecorder()

	if err := svc.Logout(context.Background(), rec, req); err == nil {
		t.Fatal("expected revoke error to be reported")
	}
	if len(store.revoked) != 0 {
		t.Fatalf("unexpected successful revoke: %v", store.revoked)
	}
	if cookie := rec.Result().Cookies(); len(cookie) != 1 || cookie[0].MaxAge >= 0 {
		t.Fatalf("cookie not cleared: %+v", cookie)
	}
}

func TestClearRefreshCookieSetsDomain(t *testing.T) {
	svc := newTestService(t, &fakeStore{}, &config.Config{
		RefreshCookieName:   "refreshToken",
		RefreshCookiePath:   "/",
		RefreshCookieDomain: "app.example.com",
	})
	rec := httptest.NewRecorder()

	svc.clearRefreshCookie(rec)

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("got %d cookies, want 1", len(cookies))
	}
	if cookies[0].Domain != "app.example.com" {
		t.Fatalf("domain = %q, want app.example.com", cookies[0].Domain)
	}
	if cookies[0].Path != "/" {
		t.Fatalf("path = %q, want /", cookies[0].Path)
	}
}

func TestClearRefreshCookieOmitsEmptyDomain(t *testing.T) {
	svc := newTestService(t, &fakeStore{}, &config.Config{
		RefreshCookieName: "refreshToken",
		RefreshCookiePath: "/",
	})
	rec := httptest.NewRecorder()

	svc.clearRefreshCookie(rec)

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Domain != "" {
		t.Fatalf("unexpected domain: %+v", cookies)
	}
}
