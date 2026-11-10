package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"innogen-backend/shared/catalog"
	"innogen-backend/shared/middleware"
	"innogen-backend/shared/response"
)

type v1Handler struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}
type packageWrite struct {
	Slug         string  `json:"slug"`
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	DisplayOrder int     `json:"displayOrder"`
}
type problemWrite struct {
	LevelID   string  `json:"levelId"`
	Slug      string  `json:"slug"`
	Title     string  `json:"title"`
	Statement string  `json:"statementMd"`
	Mode      string  `json:"executionMode"`
	Function  *string `json:"functionName"`
	Time      int     `json:"timeLimitMs"`
	Memory    int     `json:"memoryLimitMb"`
	Output    int     `json:"outputLimitBytes"`
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
var filePattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

func RegisterV1(mux *http.ServeMux, pool *pgxpool.Pool, secret string, log *slog.Logger) {
	h := &v1Handler{pool, log}
	mux.Handle("POST /api/v1/admin/problems/{id}/clone", middleware.Auth(secret)(h.guard(http.HandlerFunc(h.clone))))
	h.registerReference(mux, secret)
	routes := map[string]http.HandlerFunc{
		"GET /packages": h.packages, "POST /packages": h.createPackage,
		"GET /packages/{id}": h.packageItem, "PATCH /packages/{id}": h.packageItem, "DELETE /packages/{id}": h.packageItem,
		"POST /packages/{id}/publish": h.packageItem, "POST /packages/{id}/archive": h.packageItem,
		"GET /packages/{id}/levels": h.levels, "PATCH /levels/{id}": h.levels,
		"GET /problems": h.problems, "POST /problems": h.createProblem,
		"GET /problems/{id}": h.problemItem, "PATCH /problems/{id}": h.problemItem, "DELETE /problems/{id}": h.problemItem,
		"POST /problems/{id}/publish": h.problemItem, "POST /problems/{id}/archive": h.problemItem,
	}
	for route, fn := range routes {
		method, path, _ := strings.Cut(route, " ")
		mux.Handle(method+" /api/v1/admin"+path, middleware.Auth(secret)(h.guard(fn)))
	}
	for _, kind := range []string{"samples", "test-cases", "tags", "languages"} {
		mux.Handle("GET /api/v1/admin/problems/{id}/"+kind, middleware.Auth(secret)(h.guard(http.HandlerFunc(h.children))))
		if kind == "tags" || kind == "languages" {
			mux.Handle("PUT /api/v1/admin/problems/{id}/"+kind, middleware.Auth(secret)(h.guard(http.HandlerFunc(h.children))))
			continue
		}
		mux.Handle("POST /api/v1/admin/problems/{id}/"+kind, middleware.Auth(secret)(h.guard(http.HandlerFunc(h.children))))
		for _, method := range []string{"PATCH", "DELETE"} {
			mux.Handle(method+" /api/v1/admin/problems/{id}/"+kind+"/{itemId}", middleware.Auth(secret)(h.guard(http.HandlerFunc(h.children))))
		}
	}
	for _, method := range []string{"GET", "PUT", "DELETE"} {
		mux.Handle(method+" /api/v1/admin/problems/{id}/templates/{languageId}", middleware.Auth(secret)(h.guard(http.HandlerFunc(h.template))))
	}
}
func (h *v1Handler) guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid, _ := middleware.GetUserID(r)
		var role string
		var active bool
		err := h.pool.QueryRow(r.Context(), `SELECT role,is_active FROM users WHERE id=$1`, uid).Scan(&role, &active)
		if errors.Is(err, pgx.ErrNoRows) {
			response.ErrorSimple(w, 401, "Unauthorized")
			return
		}
		if h.fail(w, err) {
			return
		}
		if !active || role != "admin" {
			response.ErrorSimple(w, 403, "Admin access required")
			return
		}
		for _, key := range []string{"id", "itemId", "languageId"} {
			if v := r.PathValue(key); v != "" && !middleware.ValidUUID(v) {
				response.ErrorSimple(w, 400, "Invalid resource ID")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
func (h *v1Handler) fail(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, pgx.ErrNoRows) {
		response.ErrorSimple(w, 404, "Resource not found")
		return true
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		switch pg.Code {
		case "23505":
			response.ErrorSimple(w, 409, "Resource already exists")
			return true
		case "23503":
			response.ErrorSimple(w, 409, "Resource is referenced or missing")
			return true
		case "P0001", "23514", "23502", "22P02":
			response.Error(w, 422, "Invalid catalog state or field", "VALIDATION_ERROR")
			return true
		}
	}
	h.log.Error("catalog administration failed", slog.String("error", err.Error()))
	response.ErrorSimple(w, 500, "Internal server error")
	return true
}
func decode(w http.ResponseWriter, r *http.Request, out any) bool {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		response.ErrorSimple(w, 400, "Invalid request body")
		return false
	}
	if d.Decode(new(any)) != io.EOF {
		response.ErrorSimple(w, 400, "Expected one JSON object")
		return false
	}
	return true
}
func validText(v string, max int) bool {
	return strings.TrimSpace(v) != "" && (max == 0 || utf8.RuneCountInString(v) <= max)
}

const packageJSON = `jsonb_build_object('id',id,'slug',slug,'name',name,'description',description,'displayOrder',display_order,'status',status,'publishedAt',published_at,'createdAt',created_at,'updatedAt',updated_at)`
const problemJSON = `jsonb_build_object('id',id,'levelId',level_id,'slug',slug,'title',title,'statementMd',statement_md,'executionMode',execution_mode,'functionName',function_name,'timeLimitMs',time_limit_ms,'memoryLimitMb',memory_limit_mb,'outputLimitBytes',output_limit_bytes,'status',status,'publishedAt',published_at,'createdAt',created_at,'updatedAt',updated_at)`
const levelJSON = `jsonb_build_object('id',id,'packageId',package_id,'levelNo',level_no,'slug',slug,'name',name,'difficulty',difficulty,'displayOrder',display_order)`

func (h *v1Handler) object(w http.ResponseWriter, r *http.Request, key, query string, args ...any) {
	var raw json.RawMessage
	if key == "package" && r.Method == "GET" {
		query = `SELECT jsonb_build_object('package',` + packageJSON + `,'levels',(SELECT COALESCE(jsonb_agg(v ORDER BY n),'[]') FROM(SELECT ` + levelJSON + ` v,level_no n FROM levels WHERE package_id=packages.id) x)) FROM packages WHERE id=$1`
		if h.fail(w, h.pool.QueryRow(r.Context(), query, args...).Scan(&raw)) {
			return
		}
		response.Success(w, 200, raw, "Package retrieved")
		return
	}
	if h.fail(w, h.pool.QueryRow(r.Context(), query, args...).Scan(&raw)) {
		return
	}
	response.Success(w, 200, map[string]any{key: raw}, "Resource retrieved")
}
func (h *v1Handler) createPackage(w http.ResponseWriter, r *http.Request) {
	var v packageWrite
	if !decode(w, r, &v) {
		return
	}
	if !slugPattern.MatchString(v.Slug) || !validText(v.Name, 120) || v.DisplayOrder < 0 || (v.Description != nil && utf8.RuneCountInString(*v.Description) > 10000) {
		response.ErrorSimple(w, 400, "Invalid package fields")
		return
	}
	p, err := catalog.New(h.pool).CreatePackage(r.Context(), v.Slug, v.Name, v.Description, v.DisplayOrder)
	if h.fail(w, err) {
		return
	}
	var raw json.RawMessage
	if h.fail(w, h.pool.QueryRow(r.Context(), `SELECT `+packageJSON+` FROM packages WHERE id=$1`, p.ID).Scan(&raw)) {
		return
	}
	var levels json.RawMessage
	err = h.pool.QueryRow(r.Context(), `SELECT COALESCE(jsonb_agg(v ORDER BY n),'[]') FROM(SELECT `+levelJSON+` v,level_no n FROM levels WHERE package_id=$1) x`, p.ID).Scan(&levels)
	if !h.fail(w, err) {
		response.Success(w, 201, map[string]any{"package": raw, "levels": levels}, "Package created")
	}
}
func adminPage(r *http.Request) (int, int, bool) {
	page, size := 1, 20
	var e error
	if r.URL.Query().Has("page") {
		page, e = strconv.Atoi(r.URL.Query().Get("page"))
		if e != nil {
			return 0, 0, false
		}
	}
	if r.URL.Query().Has("size") {
		size, e = strconv.Atoi(r.URL.Query().Get("size"))
		if e != nil {
			return 0, 0, false
		}
	}
	return page, size, page > 0 && page <= 2147483647/100 && size > 0 && size <= 100
}
func (h *v1Handler) list(w http.ResponseWriter, r *http.Request, table, projection string) {
	page, size, ok := adminPage(r)
	status := r.URL.Query().Get("status")
	if !ok || (status != "" && status != "draft" && status != "published" && status != "archived") {
		response.ErrorSimple(w, 400, "Invalid catalog filter")
		return
	}
	q := r.URL.Query().Get("q")
	search := "name"
	if table == "problems" {
		search = "title"
	}
	where := ` WHERE ($1='' OR status::text=$1) AND ($2='' OR ` + search + ` ILIKE '%'||$2||'%' OR slug ILIKE '%'||$2||'%')`
	args := []any{status, q}
	if table == "problems" {
		for _, key := range []string{"levelId", "packageId"} {
			value := r.URL.Query().Get(key)
			if value != "" && !middleware.ValidUUID(value) {
				response.ErrorSimple(w, 400, "Invalid catalog ID filter")
				return
			}
			args = append(args, value)
		}
		where += ` AND ($3='' OR level_id=NULLIF($3,'')::uuid) AND ($4='' OR level_id IN(SELECT id FROM levels WHERE package_id=NULLIF($4,'')::uuid))`
	}
	var total int
	if h.fail(w, h.pool.QueryRow(r.Context(), `SELECT count(*) FROM `+table+where, args...).Scan(&total)) {
		return
	}
	args = append(args, size, (page-1)*size)
	rows, err := h.pool.Query(r.Context(), `SELECT `+projection+` FROM `+table+where+fmt.Sprintf(` ORDER BY created_at,id LIMIT $%d OFFSET $%d`, len(args)-1, len(args)), args...)
	if h.fail(w, err) {
		return
	}
	defer rows.Close()
	items := []json.RawMessage{}
	for rows.Next() {
		var raw json.RawMessage
		if h.fail(w, rows.Scan(&raw)) {
			return
		}
		items = append(items, raw)
	}
	if !h.fail(w, rows.Err()) {
		response.Success(w, 200, map[string]any{"items": items, "page": page, "size": size, "total": total}, "Resources retrieved")
	}
}
func (h *v1Handler) packages(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, "packages", packageJSON)
}
func (h *v1Handler) problems(w http.ResponseWriter, r *http.Request) {
	h.list(w, r, "problems", problemJSON)
}

