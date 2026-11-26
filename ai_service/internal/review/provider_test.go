package review

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// openAIEnvelope wraps content in an OpenAI-compatible chat completion response.
func openAIEnvelope(content string) string {
	payload, err := json.Marshal(map[string]any{
		"choices": []any{map[string]any{"message": map[string]any{"content": content}}},
	})
	if err != nil {
		panic(err)
	}
	return string(payload)
}

func TestProviderReturnsValidatedReview(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Error("missing provider token")
			return
		}
		var body struct {
			Model          string  `json:"model"`
			Temperature    float64 `json:"temperature"`
			ResponseFormat *struct {
				Type string `json:"type"`
			} `json:"response_format"`
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("invalid request: %v", err)
			return
		}
		if body.Model != "test-model" || body.Temperature != 0 {
			t.Errorf("unexpected model or temperature: %+v", body)
			return
		}
		if body.ResponseFormat == nil || body.ResponseFormat.Type != "json_object" {
			t.Errorf("missing json response format: %+v", body.ResponseFormat)
			return
		}
		if len(body.Messages) != 2 || body.Messages[0].Role != "system" || body.Messages[1].Role != "user" {
			t.Errorf("unexpected messages: %+v", body.Messages)
			return
		}
		if !strings.Contains(body.Messages[0].Content, "json") {
			t.Errorf("system prompt must mention json: %q", body.Messages[0].Content)
			return
		}
		user := body.Messages[1].Content
		for _, want := range []string{"print(1)", "Python", "Sum two numbers"} {
			if !strings.Contains(user, want) {
				t.Errorf("user message missing %q: %q", want, user)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(openAIEnvelope(`{"score":87.5,"summary":"Clear","dimensions":{"readability":85},"findings":[]}`)))
	}))
	defer s.Close()
	got, err := callProvider(context.Background(), s.URL, "token", "test-model", providerInput{Source: "print(1)", Language: "Python", Problem: "Sum two numbers"})
	if err != nil || got.Score == nil || *got.Score != 87.5 || got.Dimensions["readability"] != 85 {
		t.Fatalf("unexpected review: %+v %v", got, err)
	}
}

func TestProviderAcceptsFencedContent(t *testing.T) {
	content := "```json\n{\"score\":60,\"summary\":\"Fenced\",\"dimensions\":{\"maintainability\":70},\"findings\":[]}\n```"
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(openAIEnvelope(content)))
	}))
	defer s.Close()
	got, err := callProvider(context.Background(), s.URL, "", "test", providerInput{})
	if err != nil || got.Score == nil || *got.Score != 60 || got.Summary != "Fenced" {
		t.Fatalf("unexpected review: %+v %v", got, err)
	}
}

func TestProviderRejectsInvalidScore(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(openAIEnvelope(`{"score":101,"summary":"bad","dimensions":{},"findings":[]}`)))
	}))
	defer s.Close()
	_, err := callProvider(context.Background(), s.URL, "", "test", providerInput{})
	var permanent permanentProviderError
	if err == nil || !errors.As(err, &permanent) {
		t.Fatalf("expected permanent validation error, got %v", err)
	}
}

func TestProviderRejectsInvalidFinding(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(openAIEnvelope(`{"score":50,"summary":"ok","dimensions":{},"findings":[{"severity":"critical","title":"t","explanation":"e"}]}`)))
	}))
	defer s.Close()
	_, err := callProvider(context.Background(), s.URL, "", "test", providerInput{})
	var permanent permanentProviderError
	if err == nil || !errors.As(err, &permanent) {
		t.Fatalf("expected permanent validation error, got %v", err)
	}
}

func TestProviderRejectsMalformedResponse(t *testing.T) {
	for name, body := range map[string]string{
		"missing choices": `{"choices":[]}`,
		"empty content":   openAIEnvelope(""),
		"non-json text":   openAIEnvelope("not json"),
	} {
		t.Run(name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(body))
			}))
			defer s.Close()
			_, err := callProvider(context.Background(), s.URL, "", "test", providerInput{})
			var permanent permanentProviderError
			if err == nil || !errors.As(err, &permanent) {
				t.Fatalf("expected permanent error, got %v", err)
			}
		})
	}
}

func TestProviderClassifiesHTTPStatus(t *testing.T) {
	for _, tc := range []struct {
		status    int
		permanent bool
	}{
		{http.StatusTooManyRequests, false},
		{http.StatusInternalServerError, false},
		{http.StatusServiceUnavailable, false},
		{http.StatusBadRequest, true},
		{http.StatusUnauthorized, true},
	} {
		t.Run(http.StatusText(tc.status), func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
			}))
			defer s.Close()
			_, err := callProvider(context.Background(), s.URL, "", "test", providerInput{})
			var permanent permanentProviderError
			if err == nil {
				t.Fatal("expected provider error")
			}
			if got := errors.As(err, &permanent); got != tc.permanent {
				t.Fatalf("permanent=%v want=%v (err=%v)", got, tc.permanent, err)
			}
		})
	}
}
