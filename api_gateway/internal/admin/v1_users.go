package admin

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"unicode/utf8"

	"innogen-backend/shared/audit"
	"innogen-backend/shared/middleware"
	"innogen-backend/shared/response"
)

var userRoles = map[string]bool{"student": true, "teacher": true, "admin": true}

func (h *v1Handler) users(w http.ResponseWriter, r *http.Request) {
	page, size, ok := adminPage(r)
	if !ok {
		response.ErrorSimple(w, 400, "Invalid pagination")
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	role := r.URL.Query().Get("role")
	if role != "" && !userRoles[role] {
		response.ErrorSimple(w, 400, "Invalid role filter")
		return
	}
	var isActive *bool
	if raw := r.URL.Query().Get("isActive"); raw != "" {
		switch raw {
		case "true":
			value := true
			isActive = &value
		case "false":
			value := false
			isActive = &value
		default:
			response.ErrorSimple(w, 400, "Invalid active filter")
			return
		}
	}
	items, total, err := h.repo.FindUsers(r.Context(), page, size, q, role, isActive)
	if h.fail(w, err) {
		return
	}
	response.Success(w, 200, map[string]any{"items": items, "page": page, "size": size, "total": total}, "Users retrieved")
}

func (h *v1Handler) userItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if r.Method == "GET" {
		user, err := h.repo.FindUserByID(r.Context(), id)
		if h.fail(w, err) {
			return
		}
		if user == nil {
			response.ErrorSimple(w, 404, "User not found")
			return
		}
		response.Success(w, 200, map[string]any{"user": user}, "User retrieved")
		return
	}

	patch, fields, ok := decodeUserPatch(w, r)
	if !ok {
		return
	}
	if _, present := fields["role"]; present && patch.Role == nil {
		response.ErrorSimple(w, 400, "Role must be a string")
		return
	}
	if patch.Role != nil && !userRoles[*patch.Role] {
		response.ErrorSimple(w, 400, "Role must be student, teacher, or admin")
		return
	}
	if _, present := fields["isActive"]; present && patch.IsActive == nil {
		response.ErrorSimple(w, 400, "isActive must be a boolean")
		return
	}
	if _, present := fields["fullName"]; present {
		if patch.FullName == nil {
			// Explicit JSON null clears the name.
			empty := ""
			patch.FullName = &empty
		} else if utf8.RuneCountInString(*patch.FullName) > 200 {
			response.ErrorSimple(w, 400, "fullName is too long")
			return
		} else if strings.TrimSpace(*patch.FullName) == "" {
			empty := ""
			patch.FullName = &empty
		}
	}

	actorID, _ := middleware.GetUserID(r)
	if actorID != "" && id == actorID {
		_, changingRole := fields["role"]
		_, changingStatus := fields["isActive"]
		if changingRole || changingStatus {
			response.Error(w, 409, "Cannot change your own role or status", "LAST_ADMIN_PROTECTED")
			return
		}
	}

	before, err := h.repo.FindUserByID(r.Context(), id)
	if h.fail(w, err) {
		return
	}
	if before == nil {
		response.ErrorSimple(w, 404, "User not found")
		return
	}
	updated, err := h.repo.UpdateUser(r.Context(), id, patch)
	if errors.Is(err, ErrLastAdminProtected) {
		response.Error(w, 409, "Cannot remove the last active admin", "LAST_ADMIN_PROTECTED")
		return
	}
	if h.fail(w, err) {
		return
	}
	if updated == nil {
		response.ErrorSimple(w, 404, "User not found")
		return
	}
	h.auditUserPatch(r, before, updated)
	response.Success(w, 200, map[string]any{"user": updated}, "User updated")
}

func decodeUserPatch(w http.ResponseWriter, r *http.Request) (UpdateUserRequest, map[string]json.RawMessage, bool) {
	body, err := io.ReadAll(r.Body)
	if err != nil || len(bytes.TrimSpace(body)) == 0 {
		response.ErrorSimple(w, 400, "Invalid request body")
		return UpdateUserRequest{}, nil, false
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(body, &fields); err != nil || fields == nil {
		response.ErrorSimple(w, 400, "Invalid request body")
		return UpdateUserRequest{}, nil, false
	}
	var patch UpdateUserRequest
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&patch); err != nil {
		response.ErrorSimple(w, 400, "Invalid request body")
		return UpdateUserRequest{}, nil, false
	}
	if decoder.Decode(new(any)) != io.EOF {
		response.ErrorSimple(w, 400, "Expected one JSON object")
		return UpdateUserRequest{}, nil, false
	}
	if len(fields) == 0 {
		response.ErrorSimple(w, 400, "Empty patch")
		return UpdateUserRequest{}, nil, false
	}
	for key := range fields {
		if key != "role" && key != "isActive" && key != "fullName" {
			response.ErrorSimple(w, 400, "Unknown field")
			return UpdateUserRequest{}, nil, false
		}
	}
	return patch, fields, true
}

