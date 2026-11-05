package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type OAuthClient struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

func NewOAuthClient(clientID, clientSecret string) *OAuthClient {
	return &OAuthClient{clientID: clientID, clientSecret: clientSecret, httpClient: &http.Client{Timeout: 30 * time.Second}}
}

func (c *OAuthClient) ExchangeCode(ctx context.Context, code, redirectURL string) (string, error) {
	data := url.Values{
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURL},
	}
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("oauth: token exchange failed: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("oauth: decode failed: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("oauth: %s", result.Error)
	}
	return result.AccessToken, nil
}

type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

func (c *OAuthClient) GetUser(ctx context.Context, accessToken string) (*GitHubUser, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("oauth: get user failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("oauth: get user returned %d", resp.StatusCode)
	}
	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("oauth: decode user failed: %w", err)
	}
	return &user, nil
}
