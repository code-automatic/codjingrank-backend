package webhook

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"

	"innogen-backend/repo_service/repository"
)

type WebhookService struct {
	repo *repository.RepoRepository
	log  *slog.Logger
}

func NewWebhookService(repo *repository.RepoRepository, log *slog.Logger) *WebhookService {
	return &WebhookService{repo: repo, log: log}
}

func (s *WebhookService) ProcessEvent(ctx context.Context, eventType string, body []byte) error {
	log := s.log.With(slog.String("event", eventType))
	log.Info("processing webhook event")

	switch eventType {
	case "installation":
		var event InstallationEvent
		if err := json.Unmarshal(body, &event); err != nil {
			return err
		}
		return s.handleInstallation(ctx, event)
	case "installation_repositories":
		var event InstallationRepositoriesEvent
		if err := json.Unmarshal(body, &event); err != nil {
			return err
		}
		return s.handleInstallationRepositories(ctx, event)
	case "repository":
		var event RepositoryEvent
		if err := json.Unmarshal(body, &event); err != nil {
			return err
		}
		return s.handleRepository(ctx, event)
	default:
		log.Info("ignoring unsupported event type")
		return nil
	}
}

func (s *WebhookService) handleInstallation(ctx context.Context, event InstallationEvent) error {
	instID := event.InstallationID()
	owner := event.Installation.Account.Login
	ownerType := event.Installation.Account.Type

	switch event.Action {
	case "created":
		if err := s.repo.UpsertGithubInstallation(ctx, instID, owner, ownerType); err != nil {
			return err
		}
		// Backfill owner info on linked github_accounts (from OAuth callback)
		if err := s.repo.UpdateGithubAccountOwnerByInstallation(ctx, instID, owner, ownerType, "active"); err != nil {
			s.log.Error("failed to backfill account owner on create", slog.String("error", err.Error()))
		}
		s.log.Info("installation created", slog.String("installationId", instID), slog.String("owner", owner))
	case "deleted":
		if err := s.repo.UpdateGithubInstallationStatus(ctx, instID, false); err != nil {
			return err
		}
		// Secondary updates are best-effort
		if err := s.repo.UpdateGithubAccountStatusByInstallation(ctx, instID, "deleted"); err != nil {
			s.log.Error("failed to update account status on delete", slog.String("error", err.Error()))
		}
		if err := s.repo.UpdateRepositoriesStatusByOwner(ctx, owner, "deleted"); err != nil {
			s.log.Error("failed to update repo statuses on delete", slog.String("error", err.Error()))
		}
		s.log.Info("installation deleted", slog.String("installationId", instID), slog.String("owner", owner))
	case "suspend":
		if err := s.repo.UpdateGithubInstallationStatus(ctx, instID, false); err != nil {
			return err
		}
		// Secondary updates are best-effort
		if err := s.repo.UpdateGithubAccountStatusByInstallation(ctx, instID, "suspended"); err != nil {
			s.log.Error("failed to update account status on suspend", slog.String("error", err.Error()))
		}
		s.log.Info("installation suspended", slog.String("installationId", instID))
	case "unsuspend":
		if err := s.repo.UpdateGithubInstallationStatus(ctx, instID, true); err != nil {
			return err
		}
		// Secondary updates are best-effort
		if err := s.repo.UpdateGithubAccountStatusByInstallation(ctx, instID, "active"); err != nil {
			s.log.Error("failed to update account status on unsuspend", slog.String("error", err.Error()))
		}
		s.log.Info("installation unsuspended", slog.String("installationId", instID))
	default:
		s.log.Info("ignoring installation action", slog.String("action", event.Action))
	}
	return nil
}

func (s *WebhookService) handleInstallationRepositories(ctx context.Context, event InstallationRepositoriesEvent) error {
	switch event.Action {
	case "added":
		for _, repo := range event.RepositoriesAdded {
			if err := s.repo.SetRepositoryStatusByGithubRepoID(ctx, strconv.FormatInt(repo.ID, 10), "active"); err != nil {
				s.log.Error("failed to update repo status on added", slog.String("repoId", strconv.FormatInt(repo.ID, 10)), slog.String("error", err.Error()))
			}
		}
		s.log.Info("repositories added", slog.Int("count", len(event.RepositoriesAdded)))
	case "removed":
		for _, repo := range event.RepositoriesRemoved {
			if err := s.repo.SetRepositoryStatusByGithubRepoID(ctx, strconv.FormatInt(repo.ID, 10), "deleted"); err != nil {
				s.log.Error("failed to update repo status on removed", slog.String("repoId", strconv.FormatInt(repo.ID, 10)), slog.String("error", err.Error()))
			}
		}
		s.log.Info("repositories removed", slog.Int("count", len(event.RepositoriesRemoved)))
	default:
		s.log.Info("ignoring installation_repositories action", slog.String("action", event.Action))
	}
	return nil
}

func (s *WebhookService) handleRepository(ctx context.Context, event RepositoryEvent) error {
	repoID := strconv.FormatInt(event.Repository.ID, 10)
	fullName := event.Repository.FullName
	name := event.Repository.Name

	switch event.Action {
	case "renamed":
		if err := s.repo.UpdateRepositoryByGithubRepoID(ctx, repoID, fullName, name); err != nil {
			return err
		}
		s.log.Info("repository renamed", slog.String("repoId", repoID), slog.String("fullName", fullName))
	case "deleted":
		if err := s.repo.SetRepositoryStatusByGithubRepoID(ctx, repoID, "deleted"); err != nil {
			s.log.Error("failed to update repo status on deleted", slog.String("repoId", repoID), slog.String("error", err.Error()))
		}
		s.log.Info("repository deleted", slog.String("repoId", repoID))
	case "archived":
		if err := s.repo.SetRepositoryStatusByGithubRepoID(ctx, repoID, "archived"); err != nil {
			s.log.Error("failed to update repo status on archived", slog.String("repoId", repoID), slog.String("error", err.Error()))
		}
		s.log.Info("repository archived", slog.String("repoId", repoID))
	default:
		s.log.Info("ignoring repository action", slog.String("action", event.Action))
	}
	return nil
}
