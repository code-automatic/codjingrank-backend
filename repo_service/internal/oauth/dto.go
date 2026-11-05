package oauth

type OAuthStartURLResponse struct {
	URL string `json:"url"`
}

type GithubAccountResponse struct {
	Connected          bool    `json:"connected"`
	GithubUserID       *string `json:"githubUserId,omitempty"`
	GithubUsername     *string `json:"githubUsername,omitempty"`
	GithubAvatarURL    *string `json:"githubAvatarUrl,omitempty"`
	GithubNoreplyEmail *string `json:"githubNoreplyEmail,omitempty"`
	CommitAuthorName   *string `json:"commitAuthorName,omitempty"`
	OAuthStatus        *string `json:"oauthStatus,omitempty"`
}
