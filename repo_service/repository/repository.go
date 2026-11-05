package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"innogen-backend/repo_service/internal/pathbuilder"
	"innogen-backend/shared/models"
)

// RepoRepository handles database queries for repository and commit operations.
type RepoRepository struct {
	pool *pgxpool.Pool
}

// New creates a new RepoRepository.
func New(pool *pgxpool.Pool) *RepoRepository {
	return &RepoRepository{pool: pool}
}

// GetCurriculumContext loads the curriculum hierarchy for a problem.
// Returns nil, error if the problem is not linked to any lesson.
func (r *RepoRepository) GetCurriculumContext(ctx context.Context, problemID int) (*pathbuilder.CurriculumContext, error) {
	c := &pathbuilder.CurriculumContext{}
	err := r.pool.QueryRow(ctx,
		`SELECT subj.slug, subj.id, ss.order_index, l.order_index, lp.order_index, p.slug
		 FROM lesson_problems lp
		 JOIN lessons l ON lp.lesson_id = l.id
		 JOIN subject_sessions ss ON l.subject_session_id = ss.id
		 JOIN subjects subj ON ss.subject_id = subj.id
		 JOIN problems p ON lp.problem_id = p.id
		 WHERE lp.problem_id = $1
		 ORDER BY lp.order_index ASC
		 LIMIT 1`,
		problemID,
	).Scan(&c.SubjectSlug, &c.SubjectID, &c.SessionOrder, &c.LessonOrder, &c.ProblemOrder, &c.ProblemSlug)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("problem %d is not linked to any lesson", problemID)
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetCurriculumContext: %w", err)
	}
	return c, nil
}

// GetLanguageByID retrieves an active language by ID.
// Returns nil, nil if not found.
func (r *RepoRepository) GetLanguageByID(ctx context.Context, id int) (*models.Language, error) {
	l := &models.Language{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, piston_alias, piston_version, file_extension, default_file_name, is_active, created_at, updated_at
		 FROM languages WHERE id = $1 AND is_active = true`, id,
	).Scan(&l.ID, &l.Name, &l.PistonAlias, &l.PistonVersion, &l.FileExtension, &l.DefaultFileName,
		&l.IsActive, &l.CreatedAt, &l.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetLanguageByID: %w", err)
	}
	return l, nil
}

// UpsertRepository creates or updates a repository for a user+subject pair, returning the repository ID.
func (r *RepoRepository) UpsertRepository(ctx context.Context, userID, subjectID int, repoName string) (int, error) {
	var repoID int
	err := r.pool.QueryRow(ctx,
		`INSERT INTO repositories (user_id, subject_id, repo_name)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, subject_id) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
		 RETURNING id`,
		userID, subjectID, repoName,
	).Scan(&repoID)
	if err != nil {
		return 0, fmt.Errorf("repository.UpsertRepository: %w", err)
	}
	return repoID, nil
}

// UpdateSubmissionRepoInfo updates the repo_path and commit_sha on a submission.
func (r *RepoRepository) UpdateSubmissionRepoInfo(ctx context.Context, submissionID, repoPath, commitSHA string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE submissions SET repo_path = $2, commit_sha = $3 WHERE id = $1`,
		submissionID, repoPath, commitSHA,
	)
	if err != nil {
		return fmt.Errorf("repository.UpdateSubmissionRepoInfo: %w", err)
	}
	return nil
}

