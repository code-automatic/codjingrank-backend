package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"innogen-backend/shared/response"
)

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	response.Success(w, http.StatusCreated, map[string]string{"foo": "bar"}, "Created successfully")

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var env response.SuccessEnvelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if env.Code != http.StatusCreated {
		t.Errorf("Expected body.code %d, got %d", http.StatusCreated, env.Code)
	}
	if env.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", env.Status)
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	response.Error(w, http.StatusNotFound, "Not found", "NOT_FOUND")

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}

	var env response.ErrorEnvelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if env.Code != http.StatusNotFound {
		t.Errorf("Expected body.code %d, got %d", http.StatusNotFound, env.Code)
	}
	if env.Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", env.Status)
	}
}

func TestErrorValidation(t *testing.T) {
	w := httptest.NewRecorder()
	response.ErrorValidation(w, "Invalid input", "email")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	var env response.ErrorEnvelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if env.Code != http.StatusBadRequest {
		t.Errorf("Expected body.code %d, got %d", http.StatusBadRequest, env.Code)
	}
	if env.Error != "VALIDATION_ERROR" {
		t.Errorf("Expected machine code 'VALIDATION_ERROR', got '%s'", env.Error)
	}
}

func TestErrorSimpleUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	response.ErrorSimple(w, http.StatusUnauthorized, "Unauthorized")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var env response.ErrorEnvelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if env.Code != http.StatusUnauthorized {
		t.Errorf("Expected body.code %d, got %d", http.StatusUnauthorized, env.Code)
	}
	if env.Error != "UNAUTHORIZED" {
		t.Errorf("Expected machine code 'UNAUTHORIZED', got '%s'", env.Error)
	}
}

func TestNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	response.NoContent(w)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, w.Code)
	}

	if w.Body.Len() > 0 {
		t.Errorf("Expected empty body, got %s", w.Body.String())
	}
}
