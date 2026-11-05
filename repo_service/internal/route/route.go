package route

import (
	"net/http"

	"innogen-backend/repo_service/internal/handler"
	"innogen-backend/repo_service/internal/oauth"
	"innogen-backend/repo_service/internal/webhook"
	"innogen-backend/shared/middleware"
)

// Register registers all repo service routes on the given ServeMux.
func Register(mux *http.ServeMux, h *handler.Handler, wh *webhook.WebhookHandler, oh *oauth.OAuthHandler) {
	mux.Handle("GET /repositories", middleware.XUserID()(http.HandlerFunc(h.ListRepositories)))
	mux.Handle("GET /repositories/{id}/commits", middleware.XUserID()(http.HandlerFunc(h.ListCommits)))
	mux.HandleFunc("POST /internal/commits/accepted-submission", h.CommitAcceptedSubmission)
	mux.HandleFunc("POST /webhooks/github", wh.HandleWebhook)
	mux.Handle("GET /github/connection", middleware.XUserID()(http.HandlerFunc(h.GetGithubConnection)))
	mux.Handle("POST /github/installations/link", middleware.XUserID()(http.HandlerFunc(h.LinkGithubInstallation)))
	mux.Handle("POST /github/disconnect", middleware.XUserID()(http.HandlerFunc(h.DisconnectInstallation)))

	// GitHub OAuth routes
	mux.Handle("GET /github/oauth/start-url", middleware.XUserID()(http.HandlerFunc(oh.StartURL)))
	mux.HandleFunc("GET /github/oauth/callback", oh.Callback)
	mux.Handle("GET /github/account", middleware.XUserID()(http.HandlerFunc(oh.GetAccount)))
	mux.Handle("POST /github/oauth/disconnect", middleware.XUserID()(http.HandlerFunc(oh.Disconnect)))
}
