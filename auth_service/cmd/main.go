package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"innogen-backend/auth_service/internal/handler"
	"innogen-backend/auth_service/internal/jwt"
	"innogen-backend/auth_service/internal/repository"
	"innogen-backend/auth_service/internal/route"
	"innogen-backend/auth_service/internal/service"
	"innogen-backend/shared/config"
	"innogen-backend/shared/database"
	"innogen-backend/shared/logger"
	"innogen-backend/shared/response"
)

func main() {
	cfg := config.Load()
	log := logger.New("auth-service")

	ctx := context.Background()

	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.New(pool)
	jwtSvc := jwt.NewService(cfg.JWTSecret)
	svc := service.New(repo, jwtSvc, cfg)
	h := handler.New(svc, log)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	route.Register(mux, h)

	addr := ":" + cfg.AuthServicePort
	log.Info("auth-service listening on " + addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Error("server terminated with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// healthHandler responds with the service health status.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response.Success(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "auth-service",
	}, "Service is healthy")
}
