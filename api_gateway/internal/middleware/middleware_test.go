package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSAllowsCatalogPatch(t *testing.T) {
	req := httptest.NewRequest("OPTIONS", "/api/v1/admin/packages/id", nil)
	req.Header.Set("Origin", "https://frontend.example.org")
	req.Header.Set("Access-Control-Request-Method", "PATCH")
	rec := httptest.NewRecorder()
	CORS("https://frontend.example.org")(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(rec, req)
	if !strings.Contains(rec.Header().Get("Access-Control-Allow-Methods"), "PATCH") {
		t.Fatalf("PATCH not allowed: %v", rec.Header())
	}
}
