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

// ForgotPasswordRequest is the JSON body for POST /auth/forgot-password.
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// Validate returns an error message string if invalid.
func (r *ForgotPasswordRequest) Validate() string {
	if strings.TrimSpace(r.Email) == "" {
		return "Email is required"
	}
	return ""
}

// ResetPasswordRequest is the JSON body for POST /auth/reset-password.
type ResetPasswordRequest struct {
	Email       string `json:"email"`
	OTP         string `json:"otp"`
	NewPassword string `json:"newPassword"`
}

// Validate returns an error message string if invalid. It mirrors Register's
// minimum password policy.
func (r *ResetPasswordRequest) Validate() string {
	if strings.TrimSpace(r.Email) == "" {
		return "Email is required"
	}
	if strings.TrimSpace(r.OTP) == "" {
		return "OTP is required"
	}
	return ValidatePasswordPolicy(r.NewPassword)
}

// ValidatePasswordPolicy enforces the shared password contract:
// 8-72 UTF-8 bytes, at least one ASCII letter and one digit. The byte bound
// keeps bcrypt from silently truncating (or erroring on) long inputs.
func ValidatePasswordPolicy(password string) string {
	if len(password) < 8 {
		return "Password must be at least 8 characters"
	}
	if len(password) > 72 {
		return "Password must be at most 72 bytes"
	}
	var hasLetter, hasDigit bool
	for _, r := range password {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
			hasLetter = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return "Password must contain at least one letter and one digit"
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
