package route

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"innogen-backend/api_gateway/internal/proxy"
)

func RegisterCommunityProxies(mux *http.ServeMux, log *slog.Logger, forumURL, aiURL string) error {
	for _, service := range []struct {
		address, fallback string
		patterns          []string
	}{
		{forumURL, "http://localhost:8084", []string{
			"GET /api/v1/forum/posts", "POST /api/v1/forum/posts",
			"GET /api/v1/forum/posts/{id}", "PATCH /api/v1/forum/posts/{id}", "DELETE /api/v1/forum/posts/{id}",
			"GET /api/v1/forum/posts/{id}/comments", "POST /api/v1/forum/posts/{id}/comments",
			"PATCH /api/v1/forum/comments/{id}", "DELETE /api/v1/forum/comments/{id}",
			"PUT /api/v1/forum/posts/{id}/like", "DELETE /api/v1/forum/posts/{id}/like",
			"GET /api/v1/admin/forum/posts", "GET /api/v1/admin/forum/posts/{id}", "PATCH /api/v1/admin/forum/posts/{id}",
			"GET /api/v1/admin/forum/posts/{id}/comments", "PATCH /api/v1/admin/forum/comments/{id}",
		}},
		{aiURL, "http://localhost:8085", []string{"GET /api/v1/submissions/{id}/ai-review"}},
	} {
		if service.address == "" {
			service.address = service.fallback
		}
		target, err := url.Parse(service.address)
		if err != nil || (target.Scheme != "http" && target.Scheme != "https") || target.Host == "" {
			return fmt.Errorf("invalid forum or AI review service URL")
		}
		upstream, err := proxy.NewProxy(service.address, log)
		if err != nil {
			return err
		}
		for _, pattern := range service.patterns {
			mux.Handle(pattern, stripUserHeaders(upstream))
		}
	}
	return nil
}
