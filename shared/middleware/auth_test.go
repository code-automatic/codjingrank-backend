package middleware

import (
	jwt "github.com/golang-jwt/jwt/v5"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthUUIDBoundary(t *testing.T) {
	const id = "b1000000-0000-4000-8000-000000000001"
	for _, tc := range []struct {
		name   string
		uid    any
		exp    bool
		method jwt.SigningMethod
		want   int
	}{
		{"uuid", id, true, jwt.SigningMethodHS256, 204},
		{"legacy", 1, true, jwt.SigningMethodHS256, 401},
		{"invalid", "bad", true, jwt.SigningMethodHS256, 401},
		{"missing expiry", id, false, jwt.SigningMethodHS256, 401},
		{"other algorithm", id, true, jwt.SigningMethodHS512, 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			claims := jwt.MapClaims{"userId": tc.uid, "sub": id, "role": "student"}
			if tc.exp {
				claims["exp"] = time.Now().Add(time.Hour).Unix()
			}
			token, err := jwt.NewWithClaims(tc.method, claims).SignedString([]byte("test-secret"))
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()
			Auth("test-secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got, ok := GetUserID(r)
				if !ok || any(got) != id {
					t.Errorf("wrong identity: %v", got)
				}
				w.WriteHeader(204)
			})).ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("got %d want %d", rec.Code, tc.want)
			}
		})
	}
}
