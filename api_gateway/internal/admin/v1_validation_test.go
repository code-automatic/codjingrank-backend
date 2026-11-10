package admin

import (
	"context"
	jwt "github.com/golang-jwt/jwt/v5"
	"innogen-backend/shared/testutil"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPatchRejectsOversizedDescription(t *testing.T) {
	r := httptest.NewRequest("PATCH", "/api/v1/admin/packages/id", strings.NewReader(`{"description":"`+strings.Repeat("x", 10001)+`"}`))
	w := httptest.NewRecorder()
	_, _, _, ok := patchFields(w, r, map[string]string{"description": "description"})
	if ok || w.Code != 400 {
		t.Fatalf("oversized description accepted: %v %d", ok, w.Code)
	}
}

func TestAdminFunctionRequiresUsableWrapper(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.Background()
	if _, err := f.Pool.Exec(ctx, `UPDATE problems SET status='draft',published_at=NULL,execution_mode='function',function_name='solve' WHERE id=$1`, f.ProblemID); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Pool.Exec(ctx, `UPDATE users SET role='admin' WHERE id=$1`, f.UserID); err != nil {
		t.Fatal(err)
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"userId": f.UserID, "sub": f.UserID, "exp": time.Now().Add(time.Hour).Unix()}).SignedString([]byte("test-secret"))
	mux := http.NewServeMux()
	RegisterV1(mux, f.Pool, "test-secret", slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := httptest.NewRequest("POST", "/api/v1/admin/problems/"+f.ProblemID+"/publish", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 409 {
		t.Fatalf("function without wrapper published: %d %s", rec.Code, rec.Body.String())
	}
}
