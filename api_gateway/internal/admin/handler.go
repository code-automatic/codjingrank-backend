package admin

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgconn"
	"innogen-backend/shared/response"
)

// AdminHandler handles admin CRUD requests.
type AdminHandler struct {
	repo *AdminRepository
	log  *slog.Logger
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(repo *AdminRepository, log *slog.Logger) *AdminHandler {
	return &AdminHandler{repo: repo, log: log}
}

// handlePgError maps PostgreSQL errors to appropriate HTTP responses.
func (h *AdminHandler) handlePgError(w http.ResponseWriter, err error) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			response.ErrorSimple(w, http.StatusConflict, "A record with this value already exists")
			return
		case "23503": // foreign_key_violation
			response.ErrorSimple(w, http.StatusNotFound, "Referenced resource not found")
			return
		case "23502": // not_null_violation
			response.ErrorSimple(w, http.StatusBadRequest, "Required field cannot be empty")
			return
		case "23514": // check_violation
			response.ErrorSimple(w, http.StatusBadRequest, "Invalid value for field")
			return
		}
	}
	h.log.Error("database error", slog.String("error", err.Error()))
	response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
}

// =========================================================================
//  LANGUAGES
// =========================================================================

// CreateLanguage handles POST /admin/languages.
func (h *AdminHandler) CreateLanguage(w http.ResponseWriter, r *http.Request) {
	var req CreateLanguageRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		response.ErrorSimple(w, http.StatusBadRequest, "Name is required")
		return
	}
	if req.PistonAlias == "" {
		response.ErrorSimple(w, http.StatusBadRequest, "Piston alias is required")
		return
	}
	if req.PistonVersion == "" {
		response.ErrorSimple(w, http.StatusBadRequest, "Piston version is required")
		return
	}
	entity, err := h.repo.CreateLanguage(r.Context(), req)
	if err != nil {
		h.handlePgError(w, err)
		return
	}
	response.Success(w, http.StatusCreated, map[string]interface{}{"language": entity}, "Language created successfully")
}

// UpdateLanguage handles PUT /admin/languages/{id}.
func (h *AdminHandler) UpdateLanguage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	var req UpdateLanguageRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	entity, err := h.repo.UpdateLanguage(r.Context(), id, req)
	if err != nil {
		h.handlePgError(w, err)
		return
	}
	if entity == nil {
		response.ErrorSimple(w, http.StatusNotFound, "Not found")
		return
	}
	response.Success(w, http.StatusOK, map[string]interface{}{"language": entity}, "Language detail retrieved successfully")
}

