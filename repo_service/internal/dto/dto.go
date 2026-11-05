package dto

import "innogen-backend/shared/models"

// CommitAcceptedSubmissionRequest is the body for POST /internal/commits/accepted-submission.
type CommitAcceptedSubmissionRequest struct {
	SubmissionID string `json:"submissionId"`
	UserID       int    `json:"userId"`
	ProblemID    int    `json:"problemId"`
	LanguageID   int    `json:"languageId"`
	Code         string `json:"code"`
}

// CommitAcceptedSubmissionResponse is returned after the commit flow.
type CommitAcceptedSubmissionResponse struct {
	SubmissionID string `json:"submissionId"`
	RepositoryID int    `json:"repositoryId"`
	FilePath     string `json:"filePath"`
	CommitSha    string `json:"commitSha"`
	CommitURL    string `json:"commitUrl"`
	RepoFullName string `json:"repoFullName"`
	Skipped      bool   `json:"skipped,omitempty"`
}

// RepositoryResponse wraps a repository for HTTP responses.
type RepositoryResponse struct {
	Repository *models.Repository `json:"repository"`
}

// ListRepositoriesResponse wraps a list of repositories.
type ListRepositoriesResponse struct {
	Repositories []models.Repository `json:"repositories"`
}

// ListCommitsResponse wraps a list of commits.
type ListCommitsResponse struct {
	Commits []models.SubmissionCommit `json:"commits"`
}

type GithubConnectionResponse struct {
	Connected       bool    `json:"connected"`
	InstallationID  *string `json:"installationId,omitempty"`
	GithubOwner     *string `json:"githubOwner,omitempty"`
	GithubOwnerType *string `json:"githubOwnerType,omitempty"`
	GithubUsername  *string `json:"githubUsername,omitempty"`
	Status          *string `json:"status,omitempty"`
}

type LinkGithubInstallationRequest struct {
	InstallationID string `json:"installationId"`
}

type LinkGithubInstallationResponse struct {
	Linked          bool   `json:"linked"`
	InstallationID  string `json:"installationId"`
	GithubOwner     string `json:"githubOwner"`
	GithubOwnerType string `json:"githubOwnerType"`
}
