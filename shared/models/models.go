package models

import (
	"encoding/json"
	"time"
)

// User represents a row in the users table.
// Password is tagged json:"-" to prevent accidental serialization in API responses.
type User struct {
	ID            int        `json:"id"`
	Email         string     `json:"email"`
	Password      string     `json:"-"`
	Username      *string    `json:"username"`
	FullName      *string    `json:"fullName"`
	Role          string     `json:"role"`
	IsActive      bool       `json:"isActive"`
	InstalledAt   *time.Time `json:"installedAt"`
	UninstalledAt *time.Time `json:"uninstalledAt"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// Subject represents a row in the subjects table.
type Subject struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Color       *string   `json:"color"`
	IsPublished bool      `json:"isPublished"`
	LanguageID  *int      `json:"languageId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// SubjectSession represents a row in the subject_sessions table.
type SubjectSession struct {
	ID          int       `json:"id"`
	SubjectID   int       `json:"subjectId"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	OrderIndex  int       `json:"orderIndex"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Lesson represents a row in the lessons table.
type Lesson struct {
	ID               int       `json:"id"`
	SubjectSessionID int       `json:"subjectSessionId"`
	Title            string    `json:"title"`
	ContentMD        *string   `json:"contentMd"`
	OrderIndex       int       `json:"orderIndex"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// Problem represents a row in the problems table.
type Problem struct {
	ID               int             `json:"id"`
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
	ID             int       `json:"id"`
	ProblemID      int       `json:"problemId"`
	Visibility     string    `json:"visibility"`
	InputData      *string   `json:"inputData"`
	ExpectedOutput string    `json:"expectedOutput"`
	ExecuteCode    *string   `json:"executeCode"`
	OrderIndex     int       `json:"orderIndex"`
	CreatedAt      time.Time `json:"createdAt"`
}

// LessonProblem represents a row in the lesson_problems join table.
type LessonProblem struct {
	LessonID   int `json:"lessonId"`
	ProblemID  int `json:"problemId"`
	OrderIndex int `json:"orderIndex"`
}

// Language represents a row in the languages table.
type Language struct {
	ID              int       `json:"id"`
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
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

// GithubAccount represents a row in the github_accounts table.
type GithubAccount struct {
	ID                 int        `json:"id"`
	UserID             int        `json:"userId"`
	InstallationID     string     `json:"installationId"`
	GithubUserID       *string    `json:"githubUserId"`
	GithubUsername     *string    `json:"githubUsername"`
	GithubAvatarURL    *string    `json:"githubAvatarUrl"`
	GithubOwner        string     `json:"githubOwner"`
	GithubOwnerType    string     `json:"githubOwnerType"`
	Status             string     `json:"status"`
	GithubNoreplyEmail *string    `json:"githubNoreplyEmail"`
	CommitAuthorName   *string    `json:"commitAuthorName"`
	OAuthConnectedAt   *time.Time `json:"oauthConnectedAt"`
	OAuthStatus        string     `json:"oauthStatus"`
	InstalledAt        *time.Time `json:"installedAt"`
	UninstalledAt      *time.Time `json:"uninstalledAt"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

// GithubInstallation represents a row in the github_installations table.
type GithubInstallation struct {
	ID              int        `json:"id"`
	InstallationID  string     `json:"installationId"`
	GithubOwner     string     `json:"githubOwner"`
	GithubOwnerType string     `json:"githubOwnerType"`
	IsActive        bool       `json:"isActive"`
	InstalledAt     *time.Time `json:"installedAt"`
	UninstalledAt   *time.Time `json:"uninstalledAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// Repository represents a row in the repositories table.
type Repository struct {
	ID            int       `json:"id"`
	UserID        int       `json:"userId"`
	SubjectID     int       `json:"subjectId"`
	RepoName      string    `json:"repoName"`
	RepoFullName  *string   `json:"repoFullName"`
	RepoURL       *string   `json:"repoUrl"`
	GithubRepoID  *string   `json:"githubRepoId"`
	GithubOwner   *string   `json:"githubOwner"`
	Status        string    `json:"status"`
	DefaultBranch string    `json:"defaultBranch"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// SubmissionCommit represents a row in the submission_commits table.
type SubmissionCommit struct {
	ID           string    `json:"id"`
	SubmissionID string    `json:"submissionId"`
	RepositoryID int       `json:"repositoryId"`
	FilePath     string    `json:"filePath"`
	CommitSha    string    `json:"commitSha"`
	CommitURL    *string   `json:"commitUrl"`
	CreatedAt    time.Time `json:"createdAt"`
}

// Submission represents a row in the submissions table.
type Submission struct {
	ID             string     `json:"id"`
	UserID         int        `json:"userId"`
	ProblemID      int        `json:"problemId"`
	LanguageID     int        `json:"languageId"`
	Code           string     `json:"code"`
	Status         string     `json:"status"`
	RuntimeMs      *int       `json:"runtimeMs"`
	MemoryKb       *int       `json:"memoryKb"`
	ErrorMessage   *string    `json:"errorMessage"`
	PassCount      int        `json:"passCount"`
	TotalTestcases int        `json:"totalTestcases"`
	RepoPath       *string    `json:"repoPath"`
	CommitSha      *string    `json:"commitSha"`
	CommitURL      *string    `json:"commitUrl"`
	CreatedAt      time.Time  `json:"createdAt"`
	JudgedAt       *time.Time `json:"judgedAt"`
}
