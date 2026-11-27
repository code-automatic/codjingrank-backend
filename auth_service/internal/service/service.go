package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"innogen-backend/auth_service/internal/dto"
	"innogen-backend/auth_service/internal/jwt"
	"innogen-backend/shared/config"
	"innogen-backend/shared/models"
)

// Sentinel errors for the service layer.
var (
	ErrInvalidInput        = errors.New("invalid input")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrAccountInactive     = errors.New("account is inactive")
	ErrUserNotFound        = errors.New("user not found")
	ErrEmailTaken          = errors.New("email already registered")
	ErrUsernameTaken       = errors.New("username already taken")
	ErrRefreshTokenMissing = errors.New("refresh token missing")
	ErrRefreshTokenInvalid = errors.New("refresh token invalid")
	ErrRefreshTokenRevoked = errors.New("refresh token revoked")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrForbiddenOrigin     = errors.New("forbidden request origin")
	ErrInvalidResetCode    = errors.New("invalid or expired reset code")
)

// UserStore is the persistence behaviour required by AuthService.
type UserStore interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
	CreateUser(ctx context.Context, email, passwordHash, username, fullName string) (*models.User, error)
	CreateRefreshToken(ctx context.Context, userID string, tokenHash, userAgent, ipAddress string, expiresAt time.Time) error
	FindRefreshToken(ctx context.Context, tokenHash string) (id string, userID string, expiresAt time.Time, revokedAt *time.Time, err error)
	RevokeRefreshToken(ctx context.Context, tokenID string) error
	RotateRefreshToken(ctx context.Context, tokenID, userID, hash, userAgent, ip string, expires time.Time) (bool, error)
	RevokeAllUserRefreshTokens(ctx context.Context, userID string) error
	CreatePasswordReset(ctx context.Context, userID, otpHash string, expiresAt time.Time) error
	LatestPasswordResetCreatedAt(ctx context.Context, userID string) (time.Time, bool, error)
	FindLatestPasswordReset(ctx context.Context, userID string) (*models.PasswordReset, error)
	IncrementPasswordResetAttempts(ctx context.Context, id string) (int, error)
	InvalidatePasswordReset(ctx context.Context, id string) error
	CompletePasswordReset(ctx context.Context, resetID, userID, passwordHash string) (bool, error)
}

// OTPMailer delivers a password-reset one-time code to the user's email.
type OTPMailer interface {
	SendOTP(ctx context.Context, to, otp string, ttl time.Duration) error
}

// AuthService handles authentication business logic.
type AuthService struct {
	repo   UserStore
	jwtSvc *jwt.Service
	cfg    *config.Config
	mailer OTPMailer
}

// New creates a new AuthService with the given dependencies.
func New(repo UserStore, jwtSvc *jwt.Service, cfg *config.Config) *AuthService {
	return &AuthService{repo: repo, jwtSvc: jwtSvc, cfg: cfg}
}

// SetMailer assigns the optional OTP delivery dependency. Without a mailer the
// service still creates reset codes, which keeps development and tests working.
func (s *AuthService) SetMailer(m OTPMailer) *AuthService {
	s.mailer = m
	return s
}

// Login authenticates a user by email and password, returning a JWT token and user info.
func (s *AuthService) Login(ctx context.Context, w http.ResponseWriter, r *http.Request, req dto.LoginRequest) (*dto.LoginResponse, error) {
	if msg := req.Validate(); msg != "" {
		return nil, ErrInvalidInput
	}

	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.jwtSvc.GenerateToken(user.ID, user.Email, user.Role, s.cfg.AccessTokenTTLMinutes)
	if err != nil {
		return nil, err
	}

	rawRefresh, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	if err = s.repo.CreateRefreshToken(ctx, user.ID, hashToken(rawRefresh), "", "", time.Now().Add(time.Duration(s.cfg.RefreshTokenTTLDays)*24*time.Hour)); err != nil {
		return nil, err
	}
	s.setRefreshCookie(w, rawRefresh)

	return &dto.LoginResponse{
		AccessToken: token,
		User:        dto.ToUserResponse(user),
	}, nil
}

// CurrentUser returns the user info for the given user ID.
func (s *AuthService) CurrentUser(ctx context.Context, userID string) (*dto.MeResponse, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	return &dto.MeResponse{
		User: dto.ToUserResponse(user),
	}, nil
}

