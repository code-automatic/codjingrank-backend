package models

import (
	"encoding/json"
	"time"
)

// User represents a row in the users table.
// Password is tagged json:"-" to prevent accidental serialization in API responses.
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Username  *string   `json:"username"`
	FullName  *string   `json:"fullName"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Problem represents a row in the problems table.
type Problem struct {
	ID               string          `json:"id"`
	Slug             string          `json:"slug"`
	Title            string          `json:"title"`
	Difficulty       string          `json:"difficulty"`
	ProblemMD        string          `json:"problemMd"`
	TimeLimitMs      int             `json:"timeLimitMs"`
	MemoryLimitMb    int             `json:"memoryLimitMb"`
	AcceptanceRate   float64         `json:"acceptanceRate"`
	IsPublished      bool            `json:"isPublished"`
	ExecutionMode    string          `json:"executionMode"`
	FunctionName     *string         `json:"functionName"`
	InitialCode      *string         `json:"initialCode"`
	DriverCode       *string         `json:"driverCode"`
	SolutionFileName *string         `json:"solutionFileName"`
	SampleTestCases  json.RawMessage `json:"sampleTestCases"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
}

// TestCase represents a row in the test_cases table.
type TestCase struct {
	ID             string    `json:"id"`
	ProblemID      string    `json:"problemId"`
	Visibility     string    `json:"visibility"`
	InputData      *string   `json:"inputData"`
	ExpectedOutput string    `json:"expectedOutput"`
	ExecuteCode    *string   `json:"executeCode"`
	OrderIndex     int       `json:"orderIndex"`
	CreatedAt      time.Time `json:"createdAt"`
}

// Language represents a row in the languages table.
type Language struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	PistonAlias     string    `json:"pistonAlias"`
	PistonVersion   string    `json:"pistonVersion"`
	FileExtension   *string   `json:"fileExtension"`
	DefaultFileName *string   `json:"defaultFileName"`
	IsActive        bool      `json:"isActive"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// Tag represents a row in the tags table.
type Tag struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

// Submission represents a row in the submissions table.
type Submission struct {
	ID             string     `json:"id"`
	UserID         string     `json:"userId"`
	ProblemID      string     `json:"problemId"`
	LanguageID     string     `json:"languageId"`
	Code           string     `json:"code"`
	ProblemSlug    string     `json:"problemSlug"`
	ProblemTitle   string     `json:"problemTitle"`
	LanguageSlug   string     `json:"languageSlug"`
	Status         string     `json:"status"`
	RuntimeMs      *int       `json:"runtimeMs"`
	MemoryKb       *int       `json:"memoryKb"`
	ErrorMessage   *string    `json:"errorMessage"`
	PassCount      int        `json:"passCount"`
	TotalTestcases int        `json:"totalTestcases"`
	CreatedAt      time.Time  `json:"createdAt"`
	JudgedAt       *time.Time `json:"judgedAt"`
}