// InsertSubmissionCommit inserts a row into the submission_commits table.
func (r *RepoRepository) InsertSubmissionCommit(ctx context.Context, submissionID string, repositoryID int, filePath, commitSHA string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO submission_commits (submission_id, repository_id, file_path, commit_sha)
		 VALUES ($1, $2, $3, $4)`,
		submissionID, repositoryID, filePath, commitSHA,
	)
	if err != nil {
		return fmt.Errorf("repository.InsertSubmissionCommit: %w", err)
	}
	return nil
}

// FindRepositoryByID looks up a repository by its ID.
// Returns nil, nil if not found.
func (r *RepoRepository) FindRepositoryByID(ctx context.Context, id int) (*models.Repository, error) {
	repo := &models.Repository{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, subject_id, repo_name, repo_full_name, repo_url, github_repo_id, github_owner, status, default_branch, created_at, updated_at
		 FROM repositories WHERE id = $1`, id,
	).Scan(&repo.ID, &repo.UserID, &repo.SubjectID, &repo.RepoName,
		&repo.RepoFullName, &repo.RepoURL, &repo.GithubRepoID, &repo.GithubOwner, &repo.Status, &repo.DefaultBranch,
		&repo.CreatedAt, &repo.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.FindRepositoryByID: %w", err)
	}
	return repo, nil
}

// GetSubmissionRepoInfo returns the repo_path and commit_sha for a submission.
// Returns empty strings if not set.
func (r *RepoRepository) GetSubmissionRepoInfo(ctx context.Context, submissionID string) (string, string, error) {
	var repoPath, commitSHA *string
	err := r.pool.QueryRow(ctx,
		`SELECT repo_path, commit_sha FROM submissions WHERE id = $1`, submissionID,
	).Scan(&repoPath, &commitSHA)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", nil
	}
	if err != nil {
		return "", "", fmt.Errorf("repository.GetSubmissionRepoInfo: %w", err)
	}
	rp := ""
	cs := ""
	if repoPath != nil {
		rp = *repoPath
	}
	if commitSHA != nil {
		cs = *commitSHA
	}
	return rp, cs, nil
}

// BeginTx starts a new database transaction.
func (r *RepoRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

// UpsertRepositoryTx is the transactional version of UpsertRepository.
func (r *RepoRepository) UpsertRepositoryTx(ctx context.Context, tx pgx.Tx, userID, subjectID int, repoName string) (int, error) {
	var repoID int
	err := tx.QueryRow(ctx,
		`INSERT INTO repositories (user_id, subject_id, repo_name)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, subject_id) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
		 RETURNING id`,
		userID, subjectID, repoName,
	).Scan(&repoID)
	if err != nil {
		return 0, fmt.Errorf("repository.UpsertRepositoryTx: %w", err)
	}
	return repoID, nil
}

// UpdateSubmissionRepoInfoTx is the transactional version of UpdateSubmissionRepoInfo.
func (r *RepoRepository) UpdateSubmissionRepoInfoTx(ctx context.Context, tx pgx.Tx, submissionID, repoPath, commitSHA string) error {
	_, err := tx.Exec(ctx,
		`UPDATE submissions SET repo_path = $2, commit_sha = $3 WHERE id = $1`,
		submissionID, repoPath, commitSHA,
	)
	if err != nil {
		return fmt.Errorf("repository.UpdateSubmissionRepoInfoTx: %w", err)
	}
	return nil
}

// InsertSubmissionCommitTx is the transactional version of InsertSubmissionCommit.
func (r *RepoRepository) InsertSubmissionCommitTx(ctx context.Context, tx pgx.Tx, submissionID string, repositoryID int, filePath, commitSHA string) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO submission_commits (submission_id, repository_id, file_path, commit_sha)
		 VALUES ($1, $2, $3, $4)`,
		submissionID, repositoryID, filePath, commitSHA,
	)
	if err != nil {
		return fmt.Errorf("repository.InsertSubmissionCommitTx: %w", err)
	}
	return nil
}

// FindRepositoriesByUserID returns all repositories for a user.
func (r *RepoRepository) FindRepositoriesByUserID(ctx context.Context, userID int) ([]models.Repository, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, subject_id, repo_name, repo_full_name, repo_url, github_repo_id, github_owner, status, default_branch, created_at, updated_at
		 FROM repositories WHERE user_id = $1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("repository.FindRepositoriesByUserID: %w", err)
	}
	defer rows.Close()

	var repos []models.Repository
	for rows.Next() {
		var repo models.Repository
		if err := rows.Scan(&repo.ID, &repo.UserID, &repo.SubjectID, &repo.RepoName,
			&repo.RepoFullName, &repo.RepoURL, &repo.GithubRepoID, &repo.GithubOwner, &repo.Status, &repo.DefaultBranch,
			&repo.CreatedAt, &repo.UpdatedAt); err != nil {
			return nil, fmt.Errorf("repository.FindRepositoriesByUserID: %w", err)
		}
		repos = append(repos, repo)
	}
	return repos, rows.Err()
}

// FindCommitsByRepositoryID returns all commits for a repository.
func (r *RepoRepository) FindCommitsByRepositoryID(ctx context.Context, repositoryID int) ([]models.SubmissionCommit, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, submission_id, repository_id, file_path, commit_sha, commit_url, created_at
		 FROM submission_commits WHERE repository_id = $1 ORDER BY created_at DESC`, repositoryID,
	)
	if err != nil {
		return nil, fmt.Errorf("repository.FindCommitsByRepositoryID: %w", err)
	}
	defer rows.Close()

	var commits []models.SubmissionCommit
	for rows.Next() {
		var commit models.SubmissionCommit
		if err := rows.Scan(&commit.ID, &commit.SubmissionID, &commit.RepositoryID,
			&commit.FilePath, &commit.CommitSha, &commit.CommitURL, &commit.CreatedAt); err != nil {
			return nil, fmt.Errorf("repository.FindCommitsByRepositoryID: %w", err)
		}
		commits = append(commits, commit)
	}
	return commits, rows.Err()
}

