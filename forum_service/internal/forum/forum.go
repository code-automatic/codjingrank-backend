package forum

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"innogen-backend/shared/audit"
	"innogen-backend/shared/middleware"
	"innogen-backend/shared/response"
)

type service struct {
	pool   *pgxpool.Pool
	secret string
}

func New(pool *pgxpool.Pool, secret string) http.Handler {
	s := &service{pool: pool, secret: secret}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, 200, map[string]string{"status": "ok", "service": "forum"}, "Service is healthy")
	})
	mux.HandleFunc("GET /api/v1/forum/posts", s.listPosts)
	mux.HandleFunc("POST /api/v1/forum/posts", s.createPost)
	mux.HandleFunc("GET /api/v1/forum/posts/{id}", s.getPost)
	mux.HandleFunc("PATCH /api/v1/forum/posts/{id}", s.patchPost)
	mux.HandleFunc("DELETE /api/v1/forum/posts/{id}", s.deletePost)
	mux.HandleFunc("GET /api/v1/forum/posts/{id}/comments", s.listComments)
	mux.HandleFunc("POST /api/v1/forum/posts/{id}/comments", s.createComment)
	mux.HandleFunc("PATCH /api/v1/forum/comments/{id}", s.patchComment)
	mux.HandleFunc("DELETE /api/v1/forum/comments/{id}", s.deleteComment)
	mux.HandleFunc("PUT /api/v1/forum/posts/{id}/like", s.like)
	mux.HandleFunc("DELETE /api/v1/forum/posts/{id}/like", s.like)
	mux.HandleFunc("GET /api/v1/admin/forum/posts", s.listPosts)
	mux.HandleFunc("GET /api/v1/admin/forum/posts/{id}", s.getPost)
	mux.HandleFunc("PATCH /api/v1/admin/forum/posts/{id}", s.moderatePost)
	mux.HandleFunc("GET /api/v1/admin/forum/posts/{id}/comments", s.listComments)
	mux.HandleFunc("PATCH /api/v1/admin/forum/comments/{id}", s.moderateComment)
	return response.Router(mux)
}

func (s *service) auth(w http.ResponseWriter, r *http.Request, required, admin bool) (string, bool) {
	if r.Header.Get("Authorization") == "" && !required {
		return "", true
	}
	var id string
	middleware.Auth(s.secret)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { id, _ = middleware.GetUserID(r) })).ServeHTTP(w, r)
	if id == "" {
		return "", false
	}
	var active bool
	var role string
	err := s.pool.QueryRow(r.Context(), `SELECT is_active,role FROM users WHERE id=$1`, id).Scan(&active, &role)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && !active) {
		response.ErrorSimple(w, 403, "Account is inactive")
		return "", false
	}
	if err != nil {
		fail(w, err)
		return "", false
	}
	if admin && role != "admin" {
		response.ErrorSimple(w, 403, "Admin access required")
		return "", false
	}
	return id, true
}

// writeAudit records a moderation action without failing the user-facing request.
func (s *service) writeAudit(ctx context.Context, actorID, action, resourceType, resourceID string, detail any) {
	var email string
	if err := s.pool.QueryRow(ctx, `SELECT email FROM users WHERE id=$1`, actorID).Scan(&email); err != nil {
		email = ""
	}
	err := audit.Write(ctx, s.pool, audit.AuditEntry{
		ActorID:      actorID,
		ActorEmail:   email,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Detail:       detail,
	})
	if err != nil {
		slog.Warn("failed to write audit log", "action", action, "error", err.Error())
	}
}

func fail(w http.ResponseWriter, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		response.ErrorSimple(w, 404, "Not found")
	} else {
		response.ErrorSimple(w, 500, "Internal server error")
	}
}
func validID(w http.ResponseWriter, id string) bool {
	if !middleware.ValidUUID(id) {
		response.ErrorValidation(w, "Invalid UUID", "id")
		return false
	}
	return true
}
func adminPath(r *http.Request) bool { return strings.HasPrefix(r.URL.Path, "/api/v1/admin/") }
func page(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	p, n := 1, 20
	var err error
	if v := r.URL.Query().Get("page"); v != "" {
		p, err = strconv.Atoi(v)
		if err != nil || p < 1 || p > 1000000 {
			response.ErrorValidation(w, "Invalid page", "page")
			return 0, 0, false
		}
	}
	if v := r.URL.Query().Get("size"); v != "" {
		n, err = strconv.Atoi(v)
		if err != nil || n < 1 || n > 100 {
			response.ErrorValidation(w, "Invalid size", "size")
			return 0, 0, false
		}
	}
	return p, n, true
}

