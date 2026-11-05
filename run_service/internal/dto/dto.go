package dto

import (
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
	ProblemID  int    `json:"problemId"`
	LanguageID int    `json:"languageId"`
	Code       string `json:"code"`
}

// Validate returns a field name and an error message string if the request is invalid, or empty strings if valid.
func (r *RunRequest) Validate() (string, string) {
	if r.ProblemID <= 0 {
		return "problemId", "problemId must be greater than 0"
	}
	if r.LanguageID <= 0 {
		return "languageId", "languageId must be greater than 0"
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
	TestCaseID     int     `json:"testCaseId"`
	Status         string  `json:"status"`
	InputData      string  `json:"inputData"`
	ExpectedOutput string  `json:"expectedOutput"`
	ActualOutput   string  `json:"actualOutput"`
	RuntimeMs      *int    `json:"runtimeMs"`
	ErrorMessage   *string `json:"errorMessage,omitempty"`
	ErrorType      *string `json:"errorType,omitempty"`
	Stderr         *string `json:"stderr,omitempty"`
}

// RunResponse is the JSON body returned by POST /run.
type RunResponse struct {
	Status  string       `json:"status"`
	Passed  int          `json:"passed"`
	Total   int          `json:"total"`
	Results []TestResult `json:"results"`
}
