package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"innogen-backend/shared/response"
)

func TestRouterContract(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		response.Success(w, 200, map[string]string{"id": r.PathValue("id")}, "Item retrieved")
	})
	mux.HandleFunc("GET /docs/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html>docs</html>"))
	})
	handler := response.Router(mux)
	for _, tc := range []struct {
		method, path   string
		code           int
		machine, allow string
	}{
		{"GET", "/missing", 404, "NOT_FOUND", ""}, {"POST", "/items/1", 405, "METHOD_NOT_ALLOWED", "GET, HEAD"},
		{"HEAD", "/missing", 404, "", ""}, {"HEAD", "/items/1", 200, "", ""},
		{"GET", "/items/abc", 200, "", ""}, {"GET", "/docs/", 200, "", ""}, {"GET", "/docs", 301, "", ""},
		{"GET", "/items//abc", 301, "", ""},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
			expectedCode := tc.code
			if tc.code == 301 {
				baseline := httptest.NewRecorder()
				mux.ServeHTTP(baseline, httptest.NewRequest(tc.method, tc.path, nil))
				expectedCode = baseline.Code
			}
			if w.Code != expectedCode {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			if w.Header().Get("Allow") != tc.allow {
				t.Errorf("Allow=%q", w.Header().Get("Allow"))
			}
			if tc.machine != "" {
				var env map[string]any
				if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
					t.Fatal(err)
				}
				if len(env) != 5 || env["status"] != "error" || env["code"] != float64(tc.code) || env["error"] != tc.machine {
					t.Fatalf("envelope=%v", env)
				}
				details, ok := env["details"]
				if !ok || details != nil {
					t.Fatal("details must be present and null")
				}
				if w.Header().Get("Content-Type") != "application/json" {
					t.Fatal("not JSON")
				}
			}
			if tc.method == "HEAD" && w.Body.Len() != 0 {
				t.Fatalf("HEAD body=%s", w.Body.String())
			}
			if tc.path == "/items/abc" {
				var env struct{ Data struct{ ID string } }
				if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil || env.Data.ID != "abc" {
					t.Fatalf("PathValue lost: %s", w.Body.String())
				}
			}
			if tc.path == "/docs/" && w.Body.String() != "<html>docs</html>" {
				t.Fatal("documentation changed")
			}
			if tc.code == 301 && w.Header().Get("Location") == "" {
				t.Fatal("redirect lost")
			}
		})
	}
}

func TestSuccessNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	response.Success(w, 204, nil, "")
	if w.Code != 204 || w.Body.Len() != 0 {
		t.Fatalf("204 must have no body: %d %s", w.Code, w.Body.String())
	}
}

func TestEnvelopeFieldSets(t *testing.T) {
	for _, tc := range []struct {
		name   string
		write  func(http.ResponseWriter)
		keys   int
		status string
		code   int
	}{
		{"business verdict", func(w http.ResponseWriter) {
			response.Success(w, 200, map[string]string{"status": "Wrong Answer"}, "Run completed")
		}, 4, "success", 200},
		{"error", func(w http.ResponseWriter) { response.ErrorSimple(w, 500, "Internal server error") }, 5, "error", 500},
		{"validation", func(w http.ResponseWriter) { response.ErrorValidation(w, "Email is required", "email") }, 5, "error", 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tc.write(w)
			var env map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
				t.Fatal(err)
			}
			if len(env) != tc.keys || env["status"] != tc.status || env["code"] != float64(w.Code) || w.Code != tc.code {
				t.Fatalf("envelope=%v HTTP=%d", env, w.Code)
			}
			if _, ok := env["message"].(string); !ok {
				t.Fatal("missing message")
			}
			if tc.status == "error" {
				if _, ok := env["error"].(string); !ok {
					t.Fatal("error is not string")
				}
				if _, ok := env["details"]; !ok {
					t.Fatal("missing details")
				}
			} else {
				if env["data"].(map[string]any)["status"] != "Wrong Answer" {
					t.Fatal("business status changed")
				}
			}
		})
	}
}
