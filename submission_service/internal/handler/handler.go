package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"innogen-backend/shared/middleware"
	"innogen-backend/shared/response"
	"innogen-backend/submission_service/internal/dto"
	"innogen-backend/submission_service/internal/repository"
	"innogen-backend/submission_service/internal/service"
)

// Handler holds dependencies for HTTP request handlers.
type Handler struct {
	svc *service.SubmissionService
	log *slog.Logger
}

// New creates a new Handler.
func New(svc *service.SubmissionService, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// Submit handles POST /submit.
func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.ErrorSimple(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req dto.SubmitRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if field, msg := req.Validate(); msg != "" {
		response.ErrorValidation(w, msg, field)
		return
	}

	sub, err := h.svc.CreateSubmission(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			response.ErrorSimple(w, http.StatusBadRequest, "Invalid input")
		case errors.Is(err, service.ErrProblemNotFound):
			response.ErrorSimple(w, http.StatusBadRequest, "Problem not found")
		case errors.Is(err, service.ErrLanguageNotFound):
			response.ErrorSimple(w, http.StatusBadRequest, "Language not found")
		default:
			if errors.Is(err, repository.ErrSpamCooldown) {
				response.ErrorSimple(w, http.StatusTooManyRequests, "Please wait 10 seconds before submitting again")
				return
			}
			h.log.Error("create submission failed", slog.String("error", err.Error()))
			response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	response.Success(w, http.StatusCreated, dto.SubmitResponse{Submission: sub}, "Submission created successfully")
}

// GetByID handles GET /submissions/{id}.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" || !isValidUUID(id) {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid submission ID")
		return
	}

	sub, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrSubmissionNotFound) {
			response.ErrorSimple(w, http.StatusNotFound, "Submission not found")
			return
		}
		h.log.Error("get submission failed", slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.Success(w, http.StatusOK, dto.GetSubmissionResponse{Submission: sub}, "Submission retrieved successfully")
}

// ListMySubmissions handles GET /me/submissions.
func (h *Handler) ListMySubmissions(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.ErrorSimple(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	items, err := h.svc.ListByUserID(r.Context(), userID)
	if err != nil {
		h.log.Error("list submissions failed", slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.Success(w, http.StatusOK, dto.ListSubmissionsResponse{Submissions: items}, "User submissions retrieved successfully")
}

// GetLatestForProblem handles GET /me/submissions/{problemId}/latest.
func (h *Handler) GetLatestForProblem(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.ErrorSimple(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	problemIDStr := r.PathValue("problemId")
	problemID, err := strconv.Atoi(problemIDStr)
	if err != nil || problemID <= 0 {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid problem ID")
		return
	}

	sub, err := h.svc.GetLatestByUserAndProblem(r.Context(), userID, problemID)
	if err != nil {
		if errors.Is(err, service.ErrNoSubmissionForProblem) {
			response.ErrorSimple(w, http.StatusNotFound, "No submission found for this problem")
			return
		}
		h.log.Error("get latest submission failed", slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.Success(w, http.StatusOK, dto.LatestSubmissionResponse{Submission: sub}, "Latest submission retrieved successfully")
}

// isValidUUID checks if the string is a valid UUID format.
func isValidUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
