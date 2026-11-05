package oauth

import (
	"log/slog"
	"net/http"

	"innogen-backend/shared/middleware"
	"innogen-backend/shared/response"
)

type OAuthHandler struct {
	svc *OAuthService
	log *slog.Logger
}

func NewOAuthHandler(svc *OAuthService, log *slog.Logger) *OAuthHandler {
	return &OAuthHandler{svc: svc, log: log}
}

func (h *OAuthHandler) StartURL(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.ErrorSimple(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	resp, err := h.svc.GetStartURL(userID)
	if err != nil {
		h.log.Error("oauth start url failed", slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Success(w, http.StatusOK, resp, "GitHub OAuth URL generated successfully")
}

func (h *OAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" || state == "" {
		http.Redirect(w, r, h.svc.cfg.GitHubOAuthFrontendRedirectURL+"?oauth=error&message=missing_params", http.StatusFound)
		return
	}
	redirectURL, err := h.svc.HandleCallback(r.Context(), code, state)
	if err != nil {
		h.log.Error("oauth callback failed", slog.String("error", err.Error()))
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h *OAuthHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.ErrorSimple(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	resp, err := h.svc.GetAccount(r.Context(), userID)
	if err != nil {
		h.log.Error("get oauth account failed", slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Success(w, http.StatusOK, resp, "GitHub account status retrieved successfully")
}

func (h *OAuthHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.ErrorSimple(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if err := h.svc.Disconnect(r.Context(), userID); err != nil {
		h.log.Error("oauth disconnect failed", slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Success(w, http.StatusOK, map[string]string{"status": "disconnected"}, "GitHub OAuth account disconnected successfully")
}
