package admin

import (
	"context"
	"encoding/json"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestV1CatalogLifecycle(t *testing.T) {
	url := os.Getenv("CATALOG_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set disposable database URL")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	var uid string
	if err = pool.QueryRow(ctx, `INSERT INTO users(email,username,password_hash,role) VALUES('admin-v1@example.org','admin-v1','hash','admin') RETURNING id`).Scan(&uid); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, uid) })
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": uid, "userId": uid, "role": "admin", "exp": time.Now().Add(time.Hour).Unix()}).SignedString([]byte("test-secret"))
	mux := http.NewServeMux()
	RegisterV1(mux, pool, "test-secret", slog.New(slog.NewTextHandler(os.Stderr, nil)))
	call := func(method, path, body string, want int) map[string]any {
		t.Helper()
		req := httptest.NewRequest(method, "/api/v1/admin"+path, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("%s %s: %d %s", method, path, rec.Code, rec.Body.String())
		}
		out := map[string]any{}
		json.Unmarshal(rec.Body.Bytes(), &out)
		return out
	}
	call("POST", "/packages", `{"slug":"bad","name":"Bad","unknown":true}`, 400)
	p := call("POST", "/packages", `{"slug":"admin-lifecycle","name":"Lifecycle"}`, 201)["data"].(map[string]any)["package"].(map[string]any)
	id := p["id"].(string)
	t.Cleanup(func() {
		for _, q := range []string{`UPDATE problems SET status='draft',published_at=NULL WHERE level_id IN(SELECT id FROM levels WHERE package_id=$1)`, `DELETE FROM problem_test_cases WHERE problem_id IN(SELECT id FROM problems WHERE level_id IN(SELECT id FROM levels WHERE package_id=$1))`, `DELETE FROM problems WHERE level_id IN(SELECT id FROM levels WHERE package_id=$1)`, `DELETE FROM levels WHERE package_id=$1`, `DELETE FROM packages WHERE id=$1`} {
			if _, e := pool.Exec(ctx, q, id); e != nil {
				t.Error(e)
			}
		}
	})
	levels := call("GET", "/packages/"+id+"/levels", "", 200)["data"].(map[string]any)["items"].([]any)
	if len(levels) != 3 {
		t.Fatalf("wrong levels: %v", levels)
	}
	level := levels[0].(map[string]any)["id"].(string)
	problem := call("POST", "/problems", `{"levelId":"`+level+`","slug":"admin-lifecycle","title":"Problem","statementMd":"Statement","executionMode":"stdin_stdout","timeLimitMs":1000,"memoryLimitMb":128,"outputLimitBytes":1024}`, 201)["data"].(map[string]any)["problem"].(map[string]any)
	pid := problem["id"].(string)
	call("POST", "/problems/"+pid+"/publish", "", 422)
	var lang string
	pool.QueryRow(ctx, `SELECT id FROM languages WHERE is_active LIMIT 1`).Scan(&lang)
	call("PUT", "/problems/"+pid+"/languages", `{"languageIds":["`+lang+`"]}`, 200)
	call("PUT", "/problems/"+pid+"/templates/"+lang, `{"starterCode":"print(1)","wrapperCode":null,"fileName":"main.py"}`, 200)
	call("POST", "/problems/"+pid+"/samples", `{"expectedOutput":"1","displayOrder":1}`, 201)
	call("POST", "/problems/"+pid+"/test-cases", `{"expectedOutput":"1","displayOrder":1,"isActive":true}`, 201)
	call("POST", "/problems/"+pid+"/publish", "", 200)
	clone := call("POST", "/problems/"+pid+"/clone", `{"slug":"lifecycle-clone","title":"Clone"}`, 201)["data"].(map[string]any)["problem"].(map[string]any)
	cloneID := clone["id"].(string)
	call("POST", "/problems/"+cloneID+"/publish", "", 200)
	call("GET", "/problems/"+cloneID+"/samples", "", 200)
	call("POST", "/tags", `{"slug":"lifecycle-tag","name":"Tag"}`, 201)
	t.Cleanup(func() { pool.Exec(ctx, `DELETE FROM tags WHERE slug='lifecycle-tag'`) })
	call("GET", "/languages", "", 200)
	call("PATCH", "/problems/"+pid, `{"timeLimitMs":2000}`, 409)
	call("PATCH", "/problems/"+pid, `{"title":"Editorial correction"}`, 200)
	call("POST", "/problems/"+pid+"/archive", "", 200)
	call("PUT", "/problems/"+pid+"/languages", `{"languageIds":[]}`, 409)
	call("DELETE", "/problems/"+pid, "", 409)
	call("POST", "/packages/"+id+"/publish", "", 200)
	call("PATCH", "/packages/"+id, `{"description":null}`, 200)
	if _, err = pool.Exec(ctx, `UPDATE users SET role='student' WHERE id=$1`, uid); err != nil {
		t.Fatal(err)
	}
	call("GET", "/packages", "", 403)
}
