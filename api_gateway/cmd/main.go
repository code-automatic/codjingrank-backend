package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"innogen-backend/api_gateway/internal/admin"
	"innogen-backend/api_gateway/internal/catalog"
	"innogen-backend/api_gateway/internal/middleware"
	"innogen-backend/api_gateway/internal/proxy"
	"innogen-backend/api_gateway/internal/route"
	"innogen-backend/shared/config"
	"innogen-backend/shared/database"
	"innogen-backend/shared/logger"
	"innogen-backend/shared/response"
)

func main() {
	cfg := config.Load()
	log := logger.New("api-gateway")

	ctx := context.Background()

	pool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	// Create proxies for backend services
	authPublicProxy, err := proxy.NewProxy(cfg.AuthServiceURL, log)
	if err != nil {
		log.Error("failed to create auth public proxy", slog.String("error", err.Error()))
		os.Exit(1)
	}
	authProxy, err := proxy.NewAuthenticatedProxy(cfg.AuthServiceURL, log)
	if err != nil {
		log.Error("failed to create auth proxy", slog.String("error", err.Error()))
		os.Exit(1)
	}
	runProxy, err := proxy.NewAuthenticatedProxy(cfg.RunServiceURL, log)
	if err != nil {
		log.Error("failed to create run proxy", slog.String("error", err.Error()))
		os.Exit(1)
	}
	submissionProxy, err := proxy.NewAuthenticatedProxy(cfg.SubmissionServiceURL, log)
	if err != nil {
		log.Error("failed to create submission proxy", slog.String("error", err.Error()))
		os.Exit(1)
	}
	proxySet := route.ProxySet{
		AuthPublic: authPublicProxy,
		Auth:       authProxy,
		Run:        runProxy,
		Submission: submissionProxy,
	}

	mux := http.NewServeMux()

	// Swagger UI - Frontend
	mux.HandleFunc("GET /docs/fe", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html><html><head><title>RinnoGen API - Frontend</title><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head><body><div id="swagger-ui"></div><script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script><script>SwaggerUIBundle({ url: 'openapi_fe.yaml', dom_id: '#swagger-ui' });</script></body></html>`))
	})

	// Swagger UI - Full
	mux.HandleFunc("GET /docs/be", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html><html><head><title>RinnoGen API - Full</title><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head><body><div id="swagger-ui"></div><script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script><script>SwaggerUIBundle({ url: 'openapi.yaml', dom_id: '#swagger-ui' });</script></body></html>`))
	})

	// YAML specs
	mux.HandleFunc("GET /docs/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/openapi.yaml")
	})
	mux.HandleFunc("GET /docs/openapi_fe.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/openapi_fe.yaml")
	})

	// Health
	mux.HandleFunc("GET /health", healthHandler)

	catalog.Register(mux, pool, cfg.JWTSecret, log)

	route.RegisterRetiredCatalog(mux)

	// Proxy routes to backend services
	route.RegisterProxyRoutes(mux, &proxySet, log, cfg.JWTSecret)

	// Admin routes
	admin.RegisterV1(mux, pool, cfg.JWTSecret, log)

	addr := ":" + cfg.APIGatewayPort
	log.Info("api-gateway listening on " + addr)

	// Middleware chain (outermost first)
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRequests, cfg.RateLimitWindowSeconds)

	handler := middleware.Recover(log,
		middleware.RequestLogger(log,
			middleware.RequestID(
				middleware.CORS(cfg.CORSAllowedOrigins)(
					middleware.BodyLimit(cfg.MaxBodyBytes)(
						middleware.RateLimit(rateLimiter)(response.Router(mux)),
					),
				),
			),
		),
	)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Error("server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

// healthHandler responds with the service health status.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response.Success(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "api-gateway",
	}, "Service is healthy")
}
