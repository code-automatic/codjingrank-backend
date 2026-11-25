package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"innogen-backend/forum_service/internal/forum"
	"innogen-backend/shared/config"
	"innogen-backend/shared/database"
)

func main() {
	cfg := config.Load()
	if cfg.JWTSecret == "" {
		slog.Error("JWT_SECRET is required")
		os.Exit(1)
	}
	pool, err := database.NewPostgresPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	port := os.Getenv("FORUM_SERVICE_PORT")
	if port == "" {
		port = "8084"
	}
	server := &http.Server{Addr: ":" + port, Handler: forum.New(pool, cfg.JWTSecret)}
	if err := server.ListenAndServe(); err != nil {
		slog.Error("forum server failed", "error", err)
		os.Exit(1)
	}
}