// GetGithubAccountByUserID retrieves the GitHub account for a user.
// Returns nil, nil if not found.
func (r *RepoRepository) GetGithubAccountByUserID(ctx context.Context, userID int) (*models.GithubAccount, error) {
	a := &models.GithubAccount{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, installation_id, github_user_id, github_username, github_avatar_url,
                github_owner, github_owner_type, status, github_noreply_email, commit_author_name,
                oauth_connected_at, oauth_status, installed_at, uninstalled_at, created_at, updated_at
         FROM github_accounts WHERE user_id = $1`, userID,
	).Scan(&a.ID, &a.UserID, &a.InstallationID, &a.GithubUserID, &a.GithubUsername,
		&a.GithubAvatarURL, &a.GithubOwner, &a.GithubOwnerType, &a.Status,
		&a.GithubNoreplyEmail, &a.CommitAuthorName, &a.OAuthConnectedAt, &a.OAuthStatus,
		&a.InstalledAt, &a.UninstalledAt, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetGithubAccountByUserID: %w", err)
	}
	return a, nil
}

// GetGithubAccountByInstallationID retrieves a github_account by installation_id.
// Returns nil, nil if not found.
func (r *RepoRepository) GetGithubAccountByInstallationID(ctx context.Context, installationID string) (*models.GithubAccount, error) {
	a := &models.GithubAccount{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, installation_id, github_user_id, github_username, github_avatar_url,
                github_owner, github_owner_type, status, github_noreply_email, commit_author_name,
                oauth_connected_at, oauth_status, installed_at, uninstalled_at, created_at, updated_at
         FROM github_accounts WHERE installation_id = $1`, installationID,
	).Scan(&a.ID, &a.UserID, &a.InstallationID, &a.GithubUserID, &a.GithubUsername,
		&a.GithubAvatarURL, &a.GithubOwner, &a.GithubOwnerType, &a.Status,
		&a.GithubNoreplyEmail, &a.CommitAuthorName, &a.OAuthConnectedAt, &a.OAuthStatus,
		&a.InstalledAt, &a.UninstalledAt, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetGithubAccountByInstallationID: %w", err)
	}
	return a, nil
}

