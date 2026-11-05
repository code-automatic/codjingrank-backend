package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"innogen-backend/repo_service/internal/dto"
	"innogen-backend/repo_service/internal/githubapp"
	"innogen-backend/repo_service/internal/pathbuilder"
	"innogen-backend/repo_service/repository"
	"innogen-backend/shared/config"
	"innogen-backend/shared/languageutil"
	"innogen-backend/shared/models"
)

var (
	ErrRepositoryNotFound        = errors.New("repository not found")
	ErrGithubAccountNotFound     = errors.New("github account not found")
	ErrInstallationNotFound      = errors.New("installation not found")
	ErrInstallationAlreadyLinked = errors.New("installation already linked to another user")
)

// CommitAcceptedSubmissionRequest is the internal request for committing an accepted submission.
type CommitAcceptedSubmissionRequest struct {
	SubmissionID string
	UserID       int
	ProblemID    int
	LanguageID   int
	Code         string
}

// RepoService handles repository and commit business logic.
type RepoService struct {
	repo         *repository.RepoRepository
	githubClient githubapp.GitHubClient
	cfg          *config.Config
	log          *slog.Logger
}

// New creates a new RepoService.
func New(repo *repository.RepoRepository, githubClient githubapp.GitHubClient, cfg *config.Config, log *slog.Logger) *RepoService {
	return &RepoService{repo: repo, githubClient: githubClient, cfg: cfg, log: log}
}

// CommitSubmission triggers the mock commit flow for an Accepted submission.
// This is a best-effort operation: errors are logged but should not fail the caller's pipeline.
func (s *RepoService) CommitSubmission(ctx context.Context, submissionID string, userID, problemID, languageID int) error {
	// Idempotency guard: if already committed, skip
	_, commitSHA, err := s.repo.GetSubmissionRepoInfo(ctx, submissionID)
	if err != nil {
		return err
	}
	if commitSHA != "" {
		s.log.Info("submission already committed, skipping",
			slog.String("submissionId", submissionID),
		)
		return nil
	}

	// 1. Load curriculum context
	curriculum, err := s.repo.GetCurriculumContext(ctx, problemID)
	if err != nil {
		return err
	}

	// 2. Load language info
	lang, err := s.repo.GetLanguageByID(ctx, languageID)
	if err != nil {
		return err
	}
	if lang == nil {
		return fmt.Errorf("language %d not found", languageID)
	}

	// 3. Build paths
	fileName := languageutil.DetermineFileName(lang)
	newRepoPath := pathbuilder.BuildFilePath(*curriculum, fileName)
	repoName := pathbuilder.BuildRepoName(curriculum.SubjectSlug)
	newCommitSHA := pathbuilder.GenerateCommitSHA()

	// 4-6. Execute commit operations in a transaction
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	repoID, err := s.repo.UpsertRepositoryTx(ctx, tx, userID, curriculum.SubjectID, repoName)
	if err != nil {
		return err
	}
	if err := s.repo.UpdateSubmissionRepoInfoTx(ctx, tx, submissionID, newRepoPath, newCommitSHA); err != nil {
		return err
	}
	if err := s.repo.InsertSubmissionCommitTx(ctx, tx, submissionID, repoID, newRepoPath, newCommitSHA); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	s.log.Info("mock commit created",
		slog.String("submissionId", submissionID),
		slog.String("repoPath", newRepoPath),
		slog.String("commitSha", newCommitSHA),
		slog.Int("repoId", repoID),
	)

	return nil
}