// Refresh validates a refresh token and issues a new access token with token rotation.
func (s *AuthService) Refresh(ctx context.Context, w http.ResponseWriter, r *http.Request) (*dto.LoginResponse, error) {
	if err := s.validateRequestOrigin(r); err != nil {
		return nil, err
	}
	cookie, err := r.Cookie(s.cfg.RefreshCookieName)
	if err != nil {
		return nil, ErrRefreshTokenMissing
	}
	rawToken := cookie.Value
	tokenHash := hashToken(rawToken)
	tokenID, userID, expiresAt, revokedAt, err := s.repo.FindRefreshToken(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if tokenID == "" {
		return nil, ErrRefreshTokenInvalid
	}
	if revokedAt != nil {
		return nil, ErrRefreshTokenRevoked
	}
	if time.Now().After(expiresAt) {
		return nil, ErrRefreshTokenExpired
	}
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.IsActive {
		return nil, ErrUserNotFound
	}

	// rotate: revoke old, create new
	newRaw, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	rotated, err := s.repo.RotateRefreshToken(ctx, tokenID, userID, hashToken(newRaw), r.UserAgent(), ip, time.Now().Add(time.Duration(s.cfg.RefreshTokenTTLDays)*24*time.Hour))
	if err != nil {
		return nil, err
	}
	if !rotated {
		return nil, ErrRefreshTokenRevoked
	}
	s.setRefreshCookie(w, newRaw)

	accessToken, err := s.jwtSvc.GenerateToken(user.ID, user.Email, user.Role, s.cfg.AccessTokenTTLMinutes)
	if err != nil {
		return nil, err
	}
	return &dto.LoginResponse{AccessToken: accessToken, User: dto.ToUserResponse(user)}, nil
}

// Logout revokes the refresh token and clears the cookie.
func (s *AuthService) Logout(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if err := s.validateRequestOrigin(r); err != nil {
		return err
	}
	cookie, err := r.Cookie(s.cfg.RefreshCookieName)
	if err == nil && cookie.Value != "" {
		tokenID, _, _, _, findErr := s.repo.FindRefreshToken(ctx, hashToken(cookie.Value))
		if findErr != nil {
			s.clearRefreshCookie(w)
			return findErr
		}
		if tokenID != "" {
			// Best effort: the cookie is cleared even when revocation fails.
			if revokeErr := s.repo.RevokeRefreshToken(ctx, tokenID); revokeErr != nil {
				s.clearRefreshCookie(w)
				return revokeErr
			}
		}
	}
	s.clearRefreshCookie(w)
	return nil
}

// validateRequestOrigin rejects state-changing browser requests whose Origin
// (or, when Origin is absent, Referer) is not in the configured allowlist.
// Requests without either header (server-to-server, CLI) are allowed.
func (s *AuthService) validateRequestOrigin(r *http.Request) error {
	allowed := parseAllowedOrigins(s.cfg.CORSAllowedOrigins)
	if origin := strings.TrimSpace(r.Header.Get("Origin")); origin != "" {
		if !originAllowed(origin, allowed) {
			return ErrForbiddenOrigin
		}
		return nil
	}
	if referer := strings.TrimSpace(r.Header.Get("Referer")); referer != "" {
		if !originAllowed(referer, allowed) {
			return ErrForbiddenOrigin
		}
	}
	return nil
}

func parseAllowedOrigins(raw string) map[string]struct{} {
	allowed := make(map[string]struct{})
	for _, entry := range strings.Split(raw, ",") {
		if origin, ok := normalizeOrigin(entry); ok {
			allowed[origin] = struct{}{}
		}
	}
	return allowed
}

func originAllowed(raw string, allowed map[string]struct{}) bool {
	origin, ok := normalizeOrigin(raw)
	if !ok {
		return false
	}
	_, ok = allowed[origin]
	return ok
}

// normalizeOrigin reduces a URL (an Origin header or a Referer URL) to a
// lowercase scheme://host[:port] origin for exact allowlist comparison.
func normalizeOrigin(raw string) (string, bool) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", false
	}
	return strings.ToLower(u.Scheme) + "://" + strings.ToLower(u.Host), true
}

func (s *AuthService) setRefreshCookie(w http.ResponseWriter, token string) {
	maxAge := s.cfg.RefreshTokenTTLDays * 86400
	cookie := &http.Cookie{
		Name:     s.cfg.RefreshCookieName,
		Value:    token,
		HttpOnly: true,
		Secure:   s.cfg.RefreshCookieSecure,
		SameSite: parseSameSite(s.cfg.RefreshCookieSameSite),
		Path:     s.cfg.RefreshCookiePath,
		MaxAge:   maxAge,
	}
	if s.cfg.RefreshCookieDomain != "" {
		cookie.Domain = s.cfg.RefreshCookieDomain
	}
	http.SetCookie(w, cookie)
}

func (s *AuthService) clearRefreshCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name: s.cfg.RefreshCookieName, Value: "", HttpOnly: true,
		Secure: s.cfg.RefreshCookieSecure, SameSite: parseSameSite(s.cfg.RefreshCookieSameSite),
		Path: s.cfg.RefreshCookiePath, MaxAge: -1,
	}
	if s.cfg.RefreshCookieDomain != "" {
		cookie.Domain = s.cfg.RefreshCookieDomain
	}
	http.SetCookie(w, cookie)
}

