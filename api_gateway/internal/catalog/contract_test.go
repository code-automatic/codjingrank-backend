package catalog

import (
	"context"
	"encoding/json"
	jwt "github.com/golang-jwt/jwt/v5"
	"innogen-backend/shared/response"
	"innogen-backend/shared/testutil"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCatalogPublishedHTTPContract(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.Background()
	var slug, pkg string
	if err := f.Pool.QueryRow(ctx, `SELECT p.slug,k.slug FROM problems p JOIN levels l ON l.id=p.level_id JOIN packages k ON k.id=l.package_id WHERE p.id=$1`, f.ProblemID).Scan(&slug, &pkg); err != nil {
		t.Fatal(err)
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"userId": f.UserID, "sub": f.UserID, "exp": time.Now().Add(time.Hour).Unix()}).SignedString([]byte("test-secret"))
	mux := http.NewServeMux()
	Register(mux, f.Pool, "test-secret", slog.New(slog.NewTextHandler(io.Discard, nil)))
	for _, path := range []string{"/packages/" + pkg, "/packages/" + pkg + "/levels", "/problems?package=" + pkg, "/problems/" + slug, "/problems/" + slug + "/languages", "/problems/" + slug + "/samples", "/problems/" + slug + "/template?languageId=" + f.LanguageID} {
		req := httptest.NewRequest("GET", "/api/v1"+path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		response.Router(mux).ServeHTTP(rec, req)
		testutil.Capture(t, rec, req.Method, req.Pattern)
		var env map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
			t.Fatal(err)
		}
		if rec.Code != 200 || env["status"] != "success" || env["code"] != float64(200) || len(env) != 4 {
			t.Fatalf("%s: %d %s", path, rec.Code, rec.Body.String())
		}
		for _, secret := range []string{"HIDDEN INPUT", "wrapperCode", "pistonAlias", "password_hash"} {
			if strings.Contains(rec.Body.String(), secret) {
				t.Fatalf("%s leaked %s", path, secret)
			}
		}
	}
}
