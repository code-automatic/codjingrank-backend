package webhook

import "strconv"

// InstallationEvent represents the GitHub 'installation' webhook payload.
type InstallationEvent struct {
	Action       string `json:"action"`
	Installation struct {
		ID      int64 `json:"id"`
		Account struct {
			Login     string `json:"login"`
			ID        int64  `json:"id"`
			Type      string `json:"type"`
			AvatarURL string `json:"avatar_url"`
		} `json:"account"`
	} `json:"installation"`
}

func (e InstallationEvent) InstallationID() string {
	return strconv.FormatInt(e.Installation.ID, 10)
}

// InstallationRepositoriesEvent represents the GitHub 'installation_repositories' webhook payload.
type InstallationRepositoriesEvent struct {
	Action       string `json:"action"`
	Installation struct {
		ID int64 `json:"id"`
	} `json:"installation"`
	RepositoriesAdded []struct {
		ID       int64  `json:"id"`
		FullName string `json:"full_name"`
	} `json:"repositories_added"`
	RepositoriesRemoved []struct {
		ID       int64  `json:"id"`
		FullName string `json:"full_name"`
	} `json:"repositories_removed"`
}

// RepositoryEvent represents the GitHub 'repository' webhook payload.
type RepositoryEvent struct {
	Action     string `json:"action"`
	Repository struct {
		ID       int64  `json:"id"`
		FullName string `json:"full_name"`
		Name     string `json:"name"`
	} `json:"repository"`
}
