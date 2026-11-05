package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"innogen-backend/shared/config"
	"innogen-backend/shared/database"
	"innogen-backend/shared/logger"
	"innogen-backend/shared/response"
	"innogen-backend/submission_service/internal/handler"
	"innogen-backend/submission_service/internal/queue"
	"innogen-backend/submission_service/internal/repository"
	"innogen-backend/submission_service/internal/route"
	"innogen-backend/submission_service/internal/service"
)

func main() {
	cfg := config.Load()
	log := logger.New("submission-service")

	ctx := context.Background()

	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	q, err := queue.New(cfg.RedisAddr)
	if err != nil {
		log.Error("failed to connect to redis", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer q.Close()

	repo := repository.New(pool)
	svc := service.New(repo, q, log)
	h := handler.New(svc, log)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	route.Register(mux, h)

	addr := ":" + cfg.SubmissionServicePort
	log.Info("submission-service listening on " + addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Error("server terminated with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response.Success(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "submission-service",
	}, "Service is healthy")
}