// CommitAcceptedSubmission performs the real GitHub commit flow for an Accepted submission.
func (s *RepoService) CommitAcceptedSubmission(ctx context.Context, req CommitAcceptedSubmissionRequest) (*dto.CommitAcceptedSubmissionResponse, error) {
	// Idempotency guard: if already committed, skip
	_, commitSHA, err := s.repo.GetSubmissionRepoInfo(ctx, req.SubmissionID)
	if err != nil {
		return nil, err
	}
	if commitSHA != "" {
		s.log.Info("submission already committed, skipping",
			slog.String("submissionId", req.SubmissionID),
		)
		return &dto.CommitAcceptedSubmissionResponse{
			SubmissionID: req.SubmissionID,
			Skipped:      true,
		}, nil
	}

	// 1. Get GitHub account
	githubAccount, err := s.repo.GetGithubAccountByUserID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if githubAccount == nil {
		return nil, ErrGithubAccountNotFound
	}

	// Validate OAuth identity for commit author
	if githubAccount.OAuthStatus != "connected" || githubAccount.CommitAuthorName == nil || githubAccount.GithubNoreplyEmail == nil || *githubAccount.CommitAuthorName == "" || *githubAccount.GithubNoreplyEmail == "" {
		s.log.Error("cannot commit: OAuth account not connected",
			slog.Int("userId", req.UserID),
			slog.String("oauthStatus", githubAccount.OAuthStatus),
		)
		return nil, fmt.Errorf("GitHub OAuth account is not connected. Connect GitHub account before committing.")
	}

	// 2. Load curriculum context
	curriculum, err := s.repo.GetCurriculumContext(ctx, req.ProblemID)
	if err != nil {
		return nil, err
	}

	// 3. Load language info
	lang, err := s.repo.GetLanguageByID(ctx, req.LanguageID)
	if err != nil {
		return nil, err
	}
	if lang == nil {
		return nil, fmt.Errorf("language %d not found", req.LanguageID)
	}

	// 4. Build paths
	fileName := languageutil.DetermineFileName(lang)
	newRepoPath := pathbuilder.BuildFilePath(*curriculum, fileName)
	repoName := pathbuilder.BuildRepoName(curriculum.SubjectSlug)

	// 5. Check GitHub client availability
	if s.githubClient == nil {
		return nil, fmt.Errorf("github client not configured")
	}

	// 6. Get installation token
	token, err := s.githubClient.GetInstallationToken(ctx, githubAccount.InstallationID)
	if err != nil {
		return nil, fmt.Errorf("get installation token: %w", err)
	}

	// 7. Ensure repo exists on GitHub
	repoInfo, err := s.githubClient.EnsureRepo(ctx, token, githubAccount.GithubOwner, repoName, githubAccount.GithubOwnerType, s.cfg.GitHubDefaultBranch)
	if err != nil {
		return nil, fmt.Errorf("ensure repo: %w", err)
	}

	// 8. Check if file already exists on GitHub
	fileInfo, err := s.githubClient.GetFileContent(ctx, token, githubAccount.GithubOwner, repoName, newRepoPath, s.cfg.GitHubDefaultBranch)
	if err != nil {
		return nil, fmt.Errorf("get file content: %w", err)
	}

	var existingSHA *string
	if fileInfo != nil {
		existingSHA = &fileInfo.SHA
	}

	// 9. Commit the file to GitHub
	commitMessage := fmt.Sprintf("Solve %s", curriculum.ProblemSlug)
	commitResult, err := s.githubClient.CreateOrUpdateFile(ctx, token, githubAccount.GithubOwner, repoName, newRepoPath, s.cfg.GitHubDefaultBranch, req.Code, commitMessage, existingSHA, *githubAccount.CommitAuthorName, *githubAccount.GithubNoreplyEmail)
	if err != nil {
		return nil, fmt.Errorf("create or update file: %w", err)
	}

	// 10. Persist in database transaction
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	repoID, err := s.repo.UpsertRepositoryWithOwnerTx(ctx, tx, req.UserID, curriculum.SubjectID, repoName, repoInfo.FullName, repoInfo.URL, repoInfo.GithubRepoID, githubAccount.GithubOwner, s.cfg.GitHubDefaultBranch)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateSubmissionRepoInfoTx(ctx, tx, req.SubmissionID, newRepoPath, commitResult.CommitSHA); err != nil {
		return nil, err
	}
	if err := s.repo.InsertSubmissionCommitWithURLTx(ctx, tx, req.SubmissionID, repoID, newRepoPath, commitResult.CommitSHA, commitResult.CommitURL); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.log.Info("accepted submission committed to GitHub",
		slog.String("submissionId", req.SubmissionID),
		slog.String("repoFullName", repoInfo.FullName),
		slog.String("filePath", newRepoPath),
		slog.String("commitSha", commitResult.CommitSHA),
		slog.Int("repoId", repoID),
	)

	return &dto.CommitAcceptedSubmissionResponse{
		SubmissionID: req.SubmissionID,
		RepositoryID: repoID,
		FilePath:     newRepoPath,
		CommitSha:    commitResult.CommitSHA,
		CommitURL:    commitResult.CommitURL,
		RepoFullName: repoInfo.FullName,
	}, nil
}

