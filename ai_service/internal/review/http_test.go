package review

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInternalEnqueueRequiresServiceSecret(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, nil, "jwt-secret", "service-secret")
	req := httptest.NewRequest(http.MethodPost, "/internal/reviews/11111111-1111-4111-8111-111111111111", nil)
	req.Header.Set("Authorization", "Bearer incorrect")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", w.Code)
	}
}
