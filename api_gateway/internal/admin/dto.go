package admin

import (
	"encoding/json"
	"time"

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

// --- Users (admin-safe projection) ---
type AdminUser struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	FullName  *string   `json:"fullName"`
	AvatarURL *string   `json:"avatarUrl"`
	Bio       *string   `json:"bio"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// UpdateUserRequest is a partial update. A nil field is not sent by the client.
type UpdateUserRequest struct {
	Role     *string `json:"role"`
	IsActive *bool   `json:"isActive"`
	FullName *string `json:"fullName"`
}

// --- Statistics ---
type AdminStats struct {
	Users               int64               `json:"users"`
	ActiveUsers         int64               `json:"activeUsers"`
	Packages            int64               `json:"packages"`
	PublishedProblems   int64               `json:"publishedProblems"`
	Submissions         int64               `json:"submissions"`
	AcceptedSubmissions int64               `json:"acceptedSubmissions"`
	UsersWithSolutions  int64               `json:"usersWithSolutions"`
	ForumPosts          int64               `json:"forumPosts"`
	MostSolvedProblems  []MostSolvedProblem `json:"mostSolvedProblems"`
}

// MostSolvedProblem is one dashboard row ordered by distinct solvers.
type MostSolvedProblem struct {
	ProblemID       string `json:"problemId"`
	Slug            string `json:"slug"`
	Title           string `json:"title"`
	SolvedUserCount int64  `json:"solvedUserCount"`
}

// --- Audit log ---
type AuditLog struct {
	ID           string          `json:"id"`
	ActorID      *string         `json:"actorId"`
	ActorEmail   *string         `json:"actorEmail"`
	Action       string          `json:"action"`
	ResourceType *string         `json:"resourceType"`
	ResourceID   *string         `json:"resourceId"`
	Detail       json.RawMessage `json:"detail"`
	CreatedAt    time.Time       `json:"createdAt"`
}

// --- Pagination ---
type ProblemListResponse struct {
	Problems []models.Problem `json:"problems"`
	Page     int              `json:"page"`
	Limit    int              `json:"limit"`
	Total    int              `json:"total"`
}