// UpsertRepositoryWithOwnerTx creates or updates a repository including GitHub owner info.
func (r *RepoRepository) UpsertRepositoryWithOwnerTx(ctx context.Context, tx pgx.Tx, userID, subjectID int, repoName, repoFullName, repoURL, githubRepoID, githubOwner, defaultBranch string) (int, error) {
	var repoID int
	err := tx.QueryRow(ctx,
		`INSERT INTO repositories (user_id, subject_id, repo_name, repo_full_name, repo_url, github_repo_id, github_owner, default_branch)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
         ON CONFLICT (user_id, subject_id) DO UPDATE SET
             repo_full_name = EXCLUDED.repo_full_name,
             repo_url = EXCLUDED.repo_url,
             github_repo_id = EXCLUDED.github_repo_id,
             github_owner = EXCLUDED.github_owner,
             updated_at = CURRENT_TIMESTAMP
         RETURNING id`,
		userID, subjectID, repoName, repoFullName, repoURL, githubRepoID, githubOwner, defaultBranch,
	).Scan(&repoID)
	if err != nil {
		return 0, fmt.Errorf("repository.UpsertRepositoryWithOwnerTx: %w", err)
	}
	return repoID, nil
}

// InsertSubmissionCommitWithURLTx inserts a commit record with URL.
func (r *RepoRepository) InsertSubmissionCommitWithURLTx(ctx context.Context, tx pgx.Tx, submissionID string, repositoryID int, filePath, commitSHA, commitURL string) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO submission_commits (submission_id, repository_id, file_path, commit_sha, commit_url)
         VALUES ($1, $2, $3, $4, $5)`,
		submissionID, repositoryID, filePath, commitSHA, commitURL,
	)
	if err != nil {
		return fmt.Errorf("repository.InsertSubmissionCommitWithURLTx: %w", err)
	}
	return nil
}

// UpsertGithubInstallation creates or updates a GitHub installation record.
func (r *RepoRepository) UpsertGithubInstallation(ctx context.Context, installationID, owner, ownerType string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO github_installations (installation_id, github_owner, github_owner_type, is_active, installed_at)
         VALUES ($1, $2, $3, true, CURRENT_TIMESTAMP)
         ON CONFLICT (installation_id) DO UPDATE SET
             github_owner = EXCLUDED.github_owner,
             github_owner_type = EXCLUDED.github_owner_type,
             is_active = true,
             installed_at = COALESCE(github_installations.installed_at, CURRENT_TIMESTAMP),
             uninstalled_at = NULL,
             updated_at = CURRENT_TIMESTAMP`,
		installationID, owner, ownerType,
	)
	if err != nil {
		return fmt.Errorf("repository.UpsertGithubInstallation: %w", err)
	}
	return nil
}

// UpdateGithubInstallationStatus sets the active status of a GitHub installation.
func (r *RepoRepository) UpdateGithubInstallationStatus(ctx context.Context, installationID string, isActive bool) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE github_installations SET is_active=$2, uninstalled_at=CASE WHEN $2 THEN NULL ELSE COALESCE(uninstalled_at, CURRENT_TIMESTAMP) END, updated_at=CURRENT_TIMESTAMP WHERE installation_id=$1`,
		installationID, isActive,
	)
	if err != nil {
		return fmt.Errorf("repository.UpdateGithubInstallationStatus: %w", err)
	}
	return nil
}

// UpdateGithubAccountStatusByInstallation updates the status of github_accounts linked to an installation.
func (r *RepoRepository) UpdateGithubAccountStatusByInstallation(ctx context.Context, installationID, status string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE github_accounts SET status=$2, uninstalled_at=CASE WHEN $3 IN ('deleted','suspended') THEN COALESCE(uninstalled_at, CURRENT_TIMESTAMP) ELSE NULL END, updated_at=CURRENT_TIMESTAMP WHERE installation_id=$1`,
		installationID, status, status,
	)
	if err != nil {
		return fmt.Errorf("repository.UpdateGithubAccountStatusByInstallation: %w", err)
	}
	return nil
}