const postJSON = `jsonb_build_object('id',p.id,'author',jsonb_build_object('id',u.id,'username',u.username,'fullName',u.full_name,'avatarUrl',u.avatar_url),'problemId',p.problem_id,'languageId',p.language_id,'postType',p.post_type,'status',p.status,'title',p.title,'likeCount',p.like_count,'viewCount',p.view_count,'commentCount',p.comment_count,'likedByMe',EXISTS(SELECT 1 FROM forum_post_likes fl WHERE fl.post_id=p.id AND fl.user_id=NULLIF($2,'')::uuid),'createdAt',p.created_at,'updatedAt',p.updated_at,'contentMd',p.content_md,'codeSnippet',p.code_snippet)`

func (s *service) post(ctx context.Context, id, user string, admin bool) (map[string]any, error) {
	q := `SELECT ` + postJSON + ` || CASE WHEN $3 THEN jsonb_build_object('deletedAt',p.deleted_at) ELSE '{}'::jsonb END FROM forum_posts p JOIN users u ON u.id=p.user_id WHERE p.id=$1 AND ($3 OR p.status='published')`
	var result map[string]any
	err := s.pool.QueryRow(ctx, q, id, user, admin).Scan(&result)
	return result, err
}
func (s *service) getPost(w http.ResponseWriter, r *http.Request) {
	admin := adminPath(r)
	user, ok := s.auth(w, r, admin, admin)
	if !ok || !validID(w, r.PathValue("id")) {
		return
	}
	p, err := s.post(r.Context(), r.PathValue("id"), user, admin)
	if err != nil {
		fail(w, err)
		return
	}
	if !admin {
		viewer := "user:" + user
		if user == "" {
			cookie, err := r.Cookie("forum_session")
			if err != nil || len(cookie.Value) != 64 {
				var token [32]byte
				if _, err := rand.Read(token[:]); err != nil {
					fail(w, err)
					return
				}
				cookie = &http.Cookie{Name: "forum_session", Value: hex.EncodeToString(token[:]), Path: "/", HttpOnly: true, Secure: r.TLS != nil, SameSite: http.SameSiteLaxMode}
				http.SetCookie(w, cookie)
			}
			viewer = "guest:" + cookie.Value
		}
		var count int64
		err = s.pool.QueryRow(r.Context(), `WITH inserted AS (INSERT INTO forum_post_views(post_id,viewer_key) SELECT id,$2 FROM forum_posts WHERE id=$1 AND status='published' ON CONFLICT DO NOTHING RETURNING post_id) UPDATE forum_posts SET view_count=view_count+1 WHERE id IN (SELECT post_id FROM inserted) RETURNING view_count`, r.PathValue("id"), viewer).Scan(&count)
		if err == nil {
			p["viewCount"] = count
		} else if !errors.Is(err, pgx.ErrNoRows) {
			fail(w, err)
			return
		}
	}
	message := "Post retrieved"
	if admin {
		message = "Moderation post retrieved"
	}
	response.Success(w, 200, map[string]any{"post": p}, message)
}
func (s *service) listPosts(w http.ResponseWriter, r *http.Request) {
	admin := adminPath(r)
	user, ok := s.auth(w, r, admin, admin)
	if !ok {
		return
	}
	p, n, ok := page(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	problem, lang, kind, status, sortBy, search := q.Get("problemId"), q.Get("languageId"), q.Get("postType"), q.Get("status"), q.Get("sort"), q.Get("q")
	for _, v := range []string{problem, lang} {
		if v != "" && !middleware.ValidUUID(v) {
			response.ErrorValidation(w, "Invalid UUID", "filter")
			return
		}
	}
	if kind != "" && kind != "discussion" && kind != "question" && kind != "solution" {
		response.ErrorValidation(w, "Invalid post type", "postType")
		return
	}
	if status != "" && (!admin || (status != "published" && status != "hidden" && status != "deleted")) {
		response.ErrorValidation(w, "Invalid status", "status")
		return
	}
	order := "p.created_at DESC,p.id"
	switch sortBy {
	case "", "newest":
	case "likes":
		order = "p.like_count DESC,p.created_at DESC,p.id"
	case "views":
		order = "p.view_count DESC,p.created_at DESC,p.id"
	default:
		response.ErrorValidation(w, "Invalid sort", "sort")
		return
	}
	where := ` FROM forum_posts p JOIN users u ON u.id=p.user_id WHERE $1::text IS NOT NULL AND $2::text IS NOT NULL AND ($3 OR p.status='published') AND ($4='' OR p.status::text=$4) AND ($5='' OR p.problem_id=NULLIF($5,'')::uuid) AND ($6='' OR p.language_id=NULLIF($6,'')::uuid) AND ($7='' OR p.post_type::text=$7) AND ($8='' OR p.title ILIKE '%'||$8||'%' OR p.content_md ILIKE '%'||$8||'%')`
	args := []any{"", user, admin, status, problem, lang, kind, search}
	var total int
	if err := s.pool.QueryRow(r.Context(), `SELECT count(*)`+where, args...).Scan(&total); err != nil {
		fail(w, err)
		return
	}
	fields := postJSON
	if admin {
		fields += ` || jsonb_build_object('deletedAt',p.deleted_at)`
	} else {
		fields += ` - 'contentMd' - 'codeSnippet'`
	}
	rows, err := s.pool.Query(r.Context(), `SELECT `+fields+where+` ORDER BY `+order+` LIMIT $9 OFFSET $10`, append(args, n, (p-1)*n)...)
	if err != nil {
		fail(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var item map[string]any
		if err = rows.Scan(&item); err != nil {
			fail(w, err)
			return
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		fail(w, err)
		return
	}
	message := "Posts retrieved"
	if admin {
		message = "Moderation posts retrieved"
	}
	response.Success(w, 200, map[string]any{"items": items, "page": p, "size": n, "total": total}, message)
}

type postInput struct {
	Title    string  `json:"title"`
	Content  string  `json:"contentMd"`
	Code     *string `json:"codeSnippet"`
	Type     string  `json:"postType"`
	Problem  *string `json:"problemId"`
	Language *string `json:"languageId"`
}

func textValid(s string, max int) bool {
	return strings.TrimSpace(s) != "" && utf8.RuneCountInString(s) <= max
}
func validatePost(ctx context.Context, queryRow func(context.Context, string, ...any) pgx.Row, user string, p *postInput) (int, string) {
	if !textValid(p.Title, 200) || !textValid(p.Content, 20000) || (p.Code != nil && utf8.RuneCountInString(*p.Code) > 102400) || (p.Type != "discussion" && p.Type != "question" && p.Type != "solution") {
		return 400, "Invalid post"
	}
	for _, v := range []*string{p.Problem, p.Language} {
		if v != nil && !middleware.ValidUUID(*v) {
			return 400, "Invalid UUID"
		}
	}
	if p.Type == "solution" {
		if p.Problem == nil {
			return 403, "SOLUTION_NOT_ALLOWED"
		}
		var ok bool
		err := queryRow(ctx, `SELECT EXISTS(SELECT 1 FROM user_problem_progress WHERE user_id=$1 AND problem_id=$2)`, user, *p.Problem).Scan(&ok)
		if err != nil {
			return 500, "Internal server error"
		}
		if !ok {
			return 403, "SOLUTION_NOT_ALLOWED"
		}
	}
	if p.Problem != nil {
		var ok bool
		err := queryRow(ctx, `SELECT EXISTS(SELECT 1 FROM problems p JOIN levels l ON l.id=p.level_id JOIN packages k ON k.id=l.package_id WHERE p.id=$1 AND p.status='published' AND k.status='published')`, *p.Problem).Scan(&ok)
		if err != nil {
			return 500, "Internal server error"
		}
		if !ok {
			return 400, "Invalid problem"
		}
	}
	if p.Language != nil {
		var ok bool
		err := queryRow(ctx, `SELECT EXISTS(SELECT 1 FROM languages l WHERE l.id=$1 AND l.is_active AND ($2::uuid IS NULL OR EXISTS(SELECT 1 FROM problem_languages pl WHERE pl.language_id=l.id AND pl.problem_id=$2)))`, *p.Language, p.Problem).Scan(&ok)
		if err != nil {
			return 500, "Internal server error"
		}
		if !ok {
			return 400, "Invalid language"
		}
	}
	var err error
	p.Content, err = safeMarkdown(p.Content)
	if err != nil {
		return 400, "Invalid Markdown"
	}
	return 0, ""
}
func (s *service) createPost(w http.ResponseWriter, r *http.Request) {
	user, ok := s.auth(w, r, true, false)
	if !ok {
		return
	}
	var p postInput
	if response.DecodeJSON(r, &p) != nil {
		response.ErrorSimple(w, 400, "Invalid JSON")
		return
	}
	if code, msg := validatePost(r.Context(), s.pool.QueryRow, user, &p); code != 0 {
		if msg == "SOLUTION_NOT_ALLOWED" {
			response.Error(w, code, "Solution requires an accepted submission", msg)
		} else {
			response.ErrorSimple(w, code, msg)
		}
		return
	}
	var id string
	err := s.pool.QueryRow(r.Context(), `INSERT INTO forum_posts(user_id,problem_id,language_id,post_type,title,content_md,code_snippet) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id`, user, p.Problem, p.Language, p.Type, p.Title, p.Content, p.Code).Scan(&id)
	if err != nil {
		fail(w, err)
		return
	}
	result, err := s.post(r.Context(), id, user, false)
	if err != nil {
		fail(w, err)
		return
	}
	response.Success(w, 201, map[string]any{"post": result}, "Post created")
}
func (s *service) patchPost(w http.ResponseWriter, r *http.Request) {
	user, ok := s.auth(w, r, true, false)
	id := r.PathValue("id")
	if !ok || !validID(w, id) {
		return
	}
	var fields map[string]json.RawMessage
	if response.DecodeJSON(r, &fields) != nil || len(fields) == 0 {
		response.ErrorSimple(w, 400, "Invalid JSON")
		return
	}
	var p postInput
	ctx := r.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		fail(w, err)
		return
	}
	defer tx.Rollback(ctx)
	err = tx.QueryRow(ctx, `SELECT title,content_md,code_snippet,post_type,problem_id,language_id FROM forum_posts WHERE id=$1 AND user_id=$2 AND status='published' FOR UPDATE`, id, user).Scan(&p.Title, &p.Content, &p.Code, &p.Type, &p.Problem, &p.Language)
	if err != nil {
		fail(w, err)
		return
	}
	for key, value := range fields {
		var target any
		switch key {
		case "title":
			target = &p.Title
		case "contentMd":
			target = &p.Content
		case "codeSnippet":
			target = &p.Code
		case "postType":
			target = &p.Type
		case "problemId":
			target = &p.Problem
		case "languageId":
			target = &p.Language
		default:
			response.ErrorSimple(w, 400, "Unknown field")
			return
		}
		if string(value) == "null" && (key == "title" || key == "contentMd" || key == "postType") {
			response.ErrorSimple(w, 400, "Invalid field")
			return
		}
		if json.Unmarshal(value, target) != nil {
			response.ErrorSimple(w, 400, "Invalid field")
			return
		}
	}
	oldContent := p.Content
	if code, msg := validatePost(ctx, tx.QueryRow, user, &p); code != 0 {
		if msg == "SOLUTION_NOT_ALLOWED" {
			response.Error(w, code, "Solution requires an accepted submission", msg)
		} else {
			response.ErrorSimple(w, code, msg)
		}
		return
	}
	if _, changed := fields["contentMd"]; !changed {
		p.Content = oldContent
	}
	tag, err := tx.Exec(ctx, `UPDATE forum_posts SET title=$3,content_md=$4,code_snippet=$5,post_type=$6,problem_id=$7,language_id=$8 WHERE id=$1 AND user_id=$2 AND status='published'`, id, user, p.Title, p.Content, p.Code, p.Type, p.Problem, p.Language)
	if err != nil {
		fail(w, err)
		return
	}
	if tag.RowsAffected() == 0 {
		response.ErrorSimple(w, 409, "Post changed; retry the update")
		return
	}
	if err = tx.Commit(ctx); err != nil {
		fail(w, err)
		return
	}
	post, err := s.post(r.Context(), id, user, false)
	if err != nil {
		fail(w, err)
		return
	}
	response.Success(w, 200, map[string]any{"post": post}, "Post updated")
}
func (s *service) deletePost(w http.ResponseWriter, r *http.Request) {
	user, ok := s.auth(w, r, true, false)
	id := r.PathValue("id")
	if !ok || !validID(w, id) {
		return
	}
	tag, err := s.pool.Exec(r.Context(), `UPDATE forum_posts SET status='deleted',deleted_at=COALESCE(deleted_at,now()) WHERE id=$1 AND user_id=$2 AND status IN ('published','deleted')`, id, user)
	if err != nil {
		fail(w, err)
		return
	}
	if tag.RowsAffected() == 0 {
		fail(w, pgx.ErrNoRows)
		return
	}
	response.NoContent(w)
}

const commentJSON = `jsonb_build_object('id',c.id,'postId',c.post_id,'parentId',c.parent_id,'author',jsonb_build_object('id',u.id,'username',u.username,'fullName',u.full_name,'avatarUrl',u.avatar_url),'content',CASE WHEN c.is_deleted THEN NULL ELSE c.content END,'isDeleted',c.is_deleted,'createdAt',c.created_at,'updatedAt',c.updated_at)`

func (s *service) comment(ctx context.Context, id string) (map[string]any, error) {
	var result map[string]any
	err := s.pool.QueryRow(ctx, `SELECT `+commentJSON+` FROM forum_comments c JOIN users u ON u.id=c.user_id WHERE c.id=$1`, id).Scan(&result)
	return result, err
}
func (s *service) listComments(w http.ResponseWriter, r *http.Request) {
	admin := adminPath(r)
	_, ok := s.auth(w, r, admin, admin)
	id := r.PathValue("id")
	if !ok || !validID(w, id) {
		return
	}
	p, n, ok := page(w, r)
	if !ok {
		return
	}
	var exists bool
	err := s.pool.QueryRow(r.Context(), `SELECT EXISTS(SELECT 1 FROM forum_posts WHERE id=$1 AND ($2 OR status='published'))`, id, admin).Scan(&exists)
	if err != nil {
		fail(w, err)
		return
	}
	if !exists {
		fail(w, pgx.ErrNoRows)
		return
	}
	var total int
	visibility := ` AND ($2 OR NOT c.is_deleted OR (c.parent_id IS NULL AND EXISTS(SELECT 1 FROM forum_comments child WHERE child.parent_id=c.id AND NOT child.is_deleted)))`
	if err = s.pool.QueryRow(r.Context(), `SELECT count(*) FROM forum_comments c WHERE c.post_id=$1`+visibility, id, admin).Scan(&total); err != nil {
		fail(w, err)
		return
	}
	rows, err := s.pool.Query(r.Context(), `SELECT `+commentJSON+` FROM forum_comments c JOIN users u ON u.id=c.user_id WHERE c.post_id=$1`+visibility+` ORDER BY c.created_at,c.id LIMIT $3 OFFSET $4`, id, admin, n, (p-1)*n)
	if err != nil {
		fail(w, err)
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var item map[string]any
		if err = rows.Scan(&item); err != nil {
			fail(w, err)
			return
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		fail(w, err)
		return
	}
	msg := "Comments retrieved"
	if admin {
		msg = "Moderation comments retrieved"
	}
	response.Success(w, 200, map[string]any{"items": items, "page": p, "size": n, "total": total}, msg)
}
func (s *service) createComment(w http.ResponseWriter, r *http.Request) {
	user, ok := s.auth(w, r, true, false)
	id := r.PathValue("id")
	if !ok || !validID(w, id) {
		return
	}
	var input struct {
		Content string  `json:"content"`
		Parent  *string `json:"parentId"`
	}
	if response.DecodeJSON(r, &input) != nil || !textValid(input.Content, 5000) {
		response.ErrorSimple(w, 400, "Invalid comment")
		return
	}
	if input.Parent != nil && !middleware.ValidUUID(*input.Parent) {
		response.ErrorValidation(w, "Invalid UUID", "parentId")
		return
	}
	ctx := r.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		fail(w, err)
		return
	}
	defer tx.Rollback(ctx)
	var exists bool
	err = tx.QueryRow(ctx, `SELECT true FROM forum_posts WHERE id=$1 AND status='published' FOR UPDATE`, id).Scan(&exists)
	if err != nil {
		fail(w, err)
		return
	}
	if input.Parent != nil {
		err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM forum_comments WHERE id=$1 AND post_id=$2 AND parent_id IS NULL AND NOT is_deleted)`, *input.Parent, id).Scan(&exists)
		if err != nil {
			fail(w, err)
			return
		}
		if !exists {
			response.Error(w, 422, "Invalid comment parent", "INVALID_COMMENT_PARENT")
			return
		}
	}
	var commentID string
	err = tx.QueryRow(ctx, `INSERT INTO forum_comments(post_id,user_id,parent_id,content) VALUES($1,$2,$3,$4) RETURNING id`, id, user, input.Parent, input.Content).Scan(&commentID)
	if err != nil {
		fail(w, err)
		return
	}
	if _, err = tx.Exec(ctx, `UPDATE forum_posts SET comment_count=comment_count+1 WHERE id=$1`, id); err != nil {
		fail(w, err)
		return
	}
	if err = tx.Commit(ctx); err != nil {
		fail(w, err)
		return
	}
	comment, err := s.comment(ctx, commentID)
	if err != nil {
		fail(w, err)
		return
	}
	response.Success(w, 201, map[string]any{"comment": comment}, "Comment created")
}
func (s *service) patchComment(w http.ResponseWriter, r *http.Request) {
	user, ok := s.auth(w, r, true, false)
	id := r.PathValue("id")
	if !ok || !validID(w, id) {
		return
	}
	var input struct {
		Content string `json:"content"`
	}
	if response.DecodeJSON(r, &input) != nil || !textValid(input.Content, 5000) {
		response.ErrorSimple(w, 400, "Invalid comment")
		return
	}
	tag, err := s.pool.Exec(r.Context(), `UPDATE forum_comments c SET content=$3 FROM forum_posts p WHERE c.id=$1 AND c.user_id=$2 AND NOT c.is_deleted AND p.id=c.post_id AND p.status='published'`, id, user, input.Content)
	if err != nil {
		fail(w, err)
		return
	}
	if tag.RowsAffected() == 0 {
		fail(w, pgx.ErrNoRows)
		return
	}
	comment, err := s.comment(r.Context(), id)
	if err != nil {
		fail(w, err)
		return
	}
	response.Success(w, 200, map[string]any{"comment": comment}, "Comment updated")
}
func (s *service) deleteComment(w http.ResponseWriter, r *http.Request) { s.removeComment(w, r, false) }
func (s *service) moderateComment(w http.ResponseWriter, r *http.Request) {
	s.removeComment(w, r, true)
}
func (s *service) removeComment(w http.ResponseWriter, r *http.Request, admin bool) {
	user, ok := s.auth(w, r, true, admin)
	id := r.PathValue("id")
	if !ok || !validID(w, id) {
		return
	}
	if admin {
		var input struct {
			Deleted bool `json:"isDeleted"`
		}
		if response.DecodeJSON(r, &input) != nil || !input.Deleted {
			response.ErrorSimple(w, 400, "Invalid moderation")
			return
		}
	}
	ctx := r.Context()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		fail(w, err)
		return
	}
	defer tx.Rollback(ctx)
	var postID, owner, status string
	var deleted bool
	err = tx.QueryRow(ctx, `SELECT c.post_id,c.user_id,c.is_deleted,p.status FROM forum_comments c JOIN forum_posts p ON p.id=c.post_id WHERE c.id=$1 FOR UPDATE OF c,p`, id).Scan(&postID, &owner, &deleted, &status)
	if err != nil {
		fail(w, err)
		return
	}
	if !admin && (owner != user || status != "published") {
		fail(w, pgx.ErrNoRows)
		return
	}
	if !deleted {
		_, err = tx.Exec(ctx, `UPDATE forum_comments SET is_deleted=true WHERE id=$1`, id)
		if err == nil {
			_, err = tx.Exec(ctx, `UPDATE forum_posts SET comment_count=GREATEST(comment_count-1,0) WHERE id=$1`, postID)
		}
		if err != nil {
			fail(w, err)
			return
		}
	}
	if err = tx.Commit(ctx); err != nil {
		fail(w, err)
		return
	}
	if !admin {
		response.NoContent(w)
		return
	}
	if !deleted {
		s.writeAudit(ctx, user, "forum.comment_deleted", "forum_comment", id, map[string]any{"postId": postID})
	}
	comment, err := s.comment(ctx, id)
	if err != nil {
		fail(w, err)
		return
	}
	response.Success(w, 200, map[string]any{"comment": comment}, "Comment moderated")
}
func (s *service) moderatePost(w http.ResponseWriter, r *http.Request) {
	user, ok := s.auth(w, r, true, true)
	id := r.PathValue("id")
	if !ok || !validID(w, id) {
		return
	}
	var input struct {
		Status string `json:"status"`
	}
	if response.DecodeJSON(r, &input) != nil || (input.Status != "published" && input.Status != "hidden" && input.Status != "deleted") {
		response.ErrorSimple(w, 400, "Invalid status")
		return
	}
	var current string
	err := s.pool.QueryRow(r.Context(), `SELECT status FROM forum_posts WHERE id=$1`, id).Scan(&current)
	if err != nil {
		fail(w, err)
		return
	}
	if current == "deleted" && input.Status != "deleted" {
		response.Error(w, 409, "Post is deleted", "POST_DELETED")
		return
	}
	tag, err := s.pool.Exec(r.Context(), `UPDATE forum_posts SET status=$2::forum_post_status,deleted_at=CASE WHEN $2::forum_post_status='deleted' THEN COALESCE(deleted_at,now()) ELSE NULL END WHERE id=$1 AND (status<>'deleted' OR $2::forum_post_status='deleted')`, id, input.Status)
	if err != nil {
		fail(w, err)
		return
	}
	if tag.RowsAffected() == 0 {
		response.Error(w, 409, "Post is deleted", "POST_DELETED")
		return
	}
	switch {
	case input.Status == "hidden" && current != "hidden":
		s.writeAudit(r.Context(), user, "forum.post_hidden", "forum_post", id, map[string]any{"from": current, "to": input.Status})
	case input.Status == "deleted" && current != "deleted":
		s.writeAudit(r.Context(), user, "forum.post_deleted", "forum_post", id, map[string]any{"from": current, "to": input.Status})
	case input.Status == "published" && current != "published":
		s.writeAudit(r.Context(), user, "forum.post_restored", "forum_post", id, map[string]any{"from": current, "to": input.Status})
	}
	post, err := s.post(r.Context(), id, user, true)
	if err != nil {
		fail(w, err)
		return
	}
	response.Success(w, 200, map[string]any{"post": post}, "Post moderated")
}
func (s *service) like(w http.ResponseWriter, r *http.Request) {
	user, ok := s.auth(w, r, true, false)
	if !ok || !validID(w, r.PathValue("id")) {
		return
	}
	ctx := r.Context()
	id := r.PathValue("id")
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		fail(w, err)
		return
	}
	defer tx.Rollback(ctx)
	var count int
	err = tx.QueryRow(ctx, `SELECT like_count FROM forum_posts WHERE id=$1 AND status='published' FOR UPDATE`, id).Scan(&count)
	if err != nil {
		fail(w, err)
		return
	}
	liked := r.Method == http.MethodPut
	var tag string
	if liked {
		tag = `INSERT INTO forum_post_likes(post_id,user_id) VALUES($1,$2) ON CONFLICT DO NOTHING`
	} else {
		tag = `DELETE FROM forum_post_likes WHERE post_id=$1 AND user_id=$2`
	}
	res, err := tx.Exec(ctx, tag, id, user)
	if err != nil {
		fail(w, err)
		return
	}
	if res.RowsAffected() > 0 {
		delta := 1
		if !liked {
			delta = -1
		}
		count += delta
		_, err = tx.Exec(ctx, `UPDATE forum_posts SET like_count=$2 WHERE id=$1`, id, count)
		if err != nil {
			fail(w, err)
			return
		}
	}
	if err = tx.Commit(ctx); err != nil {
		fail(w, err)
		return
	}
	msg := "Post liked"
	if !liked {
		msg = "Post unliked"
	}
	response.Success(w, 200, map[string]any{"liked": liked, "likeCount": count}, msg)
}
