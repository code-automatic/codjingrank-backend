package githubapp

import "context"

// RepoInfo holds GitHub repository metadata.
type RepoInfo struct {
	FullName      string
	URL           string
	GithubRepoID  string
	DefaultBranch string
}

// FileInfo holds GitHub file content metadata.
type FileInfo struct {
	SHA     string
	Content string
}

// CommitResult holds the result of a file commit operation.
type CommitResult struct {
	ContentSHA string
	CommitSHA  string
	CommitURL  string
}

// GitHubClient defines the interface for GitHub App API operations.
type GitHubClient interface {
	// GetInstallationToken exchanges a GitHub App JWT for an installation access token.
	GetInstallationToken(ctx context.Context, installationID string) (string, error)

	// GetInstallation retrieves installation details (owner login and type) from GitHub.
	GetInstallation(ctx context.Context, installationID string) (owner, ownerType string, err error)

	// EnsureRepo checks if a repository exists and creates it if not.
	EnsureRepo(ctx context.Context, token, owner, repoName, ownerType, defaultBranch string) (*RepoInfo, error)

	// GetFileContent retrieves a file's content and SHA from a GitHub repository.
	// Returns nil, nil if the file does not exist (HTTP 404).
	GetFileContent(ctx context.Context, token, owner, repoName, filePath, branch string) (*FileInfo, error)

	// CreateOrUpdateFile creates or updates a file in a GitHub repository.
	// If existingSHA is non-nil, the file will be updated (must match current SHA).
	// If existingSHA is nil, a new file will be created.
	CreateOrUpdateFile(ctx context.Context, token, owner, repoName, filePath, branch, content, commitMessage string, existingSHA *string, authorName, authorEmail string) (*CommitResult, error)
}
