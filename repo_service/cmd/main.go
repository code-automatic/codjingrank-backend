package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"innogen-backend/repo_service/internal/githubapp"
	"innogen-backend/repo_service/internal/handler"
	"innogen-backend/repo_service/internal/oauth"
	"innogen-backend/repo_service/internal/route"
	"innogen-backend/repo_service/internal/webhook"
	"innogen-backend/repo_service/repository"
	"innogen-backend/repo_service/service"
	"innogen-backend/shared/config"
	"innogen-backend/shared/database"
	"innogen-backend/shared/logger"
	"innogen-backend/shared/response"
)

func requestLogger(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Info("request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Duration("duration", time.Since(start)),
		)
	})
}

func main() {
	cfg := config.Load()
	log := logger.New("repo-service")

	ctx := context.Background()

	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	var githubClient githubapp.GitHubClient
	if cfg.GitHubAppID != "" && cfg.GitHubPrivateKeyPath != "" {
		realClient, err := githubapp.NewRealClient(cfg.GitHubAppID, cfg.GitHubPrivateKeyPath, cfg.GitHubAPIBaseURL)
		if err != nil {
			log.Error("failed to create github client", slog.String("error", err.Error()))
			os.Exit(1)
		}
		githubClient = realClient
	} else {
		log.Warn("github app not configured, commit endpoint will fail")
		githubClient = nil // service will handle nil
	}

	repoRepo := repository.New(pool)

	// Webhook
	webhookSvc := webhook.NewWebhookService(repoRepo, log)
	webhookH := webhook.NewWebhookHandler(webhookSvc, cfg.GitHubWebhookSecret, log)
	if cfg.GitHubWebhookSecret == "" {
		log.Warn("GITHUB_WEBHOOK_SECRET is empty — all webhook requests will be rejected")
	}

	svc := service.New(repoRepo, githubClient, cfg, log)
	h := handler.New(svc, log)

	// OAuth
	oauthClient := oauth.NewOAuthClient(cfg.GitHubOAuthClientID, cfg.GitHubOAuthClientSecret)
	oauthSvc := oauth.NewOAuthService(repoRepo, oauthClient, cfg, log)
	oauthH := oauth.NewOAuthHandler(oauthSvc, log)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	route.Register(mux, h, webhookH, oauthH)

	addr := ":" + cfg.RepoServicePort
	log.Info("repo-service listening on " + addr)

	handler := requestLogger(log, mux)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Error("server terminated with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response.Success(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "repo-service",
	}, "Service is healthy")
}
