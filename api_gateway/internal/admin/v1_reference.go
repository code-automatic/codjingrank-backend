package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"innogen-backend/shared/middleware"
	"innogen-backend/shared/response"
)

func (h *v1Handler) registerReference(mux *http.ServeMux, secret string) {
	for _, route := range []string{"GET /tags", "POST /tags", "PATCH /tags/{id}", "DELETE /tags/{id}", "GET /languages", "POST /languages", "GET /languages/{id}", "PATCH /languages/{id}"} {
		method, path, _ := strings.Cut(route, " ")
		mux.Handle(method+" /api/v1/admin"+path, middleware.Auth(secret)(h.guard(http.HandlerFunc(h.reference))))
	}
}

const languageJSON = `jsonb_build_object('id',id,'slug',slug,'name',name,'runtimeVersion',piston_version,'pistonVersion',piston_version,'pistonAlias',piston_alias,'fileExtension',file_extension,'defaultFileName',default_file_name,'isActive',is_active)`
const tagJSON = `jsonb_build_object('id',id,'slug',slug,'name',name)`

func (h *v1Handler) reference(w http.ResponseWriter, r *http.Request) {
	language := strings.Contains(r.URL.Path, "/languages")
	table, key, projection := "tags", "tag", tagJSON
	columns := map[string]string{"slug": "slug", "name": "name"}
	if language {
		table, key, projection = "languages", "language", languageJSON
		for k, v := range map[string]string{"pistonAlias": "piston_alias", "pistonVersion": "piston_version", "fileExtension": "file_extension", "defaultFileName": "default_file_name", "isActive": "is_active"} {
			columns[k] = v
		}
	}
	id := r.PathValue("id")
	if r.Method == "GET" {
		if id != "" {
			h.object(w, r, key, `SELECT `+projection+` FROM `+table+` WHERE id=$1`, id)
			return
		}
		page, size, ok := adminPage(r)
		if !ok {
			response.ErrorSimple(w, 400, "Invalid pagination")
			return
		}
		var total int
		if h.fail(w, h.pool.QueryRow(r.Context(), `SELECT count(*) FROM `+table).Scan(&total)) {
			return
		}
		var items json.RawMessage
		err := h.pool.QueryRow(r.Context(), `SELECT COALESCE(jsonb_agg(v ORDER BY s,i),'[]') FROM(SELECT `+projection+` v,slug s,id i FROM `+table+` ORDER BY slug,id LIMIT $1 OFFSET $2) x`, size, (page-1)*size).Scan(&items)
		if !h.fail(w, err) {
			response.Success(w, 200, map[string]any{"items": items, "page": page, "size": size, "total": total}, "Resources retrieved")
		}
		return
	}
	if r.Method == "DELETE" {
		res, err := h.pool.Exec(r.Context(), `DELETE FROM tags WHERE id=$1 AND NOT EXISTS(SELECT 1 FROM problem_tags WHERE tag_id=$1)`, id)
		if h.fail(w, err) {
			return
		}
		if res.RowsAffected() == 0 {
			var exists bool
			if h.fail(w, h.pool.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM tags WHERE id=$1)`, id).Scan(&exists)) {
				return
			}
			if exists {
				response.ErrorSimple(w, 409, "Tag is in use")
			} else {
				response.ErrorSimple(w, 404, "Tag not found")
			}
			return
		}
		response.NoContent(w)
		return
	}
	body, sets, args, ok := patchFields(w, r, columns)
	if !ok {
		return
	}
	if raw := body["defaultFileName"]; raw != nil {
		var f string
		json.Unmarshal(raw, &f)
		if !filePattern.MatchString(f) || f == "." || f == ".." {
			response.ErrorSimple(w, 400, "Invalid file name")
			return
		}
	}
	if raw := body["name"]; raw != nil && !language {
		var n string
		json.Unmarshal(raw, &n)
		if !validText(n, 80) {
			response.ErrorSimple(w, 400, "Invalid tag name")
			return
		}
	}
	tx, err := h.pool.Begin(r.Context())
	if h.fail(w, err) {
		return
	}
	defer tx.Rollback(r.Context())
	if language && r.Method == "PATCH" {
		// Runtime selectors are frozen if referenced by any published problem, including archived ones.
		var exists bool
		err = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM languages WHERE id=$1 FOR UPDATE)`, id).Scan(&exists)
		if h.fail(w, err) {
			return
		}
		if !exists {
			response.ErrorSimple(w, 404, "Language not found")
			return
		}
		for _, field := range []string{"pistonAlias", "pistonVersion", "fileExtension", "defaultFileName"} {
			if body[field] != nil {
				var used bool
				err = tx.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM problem_languages pl JOIN problems p ON p.id=pl.problem_id WHERE pl.language_id=$1 AND p.published_at IS NOT NULL)`, id).Scan(&used)
				if h.fail(w, err) {
					return
				}
				if used {
					response.ErrorSimple(w, 409, "Runtime is used by published problems")
					return
				}
				break
			}
		}
	}
	var raw json.RawMessage
	if r.Method == "PATCH" {
		args = append(args, id)
		err = tx.QueryRow(r.Context(), `UPDATE `+table+` SET `+strings.Join(sets, ",")+fmt.Sprintf(" WHERE id=$%d RETURNING ", len(args))+projection, args...).Scan(&raw)
	} else {
		for key := range columns {
			if body[key] == nil {
				response.ErrorSimple(w, 400, "Missing required field")
				return
			}
		}
		cols := []string{}
		vals := []string{}
		params := []any{}
		for k, v := range body {
			var value any
			json.Unmarshal(v, &value)
			params = append(params, value)
			cols = append(cols, columns[k])
			vals = append(vals, fmt.Sprintf("$%d", len(params)))
		}
		err = tx.QueryRow(r.Context(), `INSERT INTO `+table+` (`+strings.Join(cols, ",")+`) VALUES(`+strings.Join(vals, ",")+`) RETURNING `+projection, params...).Scan(&raw)
	}
	if h.fail(w, err) || h.fail(w, tx.Commit(r.Context())) {
		return
	}
	status := 200
	if r.Method == "POST" {
		status = 201
	}
	response.Success(w, status, map[string]any{key: raw}, "Resource saved")
}