// SetRepositoryStatusByGithubRepoID updates the status of all local repositories
// that reference the given GitHub repo ID. If multiple users have linked the same
// GitHub repo, all of their local records will be updated (intentional — the GitHub
// repo was deleted/moved, so all references should be updated).
func (r *RepoRepository) SetRepositoryStatusByGithubRepoID(ctx context.Context, githubRepoID, status string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE repositories SET status = $2, updated_at = CURRENT_TIMESTAMP WHERE github_repo_id = $1`,
		githubRepoID, status,
	)
	if err != nil {
		return fmt.Errorf("repository.SetRepositoryStatusByGithubRepoID: %w", err)
	}
	return nil
}

// UpdateRepositoryByGithubRepoID updates the full name and name of a renamed repository.
func (r *RepoRepository) UpdateRepositoryByGithubRepoID(ctx context.Context, githubRepoID, repoFullName, repoName string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE repositories SET repo_full_name = $2, repo_name = $3, status = 'active', updated_at = CURRENT_TIMESTAMP WHERE github_repo_id = $1`,
		githubRepoID, repoFullName, repoName,
	)
	if err != nil {
		return fmt.Errorf("repository.UpdateRepositoryByGithubRepoID: %w", err)
	}
	return nil
}

// UpdateRepositoriesStatusByOwner bulk-updates repository statuses for all repos owned by a GitHub owner.
func (r *RepoRepository) UpdateRepositoriesStatusByOwner(ctx context.Context, githubOwner, status string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE repositories SET status = $2, updated_at = CURRENT_TIMESTAMP WHERE github_owner = $1`,
		githubOwner, status,
	)
	if err != nil {
		return fmt.Errorf("repository.UpdateRepositoriesStatusByOwner: %w", err)
	}
	return nil
}

// GetGithubInstallationByID retrieves a GitHub installation by its installation_id.
// Returns nil, nil if not found.
func (r *RepoRepository) GetGithubInstallationByID(ctx context.Context, installationID string) (*models.GithubInstallation, error) {
	inst := &models.GithubInstallation{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, installation_id, github_owner, github_owner_type, is_active, installed_at, uninstalled_at, created_at, updated_at
         FROM github_installations WHERE installation_id = $1`, installationID,
	).Scan(&inst.ID, &inst.InstallationID, &inst.GithubOwner, &inst.GithubOwnerType, &inst.IsActive, &inst.InstalledAt, &inst.UninstalledAt, &inst.CreatedAt, &inst.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository.GetGithubInstallationByID: %w", err)
	}
	return inst, nil
}

// UpsertGithubAccount creates or updates a github_accounts row for a user.
func (r *RepoRepository) UpsertGithubAccount(ctx context.Context, userID int, installationID, githubOwner, githubOwnerType string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO github_accounts (user_id, installation_id, github_owner, github_owner_type, status, installed_at)
         VALUES ($1, $2, $3, $4, 'active', CURRENT_TIMESTAMP)
         ON CONFLICT (user_id) DO UPDATE SET
             installation_id = EXCLUDED.installation_id,
             github_owner = EXCLUDED.github_owner,
             github_owner_type = EXCLUDED.github_owner_type,
             status = 'active',
             installed_at = COALESCE(github_accounts.installed_at, CURRENT_TIMESTAMP),
             uninstalled_at = NULL,
             updated_at = CURRENT_TIMESTAMP`,
		userID, installationID, githubOwner, githubOwnerType,
	)
	if err != nil {
		return fmt.Errorf("repository.UpsertGithubAccount: %w", err)
	}
	return nil
}

// UpdateGithubAccountOwnerByInstallation updates the installation_id, owner, owner_type, and status on github_accounts
// linked to a given installation_id. Used by webhook handler to backfill data set by OAuth callback.
func (r *RepoRepository) UpdateGithubAccountOwnerByInstallation(ctx context.Context, installationID, owner, ownerType, status string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE github_accounts SET installation_id = $2, github_owner = $3, github_owner_type = $4, status = $5, updated_at = CURRENT_TIMESTAMP WHERE installation_id = $1`,
		installationID, installationID, owner, ownerType, status,
	)
	if err != nil {
		return fmt.Errorf("repository.UpdateGithubAccountOwnerByInstallation: %w", err)
	}
	return nil
}

