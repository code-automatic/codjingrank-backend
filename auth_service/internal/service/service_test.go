package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

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

	user    *models.User
	userErr error

	reset    *models.PasswordReset
	resetErr error

	createdResetUserID string
	createdOTPHash     string
	createdExpiresAt   time.Time
	createResetErr     error
	incrementErr       error
	invalidated        []string
	revokedAll         []string
	completed          bool
	completedResetID   string
	completedUserID    string
	completedPassword  string
	completedOK        bool
	completedErr       error
}

func (f *fakeStore) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return f.user, f.userErr
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

func (f *fakeStore) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	f.revokedAll = append(f.revokedAll, userID)
	return nil
}

func (f *fakeStore) CreatePasswordReset(ctx context.Context, userID, otpHash string, expiresAt time.Time) error {
	if f.createResetErr != nil {
		return f.createResetErr
	}
	f.createdResetUserID = userID
	f.createdOTPHash = otpHash
	f.createdExpiresAt = expiresAt
	return nil
}

func (f *fakeStore) FindLatestPasswordReset(ctx context.Context, userID string) (*models.PasswordReset, error) {
	return f.reset, f.resetErr
}

func (f *fakeStore) IncrementPasswordResetAttempts(ctx context.Context, id string) (int, error) {
	if f.incrementErr != nil {
		return 0, f.incrementErr
	}
	if f.reset != nil {
		f.reset.Attempts++
		return f.reset.Attempts, nil
	}
	return 0, nil
}

func (f *fakeStore) InvalidatePasswordReset(ctx context.Context, id string) error {
	f.invalidated = append(f.invalidated, id)
	if f.reset != nil {
		now := time.Now()
		f.reset.UsedAt = &now
	}
	return nil
}

func (f *fakeStore) CompletePasswordReset(ctx context.Context, resetID, userID, passwordHash string) (bool, error) {
	if f.completedErr != nil {
		return false, f.completedErr
	}
	f.completed = true
	f.completedResetID = resetID
	f.completedUserID = userID
	f.completedPassword = passwordHash
	return f.completedOK, nil
}

type fakeMailer struct {
	sent bool
	to   string
	otp  string
	ttl  time.Duration
	err  error
}

