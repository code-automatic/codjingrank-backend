package piston

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRejectMalformedAndOversizedExecutionResponses(t *testing.T) {
	for _, body := range []string{`{}`, `{"run":{}}`, `{"run":{"code":null}}`, `{"compile":{"code":0}}`, strings.Repeat("x", 4*1024*1024+1)} {
		engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(body)) }))
		_, err := NewClient(engine.URL).Execute(context.Background(), "python", "3.10.0", "main.py", "pass", "", 1000, 128)
		engine.Close()
		if err == nil {
			t.Fatal("invalid execution response accepted")
		}
	}
}

func TestCompileFailureWithoutDiagnosticsIsNeverSuccessful(t *testing.T) {
	engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{"compile":{"code":1}}`)) }))
	defer engine.Close()
	result, err := NewClient(engine.URL).Execute(context.Background(), "python", "3.10.0", "main.py", "pass", "", 1000, 128)
	if err != nil {
		t.Fatal(err)
	}
	if result.Compile.Stderr == "" {
		t.Fatal("missing compile failure signal to judge")
	}
}

func TestProviderTimeoutAndDiagnosticsAreSafe(t *testing.T) {
	engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "PRIVATE PROVIDER KEY", 500) }))
	defer engine.Close()
	_, err := NewClient(engine.URL).Execute(context.Background(), "python", "3.10.0", "main.py", "pass", "", 1000, 128)
	if err == nil || strings.Contains(err.Error(), "PRIVATE") {
		t.Fatalf("unsafe error %v", err)
	}
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(100 * time.Millisecond) }))
	defer slow.Close()
	client := NewClient(slow.URL)
	client.httpClient.Timeout = 10 * time.Millisecond
	_, err = client.Execute(context.Background(), "python", "3.10.0", "main.py", "pass", "", 1000, 128)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("lost timeout cause: %v", err)
	}
}
