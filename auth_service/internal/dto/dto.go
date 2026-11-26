package dto

import (
	"strings"
	"time"

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

// UserResponse is the public-safe user payload (no password, no secrets).
// It matches the v1 User schema: RFC3339 timestamps, nullable profile fields.
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  *string   `json:"username"`
	FullName  *string   `json:"fullName"`
	AvatarURL *string   `json:"avatarUrl"`
	Bio       *string   `json:"bio"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ToUserResponse converts a models.User to a UserResponse.
func ToUserResponse(u *models.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		FullName:  u.FullName,
		AvatarURL: u.AvatarURL,
		Bio:       u.Bio,
		Role:      u.Role,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
