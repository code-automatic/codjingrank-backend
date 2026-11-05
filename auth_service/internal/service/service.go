package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"innogen-backend/auth_service/internal/dto"
	"innogen-backend/auth_service/internal/jwt"
	"innogen-backend/auth_service/internal/repository"
	"innogen-backend/shared/config"
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
)

// AuthService handles authentication business logic.
type AuthService struct {
	repo   *repository.UserRepository
	jwtSvc *jwt.Service
	cfg    *config.Config
}

// New creates a new AuthService with the given dependencies.
func New(repo *repository.UserRepository, jwtSvc *jwt.Service, cfg *config.Config) *AuthService {
	return &AuthService{repo: repo, jwtSvc: jwtSvc, cfg: cfg}
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
	_ = s.repo.CreateRefreshToken(ctx, user.ID, hashToken(rawRefresh), "", "", time.Now().Add(time.Duration(s.cfg.RefreshTokenTTLDays)*24*time.Hour))
	s.setRefreshCookie(w, rawRefresh)

	return &dto.LoginResponse{
		AccessToken: token,
		User:        dto.ToUserResponse(user),
	}, nil
}

// CurrentUser returns the user info for the given user ID.
func (s *AuthService) CurrentUser(ctx context.Context, userID int) (*dto.MeResponse, error) {
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
	_ = s.repo.RevokeRefreshToken(ctx, tokenID)
	newRaw, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}
	_ = s.repo.CreateRefreshToken(ctx, userID, hashToken(newRaw), r.UserAgent(), r.RemoteAddr, time.Now().Add(time.Duration(s.cfg.RefreshTokenTTLDays)*24*time.Hour))
	s.setRefreshCookie(w, newRaw)

	accessToken, err := s.jwtSvc.GenerateToken(user.ID, user.Email, user.Role, s.cfg.AccessTokenTTLMinutes)
	if err != nil {
		return nil, err
	}
	return &dto.LoginResponse{AccessToken: accessToken, User: dto.ToUserResponse(user)}, nil
}

// Logout revokes the refresh token and clears the cookie.
func (s *AuthService) Logout(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(s.cfg.RefreshCookieName)
	if err == nil {
		tokenHash := hashToken(cookie.Value)
		tokenID, _, _, _, _ := s.repo.FindRefreshToken(ctx, tokenHash)
		if tokenID != "" {
			_ = s.repo.RevokeRefreshToken(ctx, tokenID)
		}
	}
	s.clearRefreshCookie(w)
	return nil
}

// FrontendURL returns the configured frontend URL.
func (s *AuthService) FrontendURL() string {
	return s.cfg.FrontendURL
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
	http.SetCookie(w, &http.Cookie{
		Name: s.cfg.RefreshCookieName, Value: "", HttpOnly: true,
		Secure: s.cfg.RefreshCookieSecure, SameSite: parseSameSite(s.cfg.RefreshCookieSameSite),
		Path: s.cfg.RefreshCookiePath, MaxAge: -1,
	})
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

// GithubConnectURL generates the GitHub App installation URL with a state JWT.
func (s *AuthService) GithubConnectURL(userID int, userEmail string) string {
	stateToken, _ := s.jwtSvc.GenerateToken(userID, userEmail, "user", 5)
	return s.cfg.GitHubAppInstallURL + "?state=" + stateToken
}

// HandleGithubCallback processes the GitHub OAuth callback.
// Returns the frontend redirect URL.
func (s *AuthService) HandleGithubCallback(ctx context.Context, installationID string, stateToken string) (string, error) {
	// Parse state JWT to get userID
	claims, err := s.jwtSvc.ParseToken(stateToken)
	if err != nil {
		return s.cfg.FrontendURL + "/settings?github_status=error&message=invalid_state", nil
	}

	// Upsert github_account link
	if err := s.repo.UpsertGithubAccount(ctx, claims.UserID, installationID); err != nil {
		return "", err
	}

	// Try backfill owner from webhook data (if webhook arrived first)
	owner, ownerType, err := s.repo.GetGithubInstallationOwner(ctx, installationID)
	if err == nil && owner != "" {
		_ = s.repo.UpdateGithubAccountOwner(ctx, claims.UserID, owner, ownerType, "active")
	}

	return s.cfg.FrontendURL + "/settings?github_status=connected", nil
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

// GithubStatus returns the GitHub connection status for a user.
func (s *AuthService) GithubStatus(ctx context.Context, userID int) (*dto.GithubStatusResponse, error) {
	instID, owner, ownerType, status, err := s.repo.GetGithubAccountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if instID == "" {
		return &dto.GithubStatusResponse{Connected: false}, nil
	}
	return &dto.GithubStatusResponse{
		Connected:       true,
		InstallationID:  instID,
		GithubOwner:     owner,
		GithubOwnerType: ownerType,
		Status:          status,
	}, nil
}
