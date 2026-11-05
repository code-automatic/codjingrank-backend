package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"innogen-backend/repo_service/internal/dto"
	"innogen-backend/repo_service/service"
	"innogen-backend/shared/middleware"
	"innogen-backend/shared/response"
)

// Handler holds dependencies for HTTP request handlers.
type Handler struct {
	svc *service.RepoService
	log *slog.Logger
}

// New creates a new Handler.
func New(svc *service.RepoService, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// ListRepositories handles GET /repositories.
func (h *Handler) ListRepositories(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.ErrorSimple(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	repos, err := h.svc.ListRepositories(r.Context(), userID)
	if err != nil {
		h.log.Error("list repositories failed", slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, dto.ListRepositoriesResponse{Repositories: repos})
}

// CommitAcceptedSubmission handles POST /internal/commits/accepted-submission.
func (h *Handler) CommitAcceptedSubmission(w http.ResponseWriter, r *http.Request) {
	var req dto.CommitAcceptedSubmissionRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.SubmissionID == "" {
		response.ErrorSimple(w, http.StatusBadRequest, "Missing required field: submissionId")
		return
	}
	if req.UserID <= 0 {
		response.ErrorSimple(w, http.StatusBadRequest, "Missing required field: userId")
		return
	}
	if req.ProblemID <= 0 {
		response.ErrorSimple(w, http.StatusBadRequest, "Missing required field: problemId")
		return
	}
	if req.LanguageID <= 0 {
		response.ErrorSimple(w, http.StatusBadRequest, "Missing required field: languageId")
		return
	}
	if req.Code == "" {
		response.ErrorSimple(w, http.StatusBadRequest, "Missing required field: code")
		return
	}

	svcReq := service.CommitAcceptedSubmissionRequest{
		SubmissionID: req.SubmissionID,
		UserID:       req.UserID,
		ProblemID:    req.ProblemID,
		LanguageID:   req.LanguageID,
		Code:         req.Code,
	}

	resp, err := h.svc.CommitAcceptedSubmission(r.Context(), svcReq)
	if err != nil {
		if errors.Is(err, service.ErrGithubAccountNotFound) {
			response.ErrorSimple(w, http.StatusNotFound, "GitHub account not linked")
			return
		}
		h.log.Error("commit accepted submission failed", slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// ListCommits handles GET /repositories/{id}/commits.
func (h *Handler) ListCommits(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.ErrorSimple(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	idStr := r.PathValue("id")
	repoID, err := strconv.Atoi(idStr)
	if err != nil || repoID <= 0 {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid repository ID")
		return
	}

	commits, err := h.svc.ListCommits(r.Context(), userID, repoID)
	if err != nil {
		if errors.Is(err, service.ErrRepositoryNotFound) {
			response.ErrorSimple(w, http.StatusNotFound, "Repository not found")
			return
		}
		h.log.Error("list commits failed", slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, dto.ListCommitsResponse{Commits: commits})
}

// GetGithubConnection handles GET /github/connection
func (h *Handler) GetGithubConnection(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.ErrorSimple(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	resp, err := h.svc.GetGithubConnection(r.Context(), userID)
	if err != nil {
		h.log.Error("get github connection failed", slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Success(w, http.StatusOK, resp, "GitHub App connection status retrieved successfully")
}

// LinkGithubInstallation handles POST /github/installations/link
func (h *Handler) LinkGithubInstallation(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.ErrorSimple(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var req dto.LinkGithubInstallationRequest
	if err := response.DecodeJSON(r, &req); err != nil || req.InstallationID == "" {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	resp, err := h.svc.LinkGithubInstallation(r.Context(), userID, req.InstallationID)
	if err != nil {
		if errors.Is(err, service.ErrInstallationNotFound) {
			response.ErrorSimple(w, http.StatusNotFound, "Installation not yet registered. Wait for webhook or retry.")
			return
		}
		if errors.Is(err, service.ErrInstallationAlreadyLinked) {
			response.ErrorSimple(w, http.StatusConflict, "Installation already linked to another user")
			return
		}
		h.log.Error("link installation failed", slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Success(w, http.StatusOK, resp, "GitHub App installation linked successfully")
}

// DisconnectInstallation handles POST /github/disconnect
func (h *Handler) DisconnectInstallation(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.ErrorSimple(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if err := h.svc.DisconnectInstallation(r.Context(), userID); err != nil {
		h.log.Error("disconnect failed", slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Success(w, http.StatusOK, map[string]string{"status": "disconnected"}, "GitHub App connection disconnected successfully")
}