func (m *fakeMailer) SendOTP(ctx context.Context, to, otp string, ttl time.Duration) error {
	m.sent = true
	m.to = to
	m.otp = otp
	m.ttl = ttl
	return m.err
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

func newPasswordResetConfig() *config.Config {
	return &config.Config{
		PasswordResetTTLMinutes:  30,
		PasswordResetMaxAttempts: 5,
	}
}

func TestForgotPasswordUnknownEmailIsGeneric(t *testing.T) {
	store := &fakeStore{}
	mailer := &fakeMailer{}
	svc := newTestService(t, store, newPasswordResetConfig()).SetMailer(mailer)

	if err := svc.ForgotPassword(context.Background(), "missing@example.com"); err != nil {
		t.Fatalf("forgot password: %v", err)
	}
	if store.createdResetUserID != "" {
		t.Fatalf("reset row created for unknown email: %q", store.createdResetUserID)
	}
	if mailer.sent {
		t.Fatal("email sent for unknown email")
	}
}

func TestForgotPasswordStoresHashedOTPAndSendsEmail(t *testing.T) {
	user := &models.User{ID: "user-1", Email: "user@example.com"}
	store := &fakeStore{user: user}
	mailer := &fakeMailer{}
	svc := newTestService(t, store, newPasswordResetConfig()).SetMailer(mailer)

	before := time.Now()
	if err := svc.ForgotPassword(context.Background(), "uSeR@example.com"); err != nil {
		t.Fatalf("forgot password: %v", err)
	}

	if store.createdResetUserID != user.ID {
		t.Fatalf("reset user = %q, want %q", store.createdResetUserID, user.ID)
	}
	if !mailer.sent || mailer.to != user.Email {
		t.Fatalf("email not sent to user: %+v", mailer)
	}
	if len(mailer.otp) != 6 {
		t.Fatalf("otp %q is not 6 digits", mailer.otp)
	}
	for _, r := range mailer.otp {
		if r < '0' || r > '9' {
			t.Fatalf("otp %q is not numeric", mailer.otp)
		}
	}
	if store.createdOTPHash == mailer.otp {
		t.Fatal("OTP stored in plaintext")
	}
	if store.createdOTPHash != hashToken(mailer.otp) {
		t.Fatal("stored hash does not match emailed OTP")
	}
	if mailer.ttl != 30*time.Minute {
		t.Fatalf("ttl = %v, want 30m", mailer.ttl)
	}
	if store.createdExpiresAt.Before(before.Add(29 * time.Minute)) {
		t.Fatalf("expiry too early: %v", store.createdExpiresAt)
	}
}

func TestForgotPasswordSurvivesMailFailure(t *testing.T) {
	user := &models.User{ID: "user-1", Email: "user@example.com"}
	store := &fakeStore{user: user}
	mailer := &fakeMailer{err: errors.New("smtp down")}
	svc := newTestService(t, store, newPasswordResetConfig()).SetMailer(mailer)

	if err := svc.ForgotPassword(context.Background(), user.Email); err != nil {
		t.Fatalf("mail failure leaked: %v", err)
	}
	if store.createdResetUserID != user.ID {
		t.Fatal("reset row missing despite mail failure")
	}
}

func TestResetPasswordSuccessUpdatesHashAndRevokesTokens(t *testing.T) {
	user := &models.User{ID: "user-1", Email: "user@example.com"}
	reset := &models.PasswordReset{
		ID: "reset-1", UserID: user.ID, OTPHash: hashToken("123456"),
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	store := &fakeStore{user: user, reset: reset, completedOK: true}
	svc := newTestService(t, store, newPasswordResetConfig())

	if err := svc.ResetPassword(context.Background(), user.Email, "123456", "newpassword"); err != nil {
		t.Fatalf("reset password: %v", err)
	}
	if !store.completed {
		t.Fatal("password reset was not persisted")
	}
	if store.completedResetID != reset.ID || store.completedUserID != user.ID {
		t.Fatalf("completed wrong row: %+v", store)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(store.completedPassword), []byte("newpassword")); err != nil {
		t.Fatalf("new password not bcrypt hashed: %v", err)
	}
	if reset.Attempts != 0 {
		t.Fatalf("attempts changed on success: %d", reset.Attempts)
	}
}

func TestResetPasswordWrongOTPIncrementsAttempts(t *testing.T) {
	user := &models.User{ID: "user-1", Email: "user@example.com"}
	reset := &models.PasswordReset{
		ID: "reset-1", UserID: user.ID, OTPHash: hashToken("123456"),
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	store := &fakeStore{user: user, reset: reset, completedOK: true}
	svc := newTestService(t, store, newPasswordResetConfig())

	err := svc.ResetPassword(context.Background(), user.Email, "000000", "newpassword")
	if !errors.Is(err, ErrInvalidResetCode) {
		t.Fatalf("err = %v, want ErrInvalidResetCode", err)
	}
	if reset.Attempts != 1 {
		t.Fatalf("attempts = %d, want 1", reset.Attempts)
	}
	if store.completed {
		t.Fatal("password changed with wrong OTP")
	}
	if len(store.invalidated) != 0 {
		t.Fatalf("premature invalidation: %v", store.invalidated)
	}
}

func TestResetPasswordLocksAfterMaxAttempts(t *testing.T) {
	user := &models.User{ID: "user-1", Email: "user@example.com"}
	reset := &models.PasswordReset{
		ID: "reset-1", UserID: user.ID, OTPHash: hashToken("123456"),
		ExpiresAt: time.Now().Add(10 * time.Minute), Attempts: 4,
	}
	store := &fakeStore{user: user, reset: reset, completedOK: true}
	svc := newTestService(t, store, newPasswordResetConfig())

	err := svc.ResetPassword(context.Background(), user.Email, "000000", "newpassword")
	if !errors.Is(err, ErrInvalidResetCode) {
		t.Fatalf("err = %v, want ErrInvalidResetCode", err)
	}
	if reset.Attempts != 5 {
		t.Fatalf("attempts = %d, want 5", reset.Attempts)
	}
	if len(store.invalidated) != 1 || store.invalidated[0] != reset.ID {
		t.Fatalf("reset not invalidated: %v", store.invalidated)
	}
	if reset.UsedAt == nil {
		t.Fatal("lock did not mark the code used")
	}
}

func TestResetPasswordRejectsExpiredCode(t *testing.T) {
	user := &models.User{ID: "user-1", Email: "user@example.com"}
	reset := &models.PasswordReset{
		ID: "reset-1", UserID: user.ID, OTPHash: hashToken("123456"),
		ExpiresAt: time.Now().Add(-time.Minute),
	}
	store := &fakeStore{user: user, reset: reset, completedOK: true}
	svc := newTestService(t, store, newPasswordResetConfig())

	err := svc.ResetPassword(context.Background(), user.Email, "123456", "newpassword")
	if !errors.Is(err, ErrInvalidResetCode) {
		t.Fatalf("err = %v, want ErrInvalidResetCode", err)
	}
	if store.completed {
		t.Fatal("expired code accepted")
	}
	if len(store.invalidated) != 0 {
		t.Fatal("expired code should not consume attempts")
	}
}

func TestResetPasswordRejectsMissingOrUsedCode(t *testing.T) {
	user := &models.User{ID: "user-1", Email: "user@example.com"}
	store := &fakeStore{user: user}
	svc := newTestService(t, store, newPasswordResetConfig())

	if err := svc.ResetPassword(context.Background(), user.Email, "123456", "newpassword"); !errors.Is(err, ErrInvalidResetCode) {
		t.Fatalf("err = %v, want ErrInvalidResetCode", err)
	}
}

func TestResetPasswordRejectsUnknownEmail(t *testing.T) {
	store := &fakeStore{}
	svc := newTestService(t, store, newPasswordResetConfig())

	if err := svc.ResetPassword(context.Background(), "missing@example.com", "123456", "newpassword"); !errors.Is(err, ErrInvalidResetCode) {
		t.Fatalf("err = %v, want ErrInvalidResetCode", err)
	}
}

func TestResetPasswordRejectsAttemptsAtMax(t *testing.T) {
	user := &models.User{ID: "user-1", Email: "user@example.com"}
	reset := &models.PasswordReset{
		ID: "reset-1", UserID: user.ID, OTPHash: hashToken("123456"),
		ExpiresAt: time.Now().Add(10 * time.Minute), Attempts: 5,
	}
	store := &fakeStore{user: user, reset: reset, completedOK: true}
	svc := newTestService(t, store, newPasswordResetConfig())

	if err := svc.ResetPassword(context.Background(), user.Email, "123456", "newpassword"); !errors.Is(err, ErrInvalidResetCode) {
		t.Fatalf("err = %v, want ErrInvalidResetCode", err)
	}
	if reset.Attempts != 5 {
		t.Fatalf("attempts = %d, want 5", reset.Attempts)
	}
	if store.completed {
		t.Fatal("locked code accepted")
	}
}

func TestResetPasswordRejectsShortPassword(t *testing.T) {
	user := &models.User{ID: "user-1", Email: "user@example.com"}
	store := &fakeStore{user: user}
	svc := newTestService(t, store, newPasswordResetConfig())

	if err := svc.ResetPassword(context.Background(), user.Email, "123456", "12345"); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
	if store.completed {
		t.Fatal("short password accepted")
	}
}
