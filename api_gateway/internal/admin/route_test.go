package admin

import (
	"encoding/json"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"innogen-backend/shared/middleware"
	"innogen-backend/shared/response"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAdminRoutingEnvelope(t *testing.T) {
	mux := http.NewServeMux()
	RegisterAdminRoutes(mux, nil, "test-secret")
	token, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, middleware.Claims{UserID: 1, Role: "admin", RegisteredClaims: jwtlib.RegisteredClaims{ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour))}}).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		path, method string
		auth         bool
		code         int
		machine      string
	}{
		{"/admin/missing", "GET", true, 404, "NOT_FOUND"}, {"/admin/languages", "PATCH", true, 405, "METHOD_NOT_ALLOWED"}, {"/admin/missing", "GET", false, 401, "UNAUTHORIZED"},
	} {
		t.Run(tc.method+tc.path+tc.machine, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.auth {
				r.Header.Set("Authorization", "Bearer "+token)
			}
			w := httptest.NewRecorder()
			response.Router(mux).ServeHTTP(w, r)
			var env map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
				t.Fatal(err)
			}
			if w.Code != tc.code || env["code"] != float64(tc.code) || env["error"] != tc.machine || len(env) != 5 {
				t.Fatalf("%d %s", w.Code, w.Body.String())
			}
			if tc.code == 405 && w.Header().Get("Allow") == "" {
				t.Fatal("Allow missing")
			}
		})
	}
}
