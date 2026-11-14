package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"innogen-backend/run_service/internal/dto"
	"innogen-backend/run_service/internal/service"
	"innogen-backend/shared/catalog"
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
		case errors.Is(err, service.ErrExecutionTimeout):
			response.ErrorSimple(w, 504, "Code execution service timed out")
		case errors.Is(err, service.ErrExecutionUnavailable):
			response.ErrorSimple(w, 502, "Code execution service unavailable")
		case errors.Is(err, catalog.ErrProblemNotFound):
			response.ErrorSimple(w, 404, "Problem not found")
		case errors.Is(err, catalog.ErrLevelLocked):
			response.Error(w, 403, "Complete previous levels first", "LEVEL_LOCKED")
		case errors.Is(err, catalog.ErrAccountInactive):
			response.Error(w, 403, "Account is inactive", "ACCOUNT_INACTIVE")
		case errors.Is(err, catalog.ErrLanguageUnsupported):
			response.Error(w, 422, "Unsupported language", "LANGUAGE_NOT_SUPPORTED")
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