func (h *v1Handler) auditUserPatch(r *http.Request, before, after *AdminUser) {
	actorID, _ := middleware.GetUserID(r)
	var actorEmail string
	if actorID != "" {
		if err := h.pool.QueryRow(r.Context(), `SELECT email FROM users WHERE id=$1`, actorID).Scan(&actorEmail); err != nil {
			actorEmail = ""
			h.log.Warn("failed to load audit actor", slog.String("error", err.Error()))
		}
	}
	write := func(action string, detail any) {
		if err := audit.Write(r.Context(), h.pool, audit.AuditEntry{
			ActorID:      actorID,
			ActorEmail:   actorEmail,
			Action:       action,
			ResourceType: "user",
			ResourceID:   after.ID,
			Detail:       detail,
		}); err != nil {
			h.log.Warn("failed to write audit log", slog.String("action", action), slog.String("error", err.Error()))
		}
	}
	if before.Role != after.Role {
		write("user.role_changed", map[string]any{"from": before.Role, "to": after.Role})
	}
	if before.IsActive != after.IsActive {
		action := "user.locked"
		if after.IsActive {
			action = "user.unlocked"
		}
		write(action, map[string]any{"from": before.IsActive, "to": after.IsActive})
	}
	if !sameOptionalString(before.FullName, after.FullName) {
		write("user.updated", map[string]any{"field": "fullName", "from": before.FullName, "to": after.FullName})
	}
}

func sameOptionalString(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

const auditLogJSON = `jsonb_build_object('id',id,'actorId',actor_id,'actorEmail',actor_email,'action',action,'resourceType',resource_type,'resourceId',resource_id,'detail',detail,'createdAt',created_at)`

func (h *v1Handler) auditLogs(w http.ResponseWriter, r *http.Request) {
	page, size, ok := adminPage(r)
	if !ok {
		response.ErrorSimple(w, 400, "Invalid pagination")
		return
	}
	action := r.URL.Query().Get("action")
	actorEmail := strings.TrimSpace(r.URL.Query().Get("actorEmail"))
	where := ` WHERE ($1 = '' OR action = $1) AND ($2 = '' OR actor_email ILIKE '%' || $2 || '%')`
	args := []any{action, actorEmail}
	var total int
	if err := h.pool.QueryRow(r.Context(), `SELECT count(*) FROM audit_logs`+where, args...).Scan(&total); err != nil {
		h.fail(w, err)
		return
	}
	args = append(args, size, (page-1)*size)
	var raw json.RawMessage
	err := h.pool.QueryRow(r.Context(),
		`SELECT COALESCE(jsonb_agg(v ORDER BY c DESC, i DESC), '[]')
		 FROM (SELECT `+auditLogJSON+` v, created_at c, id i FROM audit_logs`+where+` ORDER BY created_at DESC, id DESC LIMIT $3 OFFSET $4) x`,
		args...,
	).Scan(&raw)
	if h.fail(w, err) {
		return
	}
	items := make([]AuditLog, 0)
	if err = json.Unmarshal(raw, &items); err != nil {
		h.log.Error("decode audit logs failed", slog.String("error", err.Error()))
		response.ErrorSimple(w, 500, "Internal server error")
		return
	}
	response.Success(w, 200, map[string]any{"items": items, "page": page, "size": size, "total": total}, "Audit logs retrieved")
}

func (h *v1Handler) stats(w http.ResponseWriter, r *http.Request) {
	var stats AdminStats
	err := h.pool.QueryRow(r.Context(), `SELECT
		(SELECT count(*) FROM users),
		(SELECT count(*) FROM users WHERE is_active),
		(SELECT count(*) FROM packages),
		(SELECT count(*) FROM problems WHERE status='published'),
		(SELECT count(*) FROM submissions),
		(SELECT count(*) FROM submissions WHERE status='Accepted'),
		(SELECT count(DISTINCT user_id) FROM accepted_solutions),
		(SELECT count(*) FROM forum_posts)`).Scan(
		&stats.Users, &stats.ActiveUsers,
		&stats.Packages, &stats.PublishedProblems,
		&stats.Submissions, &stats.AcceptedSubmissions,
		&stats.UsersWithSolutions, &stats.ForumPosts,
	)
	if h.fail(w, err) {
		return
	}
	rows, err := h.pool.Query(r.Context(), `SELECT p.id::text, p.slug, p.title, count(DISTINCT a.user_id)
		FROM problems p JOIN accepted_solutions a ON a.problem_id=p.id
		GROUP BY p.id, p.slug, p.title
		ORDER BY count(DISTINCT a.user_id) DESC, p.id
		LIMIT 5`)
	if h.fail(w, err) {
		return
	}
	defer rows.Close()
	stats.MostSolvedProblems = make([]MostSolvedProblem, 0, 5)
	for rows.Next() {
		var item MostSolvedProblem
		if err = rows.Scan(&item.ProblemID, &item.Slug, &item.Title, &item.SolvedUserCount); err != nil {
			h.fail(w, err)
			return
		}
		stats.MostSolvedProblems = append(stats.MostSolvedProblems, item)
	}
	if h.fail(w, rows.Err()) {
		return
	}
	response.Success(w, 200, stats, "Statistics retrieved")
}