// patchFields is a fixed server-owned column whitelist, never client SQL identifiers.
func patchFields(w http.ResponseWriter, r *http.Request, columns map[string]string) (map[string]json.RawMessage, []string, []any, bool) {
	var body map[string]json.RawMessage
	if !decode(w, r, &body) {
		return nil, nil, nil, false
	}
	if len(body) == 0 {
		response.ErrorSimple(w, 400, "Empty patch")
		return nil, nil, nil, false
	}
	sets := []string{}
	args := []any{}
	for key, raw := range body {
		column, ok := columns[key]
		if !ok {
			response.ErrorSimple(w, 400, "Unknown field")
			return nil, nil, nil, false
		}
		var v any
		if err := json.Unmarshal(raw, &v); err != nil {
			response.ErrorSimple(w, 400, "Invalid field")
			return nil, nil, nil, false
		}
		if v == nil && key != "description" && key != "functionName" && key != "inputData" {
			response.ErrorSimple(w, 400, "Field cannot be null")
			return nil, nil, nil, false
		}
		switch key {
		case "displayOrder", "timeLimitMs", "memoryLimitMb", "outputLimitBytes":
			n, ok := v.(float64)
			minimum := float64(1)
			if key == "displayOrder" && column == "display_order" && strings.Contains(r.URL.Path, "packages") {
				minimum = 0
			}
			if !ok || n < minimum || n > 2147483647 || n != float64(int(n)) {
				response.ErrorSimple(w, 400, "Invalid number")
				return nil, nil, nil, false
			}
		case "isActive":
			if _, ok := v.(bool); !ok {
				response.ErrorSimple(w, 400, "Invalid boolean")
				return nil, nil, nil, false
			}
		default:
			if v != nil {
				str, ok := v.(string)
				if !ok {
					response.ErrorSimple(w, 400, "Invalid string")
					return nil, nil, nil, false
				}
				max := 0
				switch key {
				case "name":
					max = 120
				case "title":
					max = 200
				case "description":
					max = 10000
				}
				if key != "inputData" && key != "expectedOutput" && key != "description" && !validText(str, max) {
					response.ErrorSimple(w, 400, "Invalid text")
					return nil, nil, nil, false
				}
				if key == "slug" && !slugPattern.MatchString(str) || key == "levelId" && !middleware.ValidUUID(str) || key == "executionMode" && (str != "function" && str != "stdin_stdout") {
					response.ErrorSimple(w, 400, "Invalid field value")
					return nil, nil, nil, false
				}
			}
		}
		args = append(args, v)
		sets = append(sets, fmt.Sprintf("%s=$%d", column, len(args)))
	}
	return body, sets, args, true
}
func (h *v1Handler) packageItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if r.Method == "GET" {
		h.object(w, r, "package", `SELECT `+packageJSON+` FROM packages WHERE id=$1`, id)
		return
	}
	tx, err := h.pool.Begin(r.Context())
	if h.fail(w, err) {
		return
	}
	defer tx.Rollback(r.Context())
	var status string
	var published bool
	if h.fail(w, tx.QueryRow(r.Context(), `SELECT status,published_at IS NOT NULL FROM packages WHERE id=$1 FOR UPDATE`, id).Scan(&status, &published)) {
		return
	}
	switch r.Method {
	case "DELETE":
		if published || status != "draft" {
			response.ErrorSimple(w, 409, "Only unpublished drafts can be deleted")
			return
		}
		_, err = tx.Exec(r.Context(), `DELETE FROM levels WHERE package_id=$1`, id)
		if err == nil {
			_, err = tx.Exec(r.Context(), `DELETE FROM packages WHERE id=$1`, id)
		}
	case "PATCH":
		body, sets, args, ok := patchFields(w, r, map[string]string{"slug": "slug", "name": "name", "description": "description", "displayOrder": "display_order"})
		if !ok {
			return
		}
		if _, ok = body["slug"]; ok && published {
			response.Error(w, 409, "Published slug is stable", "SLUG_IN_USE")
			return
		}
		args = append(args, id)
		_, err = tx.Exec(r.Context(), `UPDATE packages SET `+strings.Join(sets, ",")+fmt.Sprintf(" WHERE id=$%d", len(args)), args...)
	case "POST":
		next := "archived"
		if strings.HasSuffix(r.URL.Path, "/publish") {
			next = "published"
		}
		_, err = tx.Exec(r.Context(), `UPDATE packages SET status=$2::package_status,published_at=CASE WHEN $2::text='published' THEN COALESCE(published_at,now()) ELSE published_at END WHERE id=$1`, id, next)
	}
	if h.fail(w, err) || h.fail(w, tx.Commit(r.Context())) {
		return
	}
	if r.Method == "DELETE" {
		response.NoContent(w)
		return
	}
	h.object(w, r, "package", `SELECT `+packageJSON+` FROM packages WHERE id=$1`, id)
}
func (h *v1Handler) levels(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if r.Method == "PATCH" {
		_, sets, args, ok := patchFields(w, r, map[string]string{"name": "name"})
		if !ok {
			return
		}
		args = append(args, id)
		h.object(w, r, "level", `UPDATE levels SET `+strings.Join(sets, ",")+fmt.Sprintf(" WHERE id=$%d RETURNING ", len(args))+levelJSON, args...)
		return
	}
	var exists bool
	if h.fail(w, h.pool.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM packages WHERE id=$1)`, id).Scan(&exists)) {
		return
	}
	if !exists {
		response.ErrorSimple(w, 404, "Package not found")
		return
	}
	var items json.RawMessage
	err := h.pool.QueryRow(r.Context(), `SELECT COALESCE(jsonb_agg(v ORDER BY n),'[]') FROM(SELECT `+levelJSON+` v,level_no n FROM levels WHERE package_id=$1) x`, id).Scan(&items)
	if !h.fail(w, err) {
		response.Success(w, 200, map[string]any{"items": items}, "Levels retrieved")
	}
}
func (h *v1Handler) createProblem(w http.ResponseWriter, r *http.Request) {
	var v problemWrite
	if !decode(w, r, &v) {
		return
	}
	if !middleware.ValidUUID(v.LevelID) || !slugPattern.MatchString(v.Slug) || !validText(v.Title, 200) || !validText(v.Statement, 0) || (v.Mode != "function" && v.Mode != "stdin_stdout") || v.Time <= 0 || v.Memory <= 0 || v.Output <= 0 || (v.Mode == "function" && (v.Function == nil || !validText(*v.Function, 0))) {
		response.ErrorSimple(w, 400, "Invalid problem fields")
		return
	}
	var raw json.RawMessage
	err := h.pool.QueryRow(r.Context(), `INSERT INTO problems(level_id,slug,title,statement_md,execution_mode,function_name,time_limit_ms,memory_limit_mb,output_limit_bytes) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING `+problemJSON, v.LevelID, v.Slug, v.Title, v.Statement, v.Mode, v.Function, v.Time, v.Memory, v.Output).Scan(&raw)
	if !h.fail(w, err) {
		response.Success(w, 201, map[string]any{"problem": raw}, "Problem created")
	}
}
func lockProblem(ctx context.Context, tx pgx.Tx, id string) (bool, error) {
	var published bool
	err := tx.QueryRow(ctx, `SELECT published_at IS NOT NULL FROM problems WHERE id=$1 FOR UPDATE`, id).Scan(&published)
	return published, err
}
func (h *v1Handler) problemItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if r.Method == "GET" {
		h.object(w, r, "problem", `SELECT `+problemJSON+` FROM problems WHERE id=$1`, id)
		return
	}
	tx, err := h.pool.Begin(r.Context())
	if h.fail(w, err) {
		return
	}
	defer tx.Rollback(r.Context())
	published, err := lockProblem(r.Context(), tx, id)
	if h.fail(w, err) {
		return
	}
	switch r.Method {
	case "PATCH":
		body, sets, args, ok := patchFields(w, r, map[string]string{"slug": "slug", "levelId": "level_id", "title": "title", "statementMd": "statement_md", "executionMode": "execution_mode", "functionName": "function_name", "timeLimitMs": "time_limit_ms", "memoryLimitMb": "memory_limit_mb", "outputLimitBytes": "output_limit_bytes"})
		if !ok {
			return
		}
		if published {
			for k := range body {
				if k != "title" && k != "statementMd" {
					response.Error(w, 409, "Clone published problems to change judge inputs", "PUBLISHED_PROBLEM_IMMUTABLE")
					return
				}
			}
		}
		args = append(args, id)
		_, err = tx.Exec(r.Context(), `UPDATE problems SET `+strings.Join(sets, ",")+fmt.Sprintf(" WHERE id=$%d", len(args)), args...)
	case "DELETE":
		if published {
			response.ErrorSimple(w, 409, "Published problems cannot be deleted")
			return
		}
		_, err = tx.Exec(r.Context(), `DELETE FROM problem_test_cases WHERE problem_id=$1`, id)
		if err == nil {
			_, err = tx.Exec(r.Context(), `DELETE FROM problems WHERE id=$1`, id)
		}
	case "POST":
		next := "archived"
		if strings.HasSuffix(r.URL.Path, "/publish") {
			next = "published"
		}
		_, err = tx.Exec(r.Context(), `UPDATE problems SET status=$2::problem_status,published_at=CASE WHEN $2::text='published' THEN COALESCE(published_at,now()) ELSE published_at END WHERE id=$1`, id, next)
	}
	if h.fail(w, err) || h.fail(w, tx.Commit(r.Context())) {
		return
	}
	if r.Method == "DELETE" {
		response.NoContent(w)
		return
	}
	h.object(w, r, "problem", `SELECT `+problemJSON+` FROM problems WHERE id=$1`, id)
}

func (h *v1Handler) children(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	parts := strings.Split(r.URL.Path, "/")
	kind := parts[6]
	table := "problem_samples"
	if kind == "test-cases" {
		table = "problem_test_cases"
	}
	tx, err := h.pool.Begin(r.Context())
	if h.fail(w, err) {
		return
	}
	defer tx.Rollback(r.Context())
	published, err := lockProblem(r.Context(), tx, id)
	if h.fail(w, err) {
		return
	}
	if r.Method != "GET" && published && kind != "tags" {
		response.Error(w, 409, "Published judge inputs are immutable", "PUBLISHED_PROBLEM_IMMUTABLE")
		return
	}
	if kind == "tags" || kind == "languages" {
		column := "tag_id"
		link := "problem_tags"
		projection := `jsonb_build_object('id',v.id,'slug',v.slug,'name',v.name)`
		if kind == "languages" {
			column = "language_id"
			link = "problem_languages"
			projection = `jsonb_build_object('id',v.id,'slug',v.slug,'name',v.name,'runtimeVersion',v.piston_version,'fileExtension',v.file_extension,'pistonAlias',v.piston_alias,'pistonVersion',v.piston_version,'defaultFileName',v.default_file_name,'isActive',v.is_active)`
		}
		if r.Method == "PUT" {
			var b struct {
				Tags      []string `json:"tagIds"`
				Languages []string `json:"languageIds"`
			}
			if !decode(w, r, &b) {
				return
			}
			ids := b.Tags
			if kind == "languages" {
				ids = b.Languages
			}
			if ids == nil || (kind == "tags" && b.Languages != nil) || (kind == "languages" && b.Tags != nil) {
				response.ErrorSimple(w, 400, "Expected matching ID array")
				return
			}
			seen := map[string]bool{}
			for _, v := range ids {
				if !middleware.ValidUUID(v) || seen[v] {
					response.ErrorSimple(w, 400, "Invalid or duplicate ID")
					return
				}
				seen[v] = true
			}
			if kind == "languages" {
				_, err = tx.Exec(r.Context(), `DELETE FROM problem_languages WHERE problem_id=$1 AND NOT(language_id=ANY($2::uuid[]))`, id, ids)
			} else {
				_, err = tx.Exec(r.Context(), `DELETE FROM problem_tags WHERE problem_id=$1`, id)
			}
			if h.fail(w, err) {
				return
			}
			for _, v := range ids {
				_, err = tx.Exec(r.Context(), `INSERT INTO `+link+`(problem_id,`+column+`) VALUES($1,$2) ON CONFLICT DO NOTHING`, id, v)
				if h.fail(w, err) {
					return
				}
			}
		}
		var items json.RawMessage
		err = tx.QueryRow(r.Context(), `SELECT COALESCE(jsonb_agg(`+projection+` ORDER BY v.slug,v.id),'[]') FROM `+kind+` v JOIN `+link+` x ON x.`+column+`=v.id WHERE x.problem_id=$1`, id).Scan(&items)
		if h.fail(w, err) || h.fail(w, tx.Commit(r.Context())) {
			return
		}
		response.Success(w, 200, map[string]any{"items": items}, "Links retrieved")
		return
	}
	projection := `jsonb_build_object('id',id,'inputData',input_data,'expectedOutput',expected_output,'displayOrder',display_order`
	if kind == "test-cases" {
		projection += `,'problemId',problem_id,'isActive',is_active`
	}
	projection += `)`
	if r.Method == "GET" {
		var items json.RawMessage
		err = tx.QueryRow(r.Context(), `SELECT COALESCE(jsonb_agg(v ORDER BY n,i),'[]') FROM(SELECT `+projection+` v,display_order n,id i FROM `+table+` WHERE problem_id=$1) x`, id).Scan(&items)
		if !h.fail(w, err) {
			response.Success(w, 200, map[string]any{"items": items}, "Tests retrieved")
		}
		return
	}
	var raw json.RawMessage
	item := r.PathValue("itemId")
	switch r.Method {
	case "DELETE":
		result, e := tx.Exec(r.Context(), `DELETE FROM `+table+` WHERE problem_id=$1 AND id=$2`, id, item)
		err = e
		if e == nil && result.RowsAffected() == 0 {
			err = pgx.ErrNoRows
		}
	case "POST", "PATCH":
		columns := map[string]string{"inputData": "input_data", "expectedOutput": "expected_output", "displayOrder": "display_order"}
		if kind == "test-cases" {
			columns["isActive"] = "is_active"
		}
		body, sets, args, ok := patchFields(w, r, columns)
		if !ok {
			return
		}
		if r.Method == "PATCH" {
			args = append(args, id, item)
			err = tx.QueryRow(r.Context(), `UPDATE `+table+` SET `+strings.Join(sets, ",")+fmt.Sprintf(" WHERE problem_id=$%d AND id=$%d RETURNING ", len(args)-1, len(args))+projection, args...).Scan(&raw)
		} else {
			if body["expectedOutput"] == nil || body["displayOrder"] == nil || (kind == "test-cases" && body["isActive"] == nil) {
				response.ErrorSimple(w, 400, "Missing test fields")
				return
			}
			cols := []string{"problem_id"}
			values := []string{"$1"}
			params := []any{id}
			for k, v := range body {
				var value any
				json.Unmarshal(v, &value)
				cols = append(cols, columns[k])
				params = append(params, value)
				values = append(values, fmt.Sprintf("$%d", len(params)))
			}
			err = tx.QueryRow(r.Context(), `INSERT INTO `+table+` (`+strings.Join(cols, ",")+`) VALUES(`+strings.Join(values, ",")+`) RETURNING `+projection, params...).Scan(&raw)
		}
	}
	if h.fail(w, err) || h.fail(w, tx.Commit(r.Context())) {
		return
	}
	if r.Method == "DELETE" {
		response.NoContent(w)
		return
	}
	code := 200
	if r.Method == "POST" {
		code = 201
	}
	key := "sample"
	if kind == "test-cases" {
		key = "testCase"
	}
	response.Success(w, code, map[string]any{key: raw}, "Test saved")
}
func (h *v1Handler) template(w http.ResponseWriter, r *http.Request) {
	id, lang := r.PathValue("id"), r.PathValue("languageId")
	tx, err := h.pool.Begin(r.Context())
	if h.fail(w, err) {
		return
	}
	defer tx.Rollback(r.Context())
	published, err := lockProblem(r.Context(), tx, id)
	if h.fail(w, err) {
		return
	}
	if r.Method != "GET" && published {
		response.Error(w, 409, "Published templates are immutable", "PUBLISHED_PROBLEM_IMMUTABLE")
		return
	}
	projection := `jsonb_build_object('problemId',problem_id,'languageId',language_id,'starterCode',starter_code,'wrapperCode',wrapper_code,'fileName',file_name)`
	var raw json.RawMessage
	switch r.Method {
	case "GET":
		err = tx.QueryRow(r.Context(), `SELECT `+projection+` FROM code_templates WHERE problem_id=$1 AND language_id=$2`, id, lang).Scan(&raw)
	case "DELETE":
		var result pgconn.CommandTag
		result, err = tx.Exec(r.Context(), `DELETE FROM code_templates WHERE problem_id=$1 AND language_id=$2`, id, lang)
		if err == nil && result.RowsAffected() == 0 {
			err = pgx.ErrNoRows
		}
	case "PUT":
		var b struct {
			Starter *string `json:"starterCode"`
			Wrapper *string `json:"wrapperCode"`
			File    string  `json:"fileName"`
		}
		if !decode(w, r, &b) {
			return
		}
		if b.Starter == nil || !filePattern.MatchString(b.File) || b.File == "." || b.File == ".." || len(b.File) > 120 {
			response.ErrorSimple(w, 400, "Invalid template fields")
			return
		}
		err = tx.QueryRow(r.Context(), `INSERT INTO code_templates(problem_id,language_id,starter_code,wrapper_code,file_name) VALUES($1,$2,$3,$4,$5) ON CONFLICT(problem_id,language_id) DO UPDATE SET starter_code=EXCLUDED.starter_code,wrapper_code=EXCLUDED.wrapper_code,file_name=EXCLUDED.file_name RETURNING `+projection, id, lang, *b.Starter, b.Wrapper, b.File).Scan(&raw)
	}
	if h.fail(w, err) || h.fail(w, tx.Commit(r.Context())) {
		return
	}
	if r.Method == "DELETE" {
		response.NoContent(w)
		return
	}
	response.Success(w, 200, raw, "Template retrieved")
}
