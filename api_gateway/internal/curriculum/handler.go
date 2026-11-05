package curriculum

import (
	"log/slog"
	"net/http"
	"strconv"

	"innogen-backend/shared/response"
)

// Handler serves curriculum-related HTTP endpoints.
type Handler struct {
	repo *CurriculumRepository
	log  *slog.Logger
}

// NewHandler creates a new Handler.
func NewHandler(repo *CurriculumRepository, log *slog.Logger) *Handler {
	return &Handler{repo: repo, log: log}
}

// ListSubjects handles GET /subjects.
func (h *Handler) ListSubjects(w http.ResponseWriter, r *http.Request) {
	subjects, err := h.repo.FindAllSubjects(r.Context())
	if err != nil {
		h.log.Error("failed to list subjects", slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	dtos := make([]SubjectDTO, len(subjects))
	for i, s := range subjects {
		dtos[i] = subjectToDTO(s)
	}

	response.Success(w, http.StatusOK, map[string]any{"subjects": dtos}, "Subjects retrieved successfully")
}

// GetSubject handles GET /subjects/{slug}.
func (h *Handler) GetSubject(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		response.ErrorSimple(w, http.StatusBadRequest, "Missing subject slug")
		return
	}

	subject, err := h.repo.FindSubjectBySlug(r.Context(), slug)
	if err != nil {
		h.log.Error("failed to get subject", slog.String("slug", slug), slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if subject == nil {
		response.ErrorSimple(w, http.StatusNotFound, "Subject not found")
		return
	}

	response.Success(w, http.StatusOK, map[string]any{"subject": subjectToDTO(*subject)}, "Subject detail retrieved successfully")
}

// ListSessions handles GET /subjects/{slug}/sessions.
func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		response.ErrorSimple(w, http.StatusBadRequest, "Missing subject slug")
		return
	}

	subject, err := h.repo.FindSubjectBySlug(r.Context(), slug)
	if err != nil {
		h.log.Error("failed to find subject for sessions", slog.String("slug", slug), slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if subject == nil {
		response.ErrorSimple(w, http.StatusNotFound, "Subject not found")
		return
	}

	sessions, err := h.repo.FindSessionsBySubjectID(r.Context(), subject.ID)
	if err != nil {
		h.log.Error("failed to list sessions", slog.Int("subjectID", subject.ID), slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	dtos := make([]SessionDTO, len(sessions))
	for i, s := range sessions {
		dtos[i] = sessionToDTO(s)
	}

	response.Success(w, http.StatusOK, map[string]any{"sessions": dtos}, "Subject sessions retrieved successfully")
}

// ListLessons handles GET /sessions/{id}/lessons.
func (h *Handler) ListLessons(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid session id")
		return
	}

	lessons, err := h.repo.FindLessonsBySessionID(r.Context(), id)
	if err != nil {
		h.log.Error("failed to list lessons", slog.Int("sessionID", id), slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	dtos := make([]LessonDTO, len(lessons))
	for i, l := range lessons {
		dtos[i] = lessonToDTO(l)
	}

	response.Success(w, http.StatusOK, map[string]any{"lessons": dtos}, "Session lessons retrieved successfully")
}

// GetLesson handles GET /lessons/{id}.
func (h *Handler) GetLesson(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid lesson id")
		return
	}

	lesson, err := h.repo.FindLessonByID(r.Context(), id)
	if err != nil {
		h.log.Error("failed to get lesson", slog.Int("lessonID", id), slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if lesson == nil {
		response.ErrorSimple(w, http.StatusNotFound, "Lesson not found")
		return
	}

	response.Success(w, http.StatusOK, map[string]any{"lesson": lessonToDTO(*lesson)}, "Lesson detail retrieved successfully")
}

// ListLessonProblems handles GET /lessons/{id}/problems.
func (h *Handler) ListLessonProblems(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid lesson id")
		return
	}

	items, err := h.repo.FindProblemsByLessonID(r.Context(), id)
	if err != nil {
		h.log.Error("failed to list lesson problems", slog.Int("lessonID", id), slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.Success(w, http.StatusOK, map[string]any{"problems": items}, "Lesson problems retrieved successfully")
}
