package forum

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"innogen-backend/shared/testutil"
)

func TestForumModerationAudit(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.Background()
	var email string
	if err := f.Pool.QueryRow(ctx, `UPDATE users SET role='admin' WHERE id=$1 RETURNING email`, f.UserID).Scan(&email); err != nil {
		t.Fatal(err)
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": f.UserID, "userId": f.UserID, "email": email, "role": "admin",
		"exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	h := New(f.Pool, "secret")
	call := func(method, path, body string, want int) map[string]any {
		t.Helper()
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("%s %s: got %d want %d: %s", method, path, rec.Code, want, rec.Body.String())
		}
		if rec.Body.Len() == 0 {
			return nil
		}
		var env map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
			t.Fatal(err)
		}
		return env
	}

	env := call(http.MethodPost, "/api/v1/forum/posts", `{"title":"Audit post","contentMd":"Body","postType":"discussion"}`, 201)
	postID := env["data"].(map[string]any)["post"].(map[string]any)["id"].(string)
	var commentID string
	t.Cleanup(func() {
		if commentID != "" {
			f.Pool.Exec(ctx, `DELETE FROM audit_logs WHERE resource_id=$1`, commentID)
		}
		f.Pool.Exec(ctx, `DELETE FROM audit_logs WHERE resource_id=$1`, postID)
		f.Pool.Exec(ctx, `DELETE FROM forum_posts WHERE id=$1`, postID)
	})

	call(http.MethodPatch, "/api/v1/admin/forum/posts/"+postID, `{"status":"hidden"}`, 200)
	assertAudit := func(action, resourceID string) {
		t.Helper()
		var count int
		var actor string
		if err := f.Pool.QueryRow(ctx, `SELECT count(*), COALESCE(max(actor_email),'') FROM audit_logs WHERE action=$1 AND resource_id=$2`, action, resourceID).Scan(&count, &actor); err != nil {
			t.Fatal(err)
		}
		if count != 1 || actor != email {
			t.Fatalf("audit %s: count=%d actor=%q", action, count, actor)
		}
	}
	assertAudit("forum.post_hidden", postID)

	call(http.MethodPatch, "/api/v1/admin/forum/posts/"+postID, `{"status":"published"}`, 200)
	assertAudit("forum.post_restored", postID)

	env = call(http.MethodPost, "/api/v1/forum/posts/"+postID+"/comments", `{"content":"Hello"}`, 201)
	commentID = env["data"].(map[string]any)["comment"].(map[string]any)["id"].(string)
	call(http.MethodPatch, "/api/v1/admin/forum/comments/"+commentID, `{"isDeleted":true}`, 200)
	assertAudit("forum.comment_deleted", commentID)

	// Moderation without a state change must not duplicate the audit trail.
	call(http.MethodPatch, "/api/v1/admin/forum/comments/"+commentID, `{"isDeleted":true}`, 200)
	assertAudit("forum.comment_deleted", commentID)

	// Deleting a published post is audited.
	call(http.MethodPatch, "/api/v1/admin/forum/posts/"+postID, `{"status":"deleted"}`, 200)
	assertAudit("forum.post_deleted", postID)

	// Deleting a hidden post is audited too.
	env = call(http.MethodPost, "/api/v1/forum/posts", `{"title":"Audit hidden delete","contentMd":"Body","postType":"discussion"}`, 201)
	hiddenPostID := env["data"].(map[string]any)["post"].(map[string]any)["id"].(string)
	t.Cleanup(func() {
		f.Pool.Exec(ctx, `DELETE FROM audit_logs WHERE resource_id=$1`, hiddenPostID)
		f.Pool.Exec(ctx, `DELETE FROM forum_posts WHERE id=$1`, hiddenPostID)
	})
	call(http.MethodPatch, "/api/v1/admin/forum/posts/"+hiddenPostID, `{"status":"hidden"}`, 200)
	call(http.MethodPatch, "/api/v1/admin/forum/posts/"+hiddenPostID, `{"status":"deleted"}`, 200)
	assertAudit("forum.post_deleted", hiddenPostID)
}
