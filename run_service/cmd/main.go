package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"innogen-backend/run_service/internal/handler"
	"innogen-backend/run_service/internal/repository"
	"innogen-backend/run_service/internal/route"
	"innogen-backend/run_service/internal/service"
	"innogen-backend/shared/config"
	"innogen-backend/shared/database"
	"innogen-backend/shared/logger"
	"innogen-backend/shared/piston"
	"innogen-backend/shared/response"
)

func main() {
	cfg := config.Load()
	log := logger.New("run-service")

	ctx := context.Background()

	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	repo := repository.New(pool)
	pistonClient := piston.NewClient(cfg.PistonBaseURL)
	svc := service.New(repo, pistonClient)
	h := handler.New(svc, log)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	route.Register(mux, h)

	addr := ":" + cfg.RunServicePort
	log.Info("run-service listening on " + addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Error("server terminated with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// healthHandler responds with the service health status.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response.Success(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "run-service",
	}, "Service is healthy")
}
