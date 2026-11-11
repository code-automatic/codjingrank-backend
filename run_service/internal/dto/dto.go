package dto

import (
	"innogen-backend/shared/middleware"
	"strings"

	"innogen-backend/shared/constants"
)

// Status constants for run results.
const (
	StatusAccepted          = constants.StatusAccepted
	StatusWrongAnswer       = constants.StatusWrongAnswer
	StatusCompilationError  = constants.StatusCompilationError
	StatusRuntimeError      = constants.StatusRuntimeError
	StatusTimeLimitExceeded = constants.StatusTimeLimitExceeded
	StatusInternalError     = constants.StatusInternalError
)

// RunRequest is the JSON body for POST /run.
type RunRequest struct {
	ProblemID  string `json:"problemId"`
	LanguageID string `json:"languageId"`
	Code       string `json:"code"`
}

// Validate returns a field name and an error message string if the request is invalid, or empty strings if valid.
func (r *RunRequest) Validate() (string, string) {
	if !middleware.ValidUUID(r.ProblemID) {
		return "problemId", "Invalid problem ID"
	}
	if !middleware.ValidUUID(r.LanguageID) {
		return "languageId", "Invalid language ID"
	}
	if strings.TrimSpace(r.Code) == "" {
		return "code", "code must not be empty"
	}
	if len(r.Code) > 100*1024 {
		return "code", "code exceeds maximum allowed length (100KB)"
	}
	return "", ""
}

// TestResult represents the result of a single test case execution.
type TestResult struct {
	TestCaseID     string  `json:"testCaseId"`
	Status         string  `json:"status"`
	InputData      string  `json:"inputData"`
	ExpectedOutput string  `json:"expectedOutput"`
	ActualOutput   string  `json:"actualOutput"`
	RuntimeMs      *int    `json:"runtimeMs"`
	ErrorMessage   *string `json:"errorMessage"`
	ErrorType      *string `json:"errorType"`
	Stderr         *string `json:"stderr"`
}

// RunResponse is the JSON body returned by POST /run.
type RunResponse struct {
	Status  string       `json:"status"`
	Passed  int          `json:"passed"`
	Total   int          `json:"total"`
	Results []TestResult `json:"results"`
}
