package testutil

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// Capture records test responses only when explicitly requested for OpenAPI validation.
func Capture(t *testing.T, r *httptest.ResponseRecorder, method, pattern string) {
	t.Helper()
	path := os.Getenv("CATALOG_CONTRACT_CAPTURE")
	if path == "" || r.Code >= 400 || r.Code == 204 {
		return
	}
	_, pattern, _ = strings.Cut(pattern, " ")
	if pattern == "" {
		t.Fatal("missing registered request pattern")
	}
	var body any
	if err := json.Unmarshal(r.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(map[string]any{"method": method, "pattern": pattern, "status": r.Code, "body": body}); err != nil {
		t.Fatal(err)
	}
}
