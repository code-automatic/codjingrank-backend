package githubapp

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// RealClient implements GitHubClient using the GitHub REST API.
type RealClient struct {
	appID      string
	privateKey *rsa.PrivateKey
	baseURL    string
	httpClient *http.Client
}

// NewRealClient creates a new GitHub App client.
// appID is the GitHub App ID.
// privateKeyPath is the path to the PEM-encoded RSA private key.
func NewRealClient(appID, privateKeyPath, baseURL string) (*RealClient, error) {
	keyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("githubapp: failed to read private key: %w", err)
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("githubapp: failed to parse PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8
		key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, fmt.Errorf("githubapp: failed to parse private key (PKCS1: %v, PKCS8: %v)", err, err2)
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("githubapp: key is not an RSA private key")
		}
	}

	return &RealClient{
		appID:      appID,
		privateKey: privateKey,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// generateJWT creates a GitHub App JWT signed with RS256.
func (c *RealClient) generateJWT() (string, error) {
	now := time.Now()
	claims := &jwtlib.RegisteredClaims{
		Issuer:    c.appID,
		IssuedAt:  jwtlib.NewNumericDate(now.Add(-60 * time.Second)),
		ExpiresAt: jwtlib.NewNumericDate(now.Add(10 * time.Minute)),
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodRS256, claims)
	return token.SignedString(c.privateKey)
}

// GetInstallation retrieves installation details (owner login and type) from GitHub.
// Uses a JWT for App authentication as required by the GitHub API.
func (c *RealClient) GetInstallation(ctx context.Context, installationID string) (string, string, error) {
	jwt, err := c.generateJWT()
	if err != nil {
		return "", "", fmt.Errorf("githubapp: failed to generate JWT: %w", err)
	}

	url := fmt.Sprintf("%s/app/installations/%s", c.baseURL, installationID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("githubapp: get installation failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("githubapp: installation %s not found (status %d): %s", installationID, resp.StatusCode, string(body))
	}

	var result struct {
		Account struct {
			Login string `json:"login"`
			Type  string `json:"type"`
		} `json:"account"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("githubapp: decode installation failed: %w", err)
	}

	return result.Account.Login, result.Account.Type, nil
}

// GetInstallationToken exchanges a JWT for an installation access token.
func (c *RealClient) GetInstallationToken(ctx context.Context, installationID string) (string, error) {
	jwt, err := c.generateJWT()
	if err != nil {
		return "", fmt.Errorf("githubapp: failed to generate JWT: %w", err)
	}

	url := fmt.Sprintf("%s/app/installations/%s/access_tokens", c.baseURL, installationID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("githubapp: token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("githubapp: token exchange failed (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("githubapp: failed to decode token response: %w", err)
	}
	return result.Token, nil
}

// EnsureRepo checks if a repository exists and creates it if not.
func (c *RealClient) EnsureRepo(ctx context.Context, token, owner, repoName, ownerType, defaultBranch string) (*RepoInfo, error) {
	// Try to get existing repo
	url := fmt.Sprintf("%s/repos/%s/%s", c.baseURL, owner, repoName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("githubapp: repo lookup failed: %w", err)
	}

	if resp.StatusCode == http.StatusOK {
		var repo struct {
			FullName      string `json:"full_name"`
			HTMLURL       string `json:"html_url"`
			ID            int64  `json:"id"`
			DefaultBranch string `json:"default_branch"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("githubapp: failed to decode repo: %w", err)
		}
		resp.Body.Close()
		return &RepoInfo{
			FullName:      repo.FullName,
			URL:           repo.HTMLURL,
			GithubRepoID:  fmt.Sprintf("%d", repo.ID),
			DefaultBranch: repo.DefaultBranch,
		}, nil
	}
	resp.Body.Close()

	// Create repo
	body := map[string]interface{}{
		"name":           repoName,
		"private":        true,
		"auto_init":      true,
		"default_branch": defaultBranch,
	}
	bodyBytes, _ := json.Marshal(body)

	var createURL string
	if ownerType == "Organization" {
		createURL = fmt.Sprintf("%s/orgs/%s/repos", c.baseURL, owner)
	} else {
		createURL = fmt.Sprintf("%s/user/repos", c.baseURL)
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, createURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("githubapp: create repo failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("githubapp: create repo failed (status %d): %s", resp.StatusCode, string(body))
	}

	var repo struct {
		FullName      string `json:"full_name"`
		HTMLURL       string `json:"html_url"`
		ID            int64  `json:"id"`
		DefaultBranch string `json:"default_branch"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return nil, fmt.Errorf("githubapp: failed to decode created repo: %w", err)
	}

	return &RepoInfo{
		FullName:      repo.FullName,
		URL:           repo.HTMLURL,
		GithubRepoID:  fmt.Sprintf("%d", repo.ID),
		DefaultBranch: repo.DefaultBranch,
	}, nil
}

// GetFileContent retrieves file content from GitHub.
func (c *RealClient) GetFileContent(ctx context.Context, token, owner, repoName, filePath, branch string) (*FileInfo, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s?ref=%s", c.baseURL, owner, repoName, filePath, branch)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("githubapp: get file failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("githubapp: get file failed (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		SHA     string `json:"sha"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("githubapp: failed to decode file: %w", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(result.Content)
	if err != nil {
		return nil, fmt.Errorf("githubapp: failed to decode file content: %w", err)
	}

	return &FileInfo{SHA: result.SHA, Content: string(decoded)}, nil
}

// CreateOrUpdateFile creates or updates a file on GitHub.
func (c *RealClient) CreateOrUpdateFile(ctx context.Context, token, owner, repoName, filePath, branch, content, commitMessage string, existingSHA *string, authorName, authorEmail string) (*CommitResult, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.baseURL, owner, repoName, filePath)

	body := map[string]interface{}{
		"message": commitMessage,
		"content": base64.StdEncoding.EncodeToString([]byte(content)),
		"branch":  branch,
	}
	if existingSHA != nil {
		body["sha"] = *existingSHA
	}
	if authorName != "" && authorEmail != "" {
		body["author"] = map[string]string{
			"name":  authorName,
			"email": authorEmail,
		}
	}

	bodyBytes, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("githubapp: commit file failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("githubapp: commit file failed (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Content struct {
			SHA string `json:"sha"`
		} `json:"content"`
		Commit struct {
			SHA     string `json:"sha"`
			HTMLURL string `json:"html_url"`
		} `json:"commit"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("githubapp: failed to decode commit result: %w", err)
	}

	return &CommitResult{
		ContentSHA: result.Content.SHA,
		CommitSHA:  result.Commit.SHA,
		CommitURL:  result.Commit.HTMLURL,
	}, nil
}
