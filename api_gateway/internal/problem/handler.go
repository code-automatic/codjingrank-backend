package problem

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"innogen-backend/shared/response"
)

// Handler serves problem-related HTTP endpoints.
type Handler struct {
	repo *ProblemRepository
	log  *slog.Logger
}

// NewHandler creates a new Handler.
func NewHandler(repo *ProblemRepository, log *slog.Logger) *Handler {
	return &Handler{repo: repo, log: log}
}

// ProblemResponse is the top-level response for a single problem.
type ProblemResponse struct {
	Problem ProblemDetail `json:"problem"`
}

// ProblemDetail contains the full problem data returned to the client.
type ProblemDetail struct {
	ID              int             `json:"id"`
	Slug            string          `json:"slug"`
	Title           string          `json:"title"`
	Difficulty      string          `json:"difficulty"`
	ProblemMd       string          `json:"problemMd"`
	TimeLimitMs     int             `json:"timeLimitMs"`
	MemoryLimitMb   int             `json:"memoryLimitMb"`
	AcceptanceRate  float64         `json:"acceptanceRate"`
	ExecutionMode   string          `json:"executionMode"`
	FunctionName    *string         `json:"functionName"`
	InitialCode     *string         `json:"initialCode"`
	SampleTestCases json.RawMessage `json:"sampleTestCases"`
}

// TestCaseResponse is the response DTO for a single test case.
type TestCaseResponse struct {
	ID             int     `json:"id"`
	ProblemID      int     `json:"problemId"`
	InputData      *string `json:"inputData"`
	ExpectedOutput string  `json:"expectedOutput"`
	OrderIndex     int     `json:"orderIndex"`
}

// TestCaseListResponse is the top-level response for a list of test cases.
type TestCaseListResponse struct {
	TestCases []TestCaseResponse `json:"testCases"`
}

// GetProblem handles GET /problems/{slug}.
func (h *Handler) GetProblem(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		response.ErrorSimple(w, http.StatusBadRequest, "slug is required")
		return
	}

	problem, err := h.repo.FindBySlug(r.Context(), slug)
	if err != nil {
		h.log.Error("failed to find problem by slug", "slug", slug, "error", err)
		response.ErrorSimple(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if problem == nil {
		response.ErrorSimple(w, http.StatusNotFound, "Problem not found")
		return
	}

	testCases, err := h.repo.FindTestCasesByProblemID(r.Context(), problem.ID, "sample")
	if err != nil {
		h.log.Error("failed to find sample test cases", "problemID", problem.ID, "error", err)
		response.ErrorSimple(w, http.StatusInternalServerError, "internal server error")
		return
	}

	type sampleTestCase struct {
		InputData      json.RawMessage `json:"inputData"`
		ExpectedOutput string          `json:"expectedOutput"`
	}

	samples := make([]sampleTestCase, 0, len(testCases))
	for _, tc := range testCases {
		var inputData json.RawMessage
		if tc.InputData != nil {
			inputData = json.RawMessage(*tc.InputData)
		}
		samples = append(samples, sampleTestCase{
			InputData:      inputData,
			ExpectedOutput: tc.ExpectedOutput,
		})
	}

	sampleBytes, _ := json.Marshal(samples)

	resp := ProblemResponse{
		Problem: ProblemDetail{
			ID:              problem.ID,
			Slug:            problem.Slug,
			Title:           problem.Title,
			Difficulty:      problem.Difficulty,
			ProblemMd:       problem.ProblemMD,
			TimeLimitMs:     problem.TimeLimitMs,
			MemoryLimitMb:   problem.MemoryLimitMb,
			AcceptanceRate:  problem.AcceptanceRate,
			ExecutionMode:   problem.ExecutionMode,
			FunctionName:    problem.FunctionName,
			InitialCode:     problem.InitialCode,
			SampleTestCases: sampleBytes,
		},
	}

	response.Success(w, http.StatusOK, resp, "Problem detail retrieved successfully")
}

// ListTestCases handles GET /problems/{id}/test-cases?visibility=sample.
func (h *Handler) ListTestCases(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.ErrorSimple(w, http.StatusBadRequest, "invalid problem id")
		return
	}

	visibility := r.URL.Query().Get("visibility")
	if visibility != "sample" {
		visibility = "sample"
	}

	testCases, err := h.repo.FindTestCasesByProblemID(r.Context(), id, visibility)
	if err != nil {
		h.log.Error("failed to find test cases", "problemId", id, "error", err)
		response.ErrorSimple(w, http.StatusInternalServerError, "internal server error")
		return
	}

	tcResponses := make([]TestCaseResponse, 0, len(testCases))
	for _, tc := range testCases {
		tcResponses = append(tcResponses, TestCaseResponse{
			ID:             tc.ID,
			ProblemID:      tc.ProblemID,
			InputData:      tc.InputData,
			ExpectedOutput: tc.ExpectedOutput,
			OrderIndex:     tc.OrderIndex,
		})
	}

	resp := TestCaseListResponse{
		TestCases: tcResponses,
	}

	response.Success(w, http.StatusOK, resp, "Sample test cases retrieved successfully")
}

// DailyChallengeResponse is the response DTO for the daily challenge.
type DailyChallengeResponse struct {
	Date    string        `json:"date"`
	Problem ProblemDetail `json:"problem"`
}

// GetDailyChallenge handles GET /problems/daily-challenge.
func (h *Handler) GetDailyChallenge(w http.ResponseWriter, r *http.Request) {
	problem, err := h.repo.GetDailyChallenge(r.Context())
	if err != nil {
		h.log.Error("failed to get daily challenge", "error", err)
		response.ErrorSimple(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if problem == nil {
		response.ErrorSimple(w, http.StatusNotFound, "No daily challenge available")
		return
	}

	testCases, err := h.repo.FindTestCasesByProblemID(r.Context(), problem.ID, "sample")
	if err != nil {
		h.log.Error("failed to find sample test cases for daily challenge", "problemID", problem.ID, "error", err)
		response.ErrorSimple(w, http.StatusInternalServerError, "internal server error")
		return
	}

	type sampleTestCase struct {
		InputData      json.RawMessage `json:"inputData"`
		ExpectedOutput string          `json:"expectedOutput"`
	}

	samples := make([]sampleTestCase, 0, len(testCases))
	for _, tc := range testCases {
		var inputData json.RawMessage
		if tc.InputData != nil {
			inputData = json.RawMessage(*tc.InputData)
		}
		samples = append(samples, sampleTestCase{
			InputData:      inputData,
			ExpectedOutput: tc.ExpectedOutput,
		})
	}

	sampleBytes, _ := json.Marshal(samples)

	resp := DailyChallengeResponse{
		Date: time.Now().UTC().Format(time.DateOnly),
		Problem: ProblemDetail{
			ID:              problem.ID,
			Slug:            problem.Slug,
			Title:           problem.Title,
			Difficulty:      problem.Difficulty,
			ProblemMd:       problem.ProblemMD,
			TimeLimitMs:     problem.TimeLimitMs,
			MemoryLimitMb:   problem.MemoryLimitMb,
			AcceptanceRate:  problem.AcceptanceRate,
			ExecutionMode:   problem.ExecutionMode,
			FunctionName:    problem.FunctionName,
			InitialCode:     problem.InitialCode,
			SampleTestCases: sampleBytes,
		},
	}

	response.Success(w, http.StatusOK, resp, "Daily challenge retrieved successfully")
}
