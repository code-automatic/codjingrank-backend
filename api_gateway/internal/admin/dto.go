package admin

import (
	"encoding/json"

	"innogen-backend/shared/models"
)

// --- Languages ---
type CreateLanguageRequest struct {
	Name            string  `json:"name"`
	PistonAlias     string  `json:"pistonAlias"`
	PistonVersion   string  `json:"pistonVersion"`
	FileExtension   *string `json:"fileExtension"`
	DefaultFileName *string `json:"defaultFileName"`
	IsActive        *bool   `json:"isActive"`
}

type UpdateLanguageRequest struct {
	Name            *string `json:"name"`
	PistonAlias     *string `json:"pistonAlias"`
	PistonVersion   *string `json:"pistonVersion"`
	FileExtension   *string `json:"fileExtension"`
	DefaultFileName *string `json:"defaultFileName"`
	IsActive        *bool   `json:"isActive"`
}

// --- Problems ---
type CreateProblemRequest struct {
	Slug             string          `json:"slug"`
	Title            string          `json:"title"`
	Difficulty       string          `json:"difficulty"`
	ProblemMD        string          `json:"problemMd"`
	TimeLimitMs      *int            `json:"timeLimitMs"`
	MemoryLimitMb    *int            `json:"memoryLimitMb"`
	IsPublished      *bool           `json:"isPublished"`
	ExecutionMode    *string         `json:"executionMode"`
	FunctionName     *string         `json:"functionName"`
	InitialCode      *string         `json:"initialCode"`
	DriverCode       *string         `json:"driverCode"`
	SolutionFileName *string         `json:"solutionFileName"`
	SampleTestCases  json.RawMessage `json:"sampleTestCases"`
}

type UpdateProblemRequest struct {
	Slug             *string         `json:"slug"`
	Title            *string         `json:"title"`
	Difficulty       *string         `json:"difficulty"`
	ProblemMD        *string         `json:"problemMd"`
	TimeLimitMs      *int            `json:"timeLimitMs"`
	MemoryLimitMb    *int            `json:"memoryLimitMb"`
	IsPublished      *bool           `json:"isPublished"`
	ExecutionMode    *string         `json:"executionMode"`
	FunctionName     *string         `json:"functionName"`
	InitialCode      *string         `json:"initialCode"`
	DriverCode       *string         `json:"driverCode"`
	SolutionFileName *string         `json:"solutionFileName"`
	SampleTestCases  json.RawMessage `json:"sampleTestCases"`
}

// --- Test Cases ---
type CreateTestCaseRequest struct {
	Visibility     string  `json:"visibility"`
	InputData      *string `json:"inputData"`
	ExpectedOutput string  `json:"expectedOutput"`
	ExecuteCode    *string `json:"executeCode"`
	OrderIndex     int     `json:"orderIndex"`
}

type UpdateTestCaseRequest struct {
	Visibility     *string `json:"visibility"`
	InputData      *string `json:"inputData"`
	ExpectedOutput *string `json:"expectedOutput"`
	ExecuteCode    *string `json:"executeCode"`
	OrderIndex     *int    `json:"orderIndex"`
}

// --- Tags ---
type CreateTagRequest struct {
	Name string `json:"name"`
}

// --- Problem-Tags ---
type CreateProblemTagRequest struct {
	TagID int `json:"tagId"`
}

// --- Pagination ---
type ProblemListResponse struct {
	Problems []models.Problem `json:"problems"`
	Page     int              `json:"page"`
	Limit    int              `json:"limit"`
	Total    int              `json:"total"`
}