func parseSameSite(s string) http.SameSite {
	switch s {
	case "Strict":
		return http.SameSiteStrictMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Register creates a new user account.
func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.LoginResponse, error) {
	if msg := req.Validate(); msg != "" {
		return nil, fmt.Errorf("%w: %s", ErrInvalidInput, msg)
	}

	// Check if email exists
	existing, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailTaken
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user, err := s.repo.CreateUser(ctx, req.Email, string(hash), req.Username, req.FullName)
	if err != nil {
		return nil, err
	}

	// Generate JWT and return same as login
	token, err := s.jwtSvc.GenerateToken(user.ID, user.Email, user.Role, s.cfg.AccessTokenTTLMinutes)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken: token,
		User:        dto.ToUserResponse(user),
	}, nil
}

// ForgotPassword issues a password-reset OTP when the email belongs to an
// account. It never reveals whether the email exists: unknown addresses and
// delivery failures are handled without a distinguishable result.
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrInvalidInput
	}

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}

	// Throttle repeated requests: keep one OTP per cooldown window and stay
	// silent so the response shape never changes.
	latest, found, err := s.repo.LatestPasswordResetCreatedAt(ctx, user.ID)
	if err != nil {
		return err
	}
	if found && time.Since(latest) < passwordResetCooldown {
		return nil
	}

	otp, err := generateOTP()
	if err != nil {
		return fmt.Errorf("failed to generate reset code: %w", err)
	}
	ttl := s.passwordResetTTL()
	if err := s.repo.CreatePasswordReset(ctx, user.ID, s.hashOTP(otp), time.Now().Add(ttl)); err != nil {
		return err
	}

	// Delivery is best effort: a mail outage must not expose account existence.
	if s.mailer != nil {
		_ = s.mailer.SendOTP(ctx, user.Email, otp, ttl)
	}
	return nil
}

// ResetPassword verifies a one-time code and replaces the user's password.
// All failures return the same sentinel so callers cannot enumerate accounts
// or learn whether a code was expired, reused, or simply wrong.
func (s *AuthService) ResetPassword(ctx context.Context, w http.ResponseWriter, email, otp, newPassword string) error {
	email = strings.TrimSpace(email)
	if email == "" || strings.TrimSpace(otp) == "" {
		return ErrInvalidInput
	}
	if msg := dto.ValidatePasswordPolicy(newPassword); msg != "" {
		return ErrInvalidInput
	}

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrInvalidResetCode
	}

	reset, err := s.repo.FindLatestPasswordReset(ctx, user.ID)
	if err != nil {
		return err
	}
	if reset == nil {
		return ErrInvalidResetCode
	}
	maxAttempts := s.passwordResetMaxAttempts()
	if reset.Attempts >= maxAttempts {
		return ErrInvalidResetCode
	}
	if !time.Now().Before(reset.ExpiresAt) {
		return ErrInvalidResetCode
	}

	if !s.otpMatches(reset.OTPHash, otp) {
		attempts, err := s.repo.IncrementPasswordResetAttempts(ctx, reset.ID)
		if err != nil {
			return err
		}
		if attempts >= maxAttempts {
			if err := s.repo.InvalidatePasswordReset(ctx, reset.ID); err != nil {
				return err
			}
		}
		return ErrInvalidResetCode
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	ok, err := s.repo.CompletePasswordReset(ctx, reset.ID, user.ID, string(hash))
	if err != nil {
		return err
	}
	if !ok {
		return ErrInvalidResetCode
	}
	// Refresh tokens were revoked in the same transaction; clear the cookie so
	// the client cannot keep presenting a now-invalid token.
	s.clearRefreshCookie(w)
	return nil
}

func (s *AuthService) passwordResetTTL() time.Duration {
	if s.cfg.PasswordResetTTLMinutes <= 0 {
		return 30 * time.Minute
	}
	return time.Duration(s.cfg.PasswordResetTTLMinutes) * time.Minute
}

func (s *AuthService) passwordResetMaxAttempts() int {
	if s.cfg.PasswordResetMaxAttempts <= 0 {
		return 5
	}
	return s.cfg.PasswordResetMaxAttempts
}

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// passwordResetCooldown is the minimum interval between OTP issuance for one
// account, which keeps the generic endpoint from being used to spam a mailbox.
const passwordResetCooldown = time.Minute

// otpSecret returns the key used to MAC one-time codes. A dedicated secret is
// preferred; JWT_SECRET is a backwards-compatible fallback.
func (s *AuthService) otpSecret() []byte {
	if s.cfg.OTPSecret != "" {
		return []byte(s.cfg.OTPSecret)
	}
	return []byte(s.cfg.JWTSecret)
}

// hashOTP stores a keyed digest so a leaked table cannot be brute-forced into
// the six-digit code space the way an unsalted SHA-256 would allow.
func (s *AuthService) hashOTP(otp string) string {
	mac := hmac.New(sha256.New, s.otpSecret())
	mac.Write([]byte(otp))
	return hex.EncodeToString(mac.Sum(nil))
}

// otpMatches compares a stored keyed digest with a candidate code in constant
// time.
func (s *AuthService) otpMatches(storedHash, candidate string) bool {
	return subtle.ConstantTimeCompare([]byte(storedHash), []byte(s.hashOTP(candidate))) == 1
}