// ListRepositories returns all repositories for a user.
func (s *RepoService) ListRepositories(ctx context.Context, userID int) ([]models.Repository, error) {
	return s.repo.FindRepositoriesByUserID(ctx, userID)
}

// ListCommits returns all commits for a repository, after verifying ownership.
func (s *RepoService) ListCommits(ctx context.Context, userID, repositoryID int) ([]models.SubmissionCommit, error) {
	repo, err := s.repo.FindRepositoryByID(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	if repo == nil || repo.UserID != userID {
		return nil, ErrRepositoryNotFound
	}
	return s.repo.FindCommitsByRepositoryID(ctx, repositoryID)
}

// GetGithubConnection returns the GitHub connection status for a user.
func (s *RepoService) GetGithubConnection(ctx context.Context, userID int) (*dto.GithubConnectionResponse, error) {
	account, err := s.repo.GetGithubAccountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if account == nil || account.InstallationID == "" || strings.HasPrefix(account.InstallationID, "oauth_pending_") || account.GithubOwner == "" || account.Status != "active" {
		return &dto.GithubConnectionResponse{Connected: false}, nil
	}
	return &dto.GithubConnectionResponse{
		Connected:       true,
		InstallationID:  strPtr(account.InstallationID),
		GithubOwner:     strPtr(account.GithubOwner),
		GithubOwnerType: strPtr(account.GithubOwnerType),
		GithubUsername:  account.GithubUsername,
		Status:          strPtr(account.Status),
	}, nil
}

// LinkGithubInstallation links a GitHub App installation to a user.
func (s *RepoService) LinkGithubInstallation(ctx context.Context, userID int, installationID string) (*dto.LinkGithubInstallationResponse, error) {
	inst, err := s.repo.GetGithubInstallationByID(ctx, installationID)
	if err != nil {
		return nil, err
	}
	if inst == nil {
		// Verify via GitHub API and extract owner info
		if s.githubClient != nil {
			owner, ownerType, err := s.githubClient.GetInstallation(ctx, installationID)
			if err == nil && owner != "" {
				_ = s.repo.UpsertGithubInstallation(ctx, installationID, owner, ownerType)
				inst, _ = s.repo.GetGithubInstallationByID(ctx, installationID)
			}
		}
		if inst == nil {
			return nil, ErrInstallationNotFound
		}
	}

	// Check if installation already linked to another user
	existingAccount, err := s.repo.GetGithubAccountByInstallationID(ctx, installationID)
	if err != nil {
		return nil, err
	}
	if existingAccount != nil && existingAccount.UserID != userID {
		return nil, ErrInstallationAlreadyLinked
	}

	if err := s.repo.UpsertGithubAccount(ctx, userID, installationID, inst.GithubOwner, inst.GithubOwnerType); err != nil {
		return nil, err
	}

	s.log.Info("github installation linked",
		slog.Int("userId", userID),
		slog.String("installationId", installationID),
		slog.String("owner", inst.GithubOwner),
	)

	return &dto.LinkGithubInstallationResponse{
		Linked:          true,
		InstallationID:  installationID,
		GithubOwner:     inst.GithubOwner,
		GithubOwnerType: inst.GithubOwnerType,
	}, nil
}

// DisconnectInstallation soft-disconnects the GitHub App installation for a user.
func (s *RepoService) DisconnectInstallation(ctx context.Context, userID int) error {
	return s.repo.UpdateGithubAccountStatusByUserID(ctx, userID, "disconnected")
}

func strPtr(s string) *string { return &s }
