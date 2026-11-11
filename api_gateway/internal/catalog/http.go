// Package catalog exposes the UUID catalog through the public Gateway.
package catalog

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	data "innogen-backend/shared/catalog"
	"innogen-backend/shared/middleware"
	"innogen-backend/shared/response"
)

type Handler struct {
	repo *data.Repository
	pool *pgxpool.Pool
	log  *slog.Logger
}

func Register(mux *http.ServeMux, pool *pgxpool.Pool, secret string, log *slog.Logger) {
	h := &Handler{data.New(pool), pool, log}
	auth := middleware.Auth(secret)
	public := func(fn http.HandlerFunc) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "" {
				auth(h.account(fn)).ServeHTTP(w, r)
				return
			}
			fn(w, r)
		})
	}
	mux.Handle("GET /api/v1/languages", public(h.languages))
	mux.Handle("GET /api/v1/tags", public(h.tags))
	for path, fn := range map[string]http.HandlerFunc{
		"/packages": h.packages, "/packages/{slug}": h.packageDetail, "/packages/{slug}/levels": h.packageDetail,
		"/problems": h.problems, "/problems/{slug}": h.problem, "/problems/{slug}/languages": h.problem,
		"/problems/{slug}/samples": h.problem, "/problems/{slug}/template": h.problem,
	} {
		mux.Handle("GET /api/v1"+path, auth(h.account(fn)))
	}
}
func (h *Handler) account(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := middleware.GetUserID(r)
		if !ok {
			response.ErrorSimple(w, 401, "Unauthorized")
			return
		}
		var exists, active bool
		err := h.pool.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1),EXISTS(SELECT 1 FROM users WHERE id=$1 AND is_active)`, id).Scan(&exists, &active)
		if h.failed(w, r, err) {
			return
		}
		if !exists {
			response.ErrorSimple(w, 401, "Unauthorized")
			return
		}
		if !active {
			response.Error(w, 403, "Account is inactive", "ACCOUNT_INACTIVE")
			return
		}
		next.ServeHTTP(w, r)
	})
}
func (h *Handler) failed(w http.ResponseWriter, r *http.Request, err error) bool {
	if err == nil {
		return false
	}
	h.log.ErrorContext(r.Context(), "catalog request failed", slog.String("error", err.Error()))
	response.ErrorSimple(w, 500, "Internal server error")
	return true
}
func (h *Handler) languages(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.Languages(r.Context(), "")
	if !h.failed(w, r, err) {
		response.Success(w, 200, map[string]any{"items": items}, "Languages retrieved")
	}
}
func (h *Handler) tags(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.Tags(r.Context())
	if !h.failed(w, r, err) {
		response.Success(w, 200, map[string]any{"items": items}, "Tags retrieved")
	}
}
func pagination(r *http.Request) (int, int, bool) {
	page, size := 1, 20
	var err error
	if r.URL.Query().Has("page") {
		page, err = strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			return 0, 0, false
		}
	}
	if r.URL.Query().Has("size") {
		size, err = strconv.Atoi(r.URL.Query().Get("size"))
		if err != nil {
			return 0, 0, false
		}
	}
	return page, size, page > 0 && page <= 2147483647/100 && size > 0 && size <= 100
}
func (h *Handler) packages(w http.ResponseWriter, r *http.Request) {
	page, size, ok := pagination(r)
	if !ok {
		response.ErrorSimple(w, 400, "Invalid pagination")
		return
	}
	out, err := h.repo.Packages(r.Context(), page, size, r.URL.Query().Get("q"))
	if !h.failed(w, r, err) {
		response.Success(w, 200, out, "Packages retrieved")
	}
}
func (h *Handler) packageDetail(w http.ResponseWriter, r *http.Request) {
	pkg, err := h.repo.PackageBySlug(r.Context(), r.PathValue("slug"))
	if h.failed(w, r, err) {
		return
	}
	if pkg == nil {
		response.ErrorSimple(w, 404, "Package not found")
		return
	}
	uid, _ := middleware.GetUserID(r)
	levels, err := h.repo.Levels(r.Context(), pkg.ID, uid)
	if !h.failed(w, r, err) {
		response.Success(w, 200, map[string]any{"package": pkg, "levels": levels}, "Package retrieved")
	}
}
func (h *Handler) problems(w http.ResponseWriter, r *http.Request) {
	page, size, ok := pagination(r)
	q := r.URL.Query()
	difficulty := q.Get("difficulty")
	if !ok || (difficulty != "" && difficulty != "easy" && difficulty != "medium" && difficulty != "hard") {
		response.ErrorSimple(w, 400, "Invalid problem filter")
		return
	}
	f := data.Filter{Page: page, Size: size, Package: q.Get("package"), Difficulty: difficulty, Tag: q.Get("tag"), Query: q.Get("q")}
	if q.Has("solved") {
		v := q.Get("solved")
		if v != "true" && v != "false" {
			response.ErrorSimple(w, 400, "Invalid solved filter")
			return
		}
		b := v == "true"
		f.Solved = &b
	}
	uid, _ := middleware.GetUserID(r)
	out, err := h.repo.Problems(r.Context(), uid, f)
	if !h.failed(w, r, err) {
		response.Success(w, 200, out, "Problems retrieved")
	}
}
func (h *Handler) problem(w http.ResponseWriter, r *http.Request) {
	template := r.Pattern == "GET /api/v1/problems/{slug}/template"
	language := r.URL.Query().Get("languageId")
	if template && !middleware.ValidUUID(language) {
		response.ErrorSimple(w, 400, "Invalid language ID")
		return
	}
	uid, _ := middleware.GetUserID(r)
	detail, err := h.repo.ProblemBySlug(r.Context(), r.PathValue("slug"), uid)
	if h.failed(w, r, err) {
		return
	}
	if detail == nil {
		response.ErrorSimple(w, 404, "Problem not found")
		return
	}
	if !detail.Problem.Unlocked {
		response.Error(w, 403, "Complete previous levels first", "LEVEL_LOCKED")
		return
	}
	var out any = detail
	switch r.Pattern {
	case "GET /api/v1/problems/{slug}/languages":
		out = map[string]any{"items": detail.Languages}
	case "GET /api/v1/problems/{slug}/samples":
		out = map[string]any{"items": detail.Samples}
	case "GET /api/v1/problems/{slug}/template":
		t, e := h.repo.Template(r.Context(), detail.Problem.ID, language)
		if h.failed(w, r, e) {
			return
		}
		if t == nil {
			response.ErrorSimple(w, 404, "Template not found")
			return
		}
		out = t
	}
	response.Success(w, 200, out, "Problem retrieved")
}
