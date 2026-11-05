package oauth

import (
	"context"
	"fmt"
	"log/slog"

	"innogen-backend/repo_service/repository"
	"innogen-backend/shared/config"
)

type OAuthService struct {
	repo   *repository.RepoRepository
	client *OAuthClient
	cfg    *config.Config
	log    *slog.Logger
}

func NewOAuthService(repo *repository.RepoRepository, client *OAuthClient, cfg *config.Config, log *slog.Logger) *OAuthService {
	return &OAuthService{repo: repo, client: client, cfg: cfg, log: log}
}

func (s *OAuthService) GetStartURL(userID int) (*OAuthStartURLResponse, error) {
	state, err := GenerateState(s.cfg.JWTSecret, userID)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=read:user&state=%s",
		s.cfg.GitHubOAuthClientID, s.cfg.GitHubOAuthRedirectURL, state)
	return &OAuthStartURLResponse{URL: url}, nil
}

func (s *OAuthService) HandleCallback(ctx context.Context, code, stateToken string) (string, error) {
	claims, err := ValidateState(stateToken, s.cfg.JWTSecret)
	if err != nil {
		return s.cfg.GitHubOAuthFrontendRedirectURL + "?oauth=error&message=invalid_state", nil
	}
	accessToken, err := s.client.ExchangeCode(ctx, code, s.cfg.GitHubOAuthRedirectURL)
	if err != nil {
		s.log.Error("oauth token exchange failed", slog.String("error", err.Error()))
		return s.cfg.GitHubOAuthFrontendRedirectURL + "?oauth=error&message=token_exchange_failed", nil
	}
	user, err := s.client.GetUser(ctx, accessToken)
	if err != nil {
		s.log.Error("oauth user lookup failed", slog.String("error", err.Error()))
		return s.cfg.GitHubOAuthFrontendRedirectURL + "?oauth=error&message=user_lookup_failed", nil
	}
	noreplyEmail := fmt.Sprintf("%d+%s@users.noreply.github.com", user.ID, user.Login)
	if err := s.repo.UpsertGitHubOAuth(ctx, claims.UserID, fmt.Sprintf("%d", user.ID), user.Login, user.AvatarURL, noreplyEmail, user.Login); err != nil {
		s.log.Error("oauth upsert failed", slog.String("error", err.Error()))
		return s.cfg.GitHubOAuthFrontendRedirectURL + "?oauth=error&message=db_error", nil
	}
	s.log.Info("oauth account linked", slog.Int("userId", claims.UserID), slog.String("githubUser", user.Login))
	return s.cfg.GitHubOAuthFrontendRedirectURL + "?oauth=connected", nil
}

func (s *OAuthService) GetAccount(ctx context.Context, userID int) (*GithubAccountResponse, error) {
	ghID, ghUser, ghAvatar, noreply, authorName, oauthStatus, err := s.repo.GetGitHubOAuthByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if oauthStatus != "connected" {
		return &GithubAccountResponse{Connected: false}, nil
	}
	return &GithubAccountResponse{
		Connected:          true,
		GithubUserID:       strPtr(ghID),
		GithubUsername:     strPtr(ghUser),
		GithubAvatarURL:    strPtr(ghAvatar),
		GithubNoreplyEmail: strPtr(noreply),
		CommitAuthorName:   strPtr(authorName),
		OAuthStatus:        strPtr(oauthStatus),
	}, nil
}

func (s *OAuthService) Disconnect(ctx context.Context, userID int) error {
	return s.repo.DisconnectGitHubOAuth(ctx, userID)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
