package proxy

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type failingTransport struct{}

func (failingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("private upstream error")
}
func TestProxyFailureEnvelope(t *testing.T) {
	p, err := NewProxy("http://upstream.invalid", slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	p.Transport = failingTransport{}
	w := httptest.NewRecorder()
	p.ServeHTTP(w, httptest.NewRequest("POST", "/run", nil))
	var env map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if w.Code != 502 || env["status"] != "error" || env["code"] != float64(502) || env["error"] != "UPSTREAM_ERROR" || env["message"] != "Bad gateway" || len(env) != 5 {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
}
