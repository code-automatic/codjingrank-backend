package review

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProviderReturnsValidatedReview(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["source"] != "print(1)" || body["model"] != "test-model" {
			t.Errorf("invalid request: %v, %v", body, err)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Error("missing provider token")
		}
		w.Write([]byte(`{"score":87.5,"summary":"Clear","dimensions":{"readability":85},"findings":[]}`))
	}))
	defer s.Close()
	got, err := callProvider(context.Background(), s.URL, "token", "test-model", providerInput{Source: "print(1)"})
	if err != nil || got.Score == nil || *got.Score != 87.5 || got.Dimensions["readability"] != 85 {
		t.Fatalf("unexpected review: %+v %v", got, err)
	}
}

func TestProviderRejectsInvalidScore(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"score":101,"summary":"bad","dimensions":{},"findings":[]}`))
	}))
	defer s.Close()
	_, err := callProvider(context.Background(), s.URL, "", "test", providerInput{})
	if err == nil {
		t.Fatal("accepted out-of-range score")
	}
}