func (r *RepoRepository) UpsertGitHubOAuth(ctx context.Context, userID int, githubUserID, githubUsername, githubAvatarURL, githubNoreplyEmail, commitAuthorName string) error {
	installationID := fmt.Sprintf("oauth_pending_%d", userID)
	_, err := r.pool.Exec(ctx,
		`INSERT INTO github_accounts (user_id, installation_id, github_owner, github_owner_type, status, github_user_id, github_username, github_avatar_url, github_noreply_email, commit_author_name, oauth_connected_at, oauth_status)
         VALUES ($1, $2, '', 'User', 'active', $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, 'connected')
         ON CONFLICT (user_id) DO UPDATE SET
             github_user_id = EXCLUDED.github_user_id,
             github_username = EXCLUDED.github_username,
             github_avatar_url = EXCLUDED.github_avatar_url,
             github_noreply_email = EXCLUDED.github_noreply_email,
             commit_author_name = EXCLUDED.commit_author_name,
             oauth_connected_at = EXCLUDED.oauth_connected_at,
             oauth_status = EXCLUDED.oauth_status,
             updated_at = CURRENT_TIMESTAMP`,
		userID, installationID, githubUserID, githubUsername, githubAvatarURL, githubNoreplyEmail, commitAuthorName,
	)
	if err != nil {
		return fmt.Errorf("repository.UpsertGitHubOAuth: %w", err)
	}
	return nil
}

func (r *RepoRepository) GetGitHubOAuthByUserID(ctx context.Context, userID int) (githubUserID, githubUsername, githubAvatarURL, githubNoreplyEmail, commitAuthorName, oauthStatus string, err error) {
	err = r.pool.QueryRow(ctx,
		`SELECT COALESCE(github_user_id,''), COALESCE(github_username,''), COALESCE(github_avatar_url,''), COALESCE(github_noreply_email,''), COALESCE(commit_author_name,''), COALESCE(oauth_status,'disconnected') FROM github_accounts WHERE user_id=$1`, userID,
	).Scan(&githubUserID, &githubUsername, &githubAvatarURL, &githubNoreplyEmail, &commitAuthorName, &oauthStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", "", "", "", "disconnected", nil
	}
	if err != nil {
		return "", "", "", "", "", "", fmt.Errorf("repository.GetGitHubOAuthByUserID: %w", err)
	}
	return
}

func (r *RepoRepository) DisconnectGitHubOAuth(ctx context.Context, userID int) error {
	_, err := r.pool.Exec(ctx, `UPDATE github_accounts SET oauth_status='disconnected', updated_at=CURRENT_TIMESTAMP WHERE user_id=$1`, userID)
	if err != nil {
		return fmt.Errorf("repository.DisconnectGitHubOAuth: %w", err)
	}
	return nil
}

func (r *RepoRepository) UpdateGithubAccountStatusByUserID(ctx context.Context, userID int, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE github_accounts SET status=$2, uninstalled_at=CASE WHEN $3 IN ('deleted','suspended','disconnected') THEN COALESCE(uninstalled_at, CURRENT_TIMESTAMP) ELSE NULL END, updated_at=CURRENT_TIMESTAMP WHERE user_id=$1`, userID, status, status)
	if err != nil {
		return fmt.Errorf("repository.UpdateGithubAccountStatusByUserID: %w", err)
	}
	return nil
}