// ListLanguages handles GET /admin/languages.
func (h *AdminHandler) ListLanguages(w http.ResponseWriter, r *http.Request) {
	entities, err := h.repo.FindAllLanguages(r.Context())
	if err != nil {
		h.log.Error("list languages failed", slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Success(w, http.StatusOK, map[string]interface{}{"languages": entities}, "Languages retrieved successfully")
}

// GetLanguage handles GET /admin/languages/{id}.
func (h *AdminHandler) GetLanguage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	entity, err := h.repo.FindLanguageByID(r.Context(), id)
	if err != nil {
		h.log.Error("get language failed", slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if entity == nil {
		response.ErrorSimple(w, http.StatusNotFound, "Not found")
		return
	}
	response.Success(w, http.StatusOK, map[string]interface{}{"language": entity}, "Language detail retrieved successfully")
}

// =========================================================================
//  PROBLEMS
// =========================================================================

// CreateProblem handles POST /admin/problems.
func (h *AdminHandler) CreateProblem(w http.ResponseWriter, r *http.Request) {
	var req CreateProblemRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Slug == "" {
		response.ErrorSimple(w, http.StatusBadRequest, "Slug is required")
		return
	}
	if req.Title == "" {
		response.ErrorSimple(w, http.StatusBadRequest, "Title is required")
		return
	}
	if req.Difficulty == "" {
		response.ErrorSimple(w, http.StatusBadRequest, "Difficulty is required")
		return
	}
	if req.Difficulty != "Easy" && req.Difficulty != "Medium" && req.Difficulty != "Hard" {
		response.ErrorSimple(w, http.StatusBadRequest, "Difficulty must be Easy, Medium, or Hard")
		return
	}
	if req.ProblemMD == "" {
		response.ErrorSimple(w, http.StatusBadRequest, "Problem markdown is required")
		return
	}
	entity, err := h.repo.CreateProblem(r.Context(), req)
	if err != nil {
		h.handlePgError(w, err)
		return
	}
	response.Success(w, http.StatusCreated, map[string]interface{}{"problem": entity}, "Problem created successfully")
}

// UpdateProblem handles PUT /admin/problems/{id}.
func (h *AdminHandler) UpdateProblem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	var req UpdateProblemRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Difficulty != nil {
		d := *req.Difficulty
		if d != "Easy" && d != "Medium" && d != "Hard" {
			response.ErrorSimple(w, http.StatusBadRequest, "Difficulty must be Easy, Medium, or Hard")
			return
		}
	}
	entity, err := h.repo.UpdateProblem(r.Context(), id, req)
	if err != nil {
		h.handlePgError(w, err)
		return
	}
	if entity == nil {
		response.ErrorSimple(w, http.StatusNotFound, "Not found")
		return
	}
	response.Success(w, http.StatusOK, map[string]interface{}{"problem": entity}, "Problem detail retrieved successfully")
}

// DeleteProblem handles DELETE /admin/problems/{id}.
func (h *AdminHandler) DeleteProblem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	deleted, err := h.repo.DeleteProblem(r.Context(), id)
	if err != nil {
		h.handlePgError(w, err)
		return
	}
	if !deleted {
		response.ErrorSimple(w, http.StatusNotFound, "Not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListProblems handles GET /admin/problems with optional ?page and ?limit.
func (h *AdminHandler) ListProblems(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	entities, total, err := h.repo.FindAllProblems(r.Context(), page, limit)
	if err != nil {
		h.log.Error("list problems failed", slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, ProblemListResponse{
		Problems: entities,
		Page:     page,
		Limit:    limit,
		Total:    total,
	})
}

// GetProblem handles GET /admin/problems/{id}.
func (h *AdminHandler) GetProblem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	entity, err := h.repo.FindProblemByID(r.Context(), id)
	if err != nil {
		h.log.Error("get problem failed", slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	if entity == nil {
		response.ErrorSimple(w, http.StatusNotFound, "Not found")
		return
	}
	response.Success(w, http.StatusOK, map[string]interface{}{"problem": entity}, "Problem detail retrieved successfully")
}

// =========================================================================
//  TEST CASES
// =========================================================================

// CreateTestCase handles POST /admin/problems/{problemId}/test-cases.
func (h *AdminHandler) CreateTestCase(w http.ResponseWriter, r *http.Request) {
	problemID, err := strconv.Atoi(r.PathValue("problemId"))
	if err != nil || problemID <= 0 {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid problem ID")
		return
	}
	var req CreateTestCaseRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Visibility == "" {
		req.Visibility = "hidden"
	}
	if req.Visibility != "sample" && req.Visibility != "hidden" {
		response.ErrorSimple(w, http.StatusBadRequest, "Visibility must be 'sample' or 'hidden'")
		return
	}
	if req.ExpectedOutput == "" {
		response.ErrorSimple(w, http.StatusBadRequest, "Expected output is required")
		return
	}
	entity, err := h.repo.CreateTestCase(r.Context(), problemID, req)
	if err != nil {
		h.handlePgError(w, err)
		return
	}
	response.Success(w, http.StatusCreated, map[string]interface{}{"testCase": entity}, "Test case created successfully")
}

// UpdateTestCase handles PUT /admin/test-cases/{id}.
func (h *AdminHandler) UpdateTestCase(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	var req UpdateTestCaseRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Visibility != nil && *req.Visibility != "sample" && *req.Visibility != "hidden" {
		response.ErrorSimple(w, http.StatusBadRequest, "Visibility must be 'sample' or 'hidden'")
		return
	}
	entity, err := h.repo.UpdateTestCase(r.Context(), id, req)
	if err != nil {
		h.handlePgError(w, err)
		return
	}
	if entity == nil {
		response.ErrorSimple(w, http.StatusNotFound, "Not found")
		return
	}
	response.Success(w, http.StatusOK, map[string]interface{}{"testCase": entity}, "Test case detail retrieved successfully")
}

// DeleteTestCase handles DELETE /admin/test-cases/{id}.
func (h *AdminHandler) DeleteTestCase(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	deleted, err := h.repo.DeleteTestCase(r.Context(), id)
	if err != nil {
		h.handlePgError(w, err)
		return
	}
	if !deleted {
		response.ErrorSimple(w, http.StatusNotFound, "Not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ListTestCases handles GET /admin/problems/{problemId}/test-cases.
func (h *AdminHandler) ListTestCases(w http.ResponseWriter, r *http.Request) {
	problemID, err := strconv.Atoi(r.PathValue("problemId"))
	if err != nil || problemID <= 0 {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid problem ID")
		return
	}
	entities, err := h.repo.FindTestCasesByProblemID(r.Context(), problemID)
	if err != nil {
		h.log.Error("list test cases failed", slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Success(w, http.StatusOK, map[string]interface{}{"testCases": entities}, "Test cases retrieved successfully")
}

// =========================================================================
//  TAGS
// =========================================================================

// CreateTag handles POST /admin/tags.
func (h *AdminHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	var req CreateTagRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Name == "" {
		response.ErrorSimple(w, http.StatusBadRequest, "Name is required")
		return
	}
	entity, err := h.repo.CreateTag(r.Context(), req)
	if err != nil {
		h.handlePgError(w, err)
		return
	}
	response.Success(w, http.StatusCreated, map[string]interface{}{"tag": entity}, "Problem tag linked successfully")
}

// ListTags handles GET /admin/tags.
func (h *AdminHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	entities, err := h.repo.FindAllTags(r.Context())
	if err != nil {
		h.log.Error("list tags failed", slog.String("error", err.Error()))
		response.ErrorSimple(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Success(w, http.StatusOK, map[string]interface{}{"tags": entities}, "Problem tags retrieved successfully")
}

// =========================================================================
//  PROBLEM-TAGS
// =========================================================================

// CreateProblemTag handles POST /admin/problems/{problemId}/tags.
func (h *AdminHandler) CreateProblemTag(w http.ResponseWriter, r *http.Request) {
	problemID, err := strconv.Atoi(r.PathValue("problemId"))
	if err != nil || problemID <= 0 {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid problem ID")
		return
	}
	var req CreateProblemTagRequest
	if err := response.DecodeJSON(r, &req); err != nil {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.TagID <= 0 {
		response.ErrorSimple(w, http.StatusBadRequest, "Tag ID is required")
		return
	}
	if err := h.repo.CreateProblemTag(r.Context(), problemID, req); err != nil {
		h.handlePgError(w, err)
		return
	}
	response.Success(w, http.StatusOK, map[string]interface{}{"status": "deleted"}, "Resource deleted successfully")
}

// DeleteProblemTag handles DELETE /admin/problems/{problemId}/tags/{tagId}.
func (h *AdminHandler) DeleteProblemTag(w http.ResponseWriter, r *http.Request) {
	problemID, err := strconv.Atoi(r.PathValue("problemId"))
	if err != nil || problemID <= 0 {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid problem ID")
		return
	}
	tagID, err := strconv.Atoi(r.PathValue("tagId"))
	if err != nil || tagID <= 0 {
		response.ErrorSimple(w, http.StatusBadRequest, "Invalid tag ID")
		return
	}
	deleted, err := h.repo.DeleteProblemTag(r.Context(), problemID, tagID)
	if err != nil {
		h.handlePgError(w, err)
		return
	}
	if !deleted {
		response.ErrorSimple(w, http.StatusNotFound, "Not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
