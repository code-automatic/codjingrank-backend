package githubapp

import "context"

// MockClient implements GitHubClient for testing.
type MockClient struct {
	GetInstallationTokenFn func(ctx context.Context, installationID string) (string, error)
	GetInstallationFn      func(ctx context.Context, installationID string) (string, string, error)
	EnsureRepoFn           func(ctx context.Context, token, owner, repoName, ownerType, defaultBranch string) (*RepoInfo, error)
	GetFileContentFn       func(ctx context.Context, token, owner, repoName, filePath, branch string) (*FileInfo, error)
	CreateOrUpdateFileFn   func(ctx context.Context, token, owner, repoName, filePath, branch, content, commitMessage string, existingSHA *string, authorName, authorEmail string) (*CommitResult, error)
}

func (m *MockClient) GetInstallationToken(ctx context.Context, installationID string) (string, error) {
	return m.GetInstallationTokenFn(ctx, installationID)
}

func (m *MockClient) GetInstallation(ctx context.Context, installationID string) (string, string, error) {
	return m.GetInstallationFn(ctx, installationID)
}

func (m *MockClient) EnsureRepo(ctx context.Context, token, owner, repoName, ownerType, defaultBranch string) (*RepoInfo, error) {
	return m.EnsureRepoFn(ctx, token, owner, repoName, ownerType, defaultBranch)
}

func (m *MockClient) GetFileContent(ctx context.Context, token, owner, repoName, filePath, branch string) (*FileInfo, error) {
	return m.GetFileContentFn(ctx, token, owner, repoName, filePath, branch)
}

func (m *MockClient) CreateOrUpdateFile(ctx context.Context, token, owner, repoName, filePath, branch, content, commitMessage string, existingSHA *string, authorName, authorEmail string) (*CommitResult, error) {
	return m.CreateOrUpdateFileFn(ctx, token, owner, repoName, filePath, branch, content, commitMessage, existingSHA, authorName, authorEmail)
}
