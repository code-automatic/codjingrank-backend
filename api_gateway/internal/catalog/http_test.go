package catalog

import (
	"context"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestCatalogHTTPBoundary(t *testing.T) {
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
	var id string
	if err = pool.QueryRow(ctx, `INSERT INTO users(email,username,password_hash) VALUES('http-catalog@example.org','http-catalog','hash') RETURNING id`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, id) })
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"userId": id, "sub": id, "exp": time.Now().Add(time.Hour).Unix()}).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	Register(mux, pool, "test-secret", slog.New(slog.NewTextHandler(io.Discard, nil)))
	for _, tc := range []struct {
		url      string
		auth     bool
		code     int
		fragment string
	}{
		{"/api/v1/languages", false, 200, `"items":[`},
		{"/api/v1/tags", false, 200, `"items":[`},
		{"/api/v1/packages", false, 401, `"error":"UNAUTHORIZED"`},
		{"/api/v1/packages?page=0", true, 400, `"status":"error"`},
		{"/api/v1/packages?size=101", true, 400, `"status":"error"`},
		{"/api/v1/packages", true, 200, `"items":[]`},
		{"/api/v1/problems?solved=maybe", true, 400, `"status":"error"`},
		{"/api/v1/problems?difficulty=wrong", true, 400, `"status":"error"`},
		{"/api/v1/problems/missing", true, 404, `"error":"NOT_FOUND"`},
		{"/api/v1/problems/missing/template?languageId=bad", true, 400, `"status":"error"`},
	} {
		t.Run(tc.url, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.url, nil)
			if tc.auth {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != tc.code || !strings.Contains(rec.Body.String(), tc.fragment) {
				t.Fatalf("%d %s", rec.Code, rec.Body.String())
			}
		})
	}
	if _, err = pool.Exec(ctx, `UPDATE users SET is_active=false WHERE id=$1`, id); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/api/v1/packages", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 403 {
		t.Fatalf("inactive account: %d %s", rec.Code, rec.Body.String())
	}
}
