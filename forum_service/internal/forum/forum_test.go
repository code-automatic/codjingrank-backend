package forum

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"innogen-backend/shared/testutil"
)

func TestForumOwnershipAndIdempotentLike(t *testing.T) {
	f := testutil.Catalog(t)
	t.Cleanup(func() {
		if _, err := f.Pool.Exec(context.Background(), `DELETE FROM forum_posts WHERE user_id=$1`, f.UserID); err != nil {
			t.Error(err)
		}
	})
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"userId": f.UserID, "sub": f.UserID, "exp": time.Now().Add(time.Hour).Unix()}).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := f.Pool.Config()
	cfg.MaxConns = 1
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	h := New(pool, "secret")
	call := func(method, path, body string, auth bool) *httptest.ResponseRecorder {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		r := httptest.NewRequestWithContext(ctx, method, path, bytes.NewBufferString(body))
		if auth {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code < 400 && w.Code != 204 {
			testutil.Capture(t, w, method, r.Pattern)
		}
		return w
	}
	w := call(http.MethodPost, "/api/v1/forum/posts", `{"title":"Question","contentMd":"Hello","postType":"question"}`, true)
	if w.Code != 201 {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	var result struct {
		Data struct {
			Post struct {
				ID string `json:"id"`
			} `json:"post"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.Data.Post.ID == "" {
		t.Fatalf("response: %s %v", w.Body, err)
	}
	w = call(http.MethodGet, "/api/v1/forum/posts", "", false)
	if w.Code != 200 || !bytes.Contains(w.Body.Bytes(), []byte(`"title":"Question"`)) {
		t.Fatalf("list: %d %s", w.Code, w.Body)
	}
	postPath := "/api/v1/forum/posts/" + result.Data.Post.ID
	w = call(http.MethodPatch, postPath, `{"title":"Edited question","problemId":"`+f.ProblemID+`"}`, true)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "Edited question") {
		t.Fatalf("update post: %d %s", w.Code, w.Body)
	}
	for i := 0; i < 2; i++ {
		w = call(http.MethodGet, postPath, "", true)
		if w.Code != 200 || !bytes.Contains(w.Body.Bytes(), []byte(`"viewCount":1`)) {
			t.Fatalf("deduplicated views: %d %s", w.Code, w.Body)
		}
	}
	w = call(http.MethodPatch, postPath, `{"postType":"solution"}`, true)
	if w.Code != 403 {
		t.Fatalf("convert unsolved post: %d %s", w.Code, w.Body)
	}
	w = call(http.MethodPost, postPath+"/comments", `{"content":"root"}`, true)
	if w.Code != 201 {
		t.Fatalf("comment: %d %s", w.Code, w.Body)
	}
	var root struct {
		Data struct {
			Comment struct {
				ID string `json:"id"`
			} `json:"comment"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &root); err != nil {
		t.Fatal(err)
	}
	w = call(http.MethodPost, postPath+"/comments", `{"content":"reply","parentId":"`+root.Data.Comment.ID+`"}`, true)
	if w.Code != 201 {
		t.Fatalf("reply: %d %s", w.Code, w.Body)
	}
	var reply struct {
		Data struct {
			Comment struct {
				ID string `json:"id"`
			} `json:"comment"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &reply); err != nil {
		t.Fatal(err)
	}
	w = call(http.MethodPost, postPath+"/comments", `{"content":"nested","parentId":"`+reply.Data.Comment.ID+`"}`, true)
	if w.Code != 422 {
		t.Fatalf("nested reply: %d %s", w.Code, w.Body)
	}
	w = call(http.MethodPatch, "/api/v1/forum/comments/"+reply.Data.Comment.ID, `{"content":"edited reply"}`, true)
	if w.Code != 200 || !strings.Contains(w.Body.String(), "edited reply") {
		t.Fatalf("update comment: %d %s", w.Code, w.Body)
	}
	w = call(http.MethodDelete, "/api/v1/forum/comments/"+root.Data.Comment.ID, "", true)
	if w.Code != 204 {
		t.Fatalf("delete comment: %d %s", w.Code, w.Body)
	}
	w = call(http.MethodGet, postPath+"/comments", "", false)
	if w.Code != 200 || !bytes.Contains(w.Body.Bytes(), []byte(`"content":null`)) {
		t.Fatalf("tombstone: %d %s", w.Code, w.Body)
	}
	w = call(http.MethodDelete, "/api/v1/forum/comments/"+reply.Data.Comment.ID, "", true)
	if w.Code != 204 {
		t.Fatalf("delete reply: %d %s", w.Code, w.Body)
	}
	w = call(http.MethodGet, postPath+"/comments", "", false)
	if w.Code != 200 || !bytes.Contains(w.Body.Bytes(), []byte(`"total":0`)) || !bytes.Contains(w.Body.Bytes(), []byte(`"items":[]`)) {
		t.Fatalf("deleted leaves: %d %s", w.Code, w.Body)
	}
	path := "/api/v1/forum/posts/" + result.Data.Post.ID + "/like"
	for i := 0; i < 2; i++ {
		w = call(http.MethodPut, path, "", true)
		if w.Code != 200 || !bytes.Contains(w.Body.Bytes(), []byte(`"likeCount":1`)) {
			t.Fatalf("like: %d %s", w.Code, w.Body)
		}
	}
	w = call(http.MethodDelete, path, "", true)
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"liked":false`) {
		t.Fatalf("unlike: %d %s", w.Code, w.Body)
	}
	w = call(http.MethodPost, "/api/v1/forum/posts", `{"title":"x","contentMd":"x","postType":"solution"}`, true)
	if w.Code != 403 {
		t.Fatalf("unsolved solution: %d %s", w.Code, w.Body)
	}
	w = call(http.MethodDelete, "/api/v1/forum/posts/"+result.Data.Post.ID, "", false)
	if w.Code != 401 {
		t.Fatalf("anonymous delete: %d %s", w.Code, w.Body)
	}
	w = call(http.MethodDelete, postPath, "", true)
	if w.Code != 204 {
		t.Fatalf("delete post: %d %s", w.Code, w.Body)
	}
	w = call(http.MethodGet, postPath, "", true)
	if w.Code != 404 {
		t.Fatalf("deleted post visible: %d %s", w.Code, w.Body)
	}
}

func TestForumAdminModeration(t *testing.T) {
	if path := os.Getenv("FORUM_ADMIN_CONTRACT_CAPTURE"); path != "" {
		t.Setenv("CATALOG_CONTRACT_CAPTURE", path)
	} else {
		t.Setenv("CATALOG_CONTRACT_CAPTURE", "")
	}
	f := testutil.Catalog(t)
	t.Cleanup(func() {
		if _, err := f.Pool.Exec(context.Background(), `DELETE FROM forum_posts WHERE user_id=$1`, f.UserID); err != nil {
			t.Error(err)
		}
	})
	if _, err := f.Pool.Exec(context.Background(), `UPDATE users SET role='admin' WHERE id=$1`, f.UserID); err != nil {
		t.Fatal(err)
	}
	var postID, commentID string
	if err := f.Pool.QueryRow(context.Background(), `INSERT INTO forum_posts(user_id,post_type,title,content_md) VALUES($1,'discussion','Example','Content') RETURNING id`, f.UserID).Scan(&postID); err != nil {
		t.Fatal(err)
	}
	if err := f.Pool.QueryRow(context.Background(), `INSERT INTO forum_comments(post_id,user_id,content) VALUES($1,$2,'Comment') RETURNING id`, postID, f.UserID).Scan(&commentID); err != nil {
		t.Fatal(err)
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"userId": f.UserID, "sub": f.UserID, "exp": time.Now().Add(time.Hour).Unix()}).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	h := New(f.Pool, "secret")
	for _, tc := range []struct{ method, path, body string }{
		{"GET", "/api/v1/admin/forum/posts", ""},
		{"GET", "/api/v1/admin/forum/posts/" + postID, ""},
		{"PATCH", "/api/v1/admin/forum/posts/" + postID, `{"status":"hidden"}`},
		{"GET", "/api/v1/admin/forum/posts/" + postID + "/comments", ""},
		{"PATCH", "/api/v1/admin/forum/comments/" + commentID, `{"isDeleted":true}`},
	} {
		r := httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.body))
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatalf("%s %s: %d %s", tc.method, tc.path, w.Code, w.Body)
		}
		testutil.Capture(t, w, tc.method, r.Pattern)
	}
}
