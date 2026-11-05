package webhook

import (
	"io"
	"log/slog"
	"net/http"
	"strings"

	"innogen-backend/shared/response"
)

type WebhookHandler struct {
	svc    *WebhookService
	secret []byte
	log    *slog.Logger
}

func NewWebhookHandler(svc *WebhookService, secret string, log *slog.Logger) *WebhookHandler {
	return &WebhookHandler{
		svc:    svc,
		secret: []byte(secret),
		log:    log,
	}
}

func (h *WebhookHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	deliveryID := r.Header.Get("X-GitHub-Delivery")
	eventType := r.Header.Get("X-GitHub-Event")
	signature := r.Header.Get("X-Hub-Signature-256")

	log := h.log.With(
		slog.String("deliveryId", deliveryID),
		slog.String("event", eventType),
	)

	if eventType == "" {
		response.ErrorSimple(w, http.StatusBadRequest, "Missing X-GitHub-Event header")
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20)) // 1MB limit
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			response.ErrorSimple(w, http.StatusRequestEntityTooLarge, "Request body too large")
		} else {
			response.ErrorSimple(w, http.StatusBadRequest, "Failed to read request body")
		}
		return
	}

	if !VerifySignature(body, signature, h.secret) {
		response.ErrorSimple(w, http.StatusUnauthorized, "Invalid webhook signature")
		return
	}

	if err := h.svc.ProcessEvent(r.Context(), eventType, body); err != nil {
		log.Error("webhook processing failed", slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.Success(w, http.StatusOK, map[string]string{"status": "ok"}, "Success")
}
