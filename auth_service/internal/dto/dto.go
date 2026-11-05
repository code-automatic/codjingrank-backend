package dto

import (
	"strings"

	"innogen-backend/shared/models"
)

// LoginRequest is the JSON body for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate returns an error message string if the request is invalid, or empty string if valid.
func (r *LoginRequest) Validate() string {
	if strings.TrimSpace(r.Email) == "" {
		return "Email is required"
	}
	if r.Password == "" {
		return "Password is required"
	}
	return ""
}

// RegisterRequest is the JSON body for POST /auth/register.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
	FullName string `json:"fullName"`
}

// Validate returns an error message string if invalid.
func (r *RegisterRequest) Validate() string {
	if strings.TrimSpace(r.Email) == "" {
		return "Email is required"
	}
	if len(r.Password) < 6 {
		return "Password must be at least 6 characters"
	}
	if strings.TrimSpace(r.Username) == "" {
		return "Username is required"
	}
	return ""
}

// LoginResponse is the JSON body returned by POST /auth/login.
type LoginResponse struct {
	AccessToken string       `json:"accessToken"`
	User        UserResponse `json:"user"`
}

// MeResponse is the JSON body returned by GET /auth/me.
type MeResponse struct {
	User UserResponse `json:"user"`
}

// UserResponse is the public-safe user payload (no password, no timestamps).
type UserResponse struct {
	ID       int     `json:"id"`
	Email    string  `json:"email"`
	Username *string `json:"username"`
	FullName *string `json:"fullName"`
	Role     string  `json:"role"`
}

// ToUserResponse converts a models.User to a UserResponse.
func ToUserResponse(u *models.User) UserResponse {
	return UserResponse{
		ID:       u.ID,
		Email:    u.Email,
		Username: u.Username,
		FullName: u.FullName,
		Role:     u.Role,
	}
}

type GithubConnectResponse struct {
	InstallURL string `json:"installUrl"`
}

type GithubStatusResponse struct {
	Connected       bool   `json:"connected"`
	InstallationID  string `json:"installationId,omitempty"`
	GithubOwner     string `json:"githubOwner,omitempty"`
	GithubOwnerType string `json:"githubOwnerType,omitempty"`
	Status          string `json:"status,omitempty"`
}
