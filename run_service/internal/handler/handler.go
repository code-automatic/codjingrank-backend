package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"innogen-backend/run_service/internal/dto"
	"innogen-backend/run_service/internal/service"
	"innogen-backend/shared/response"
)

// Handler holds dependencies for HTTP request handlers.
type Handler struct {
	svc *service.RunService
	log *slog.Logger
}

// New creates a new Handler with the given dependencies.
func New(svc *service.RunService, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// Run handles POST /run.
func (h *Handler) Run(w http.ResponseWriter, r *http.Request) {
	var req dto.RunRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if field, msg := req.Validate(); msg != "" {
		response.ErrorValidation(w, msg, field)
		return
	}

	resp, err := h.svc.Run(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			response.ErrorSimple(w, http.StatusBadRequest, "Invalid input")
		case errors.Is(err, service.ErrProblemNotFound):
			response.ErrorSimple(w, http.StatusNotFound, "Problem not found")
		case errors.Is(err, service.ErrLanguageNotFound):
			response.ErrorSimple(w, http.StatusNotFound, "Language not found")
		default:
			h.log.Error("run failed", slog.String("error", err.Error()))
			response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	response.Success(w, http.StatusOK, resp, "Code executed successfully")
}
