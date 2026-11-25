package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"innogen-backend/ai_service/internal/review"
	"innogen-backend/shared/config"
	"innogen-backend/shared/database"
	"innogen-backend/shared/response"
)

func main() {
	cfg := config.Load()
	pool, err := database.NewPostgresPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("database connection failed")
		os.Exit(1)
	}
	defer pool.Close()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, http.StatusOK, map[string]string{"status": "ok", "service": "ai-service"}, "Service is healthy")
	})
	review.Register(mux, pool, cfg.JWTSecret, os.Getenv("AI_INTERNAL_SECRET"))
	url, model, cutoff := os.Getenv("AI_PROVIDER_URL"), os.Getenv("AI_PROVIDER_MODEL"), os.Getenv("AI_ENABLED_SINCE")
	if url != "" && model != "" && cutoff != "" {
		since, err := time.Parse(time.RFC3339, cutoff)
		if err != nil {
			slog.Error("invalid AI_ENABLED_SINCE")
			os.Exit(1)
		}
		go review.Run(context.Background(), pool, since, url, os.Getenv("AI_PROVIDER_TOKEN"), model, slog.Default())
	}
	port := os.Getenv("AI_SERVICE_PORT")
	if port == "" {
		port = "8085"
	}
	if err := http.ListenAndServe(":"+port, response.Router(mux)); err != nil {
		slog.Error("AI service failed")
		os.Exit(1)
	}
}
