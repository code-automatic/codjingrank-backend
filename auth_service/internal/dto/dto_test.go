package dto

import (
	"encoding/json"
	"testing"
	"time"

	"innogen-backend/shared/models"
)

func TestToUserResponseV1Shape(t *testing.T) {
	username := "bradar"
	fullName := "Trung Hieu"
	avatarURL := "https://example.com/avatar.png"
	bio := "Learning algorithms"
	createdAt := time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)

	got := ToUserResponse(&models.User{
		ID:        "11111111-1111-4111-8111-111111111111",
		Email:     "learner@example.com",
		Username:  &username,
		FullName:  &fullName,
		AvatarURL: &avatarURL,
		Bio:       &bio,
		Role:      "student",
		IsActive:  true,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	})

	body, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(body, &fields); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"id", "email", "username", "fullName", "avatarUrl", "bio", "role", "isActive", "createdAt", "updatedAt"} {
		if _, ok := fields[key]; !ok {
			t.Errorf("missing %q in %s", key, body)
		}
	}
	if _, ok := fields["password"]; ok {
		t.Errorf("password must never be serialized: %s", body)
	}
	if fields["avatarUrl"] != avatarURL || fields["bio"] != bio {
		t.Errorf("profile fields not mapped: %s", body)
	}
	if fields["isActive"] != true {
		t.Errorf("isActive not mapped: %s", body)
	}
	if fields["createdAt"] != "2026-11-07T00:00:00Z" {
		t.Errorf("createdAt = %v, want RFC3339", fields["createdAt"])
	}
	if fields["role"] != "student" {
		t.Errorf("role = %v", fields["role"])
	}
}

func TestToUserResponseNullableProfileFields(t *testing.T) {
	got := ToUserResponse(&models.User{
		ID:    "11111111-1111-4111-8111-111111111111",
		Email: "learner@example.com",
		Role:  "student",
	})
	body, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(body, &fields); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"username", "fullName", "avatarUrl", "bio"} {
		value, ok := fields[key]
		if !ok || value != nil {
			t.Errorf("%q = %v, want null", key, value)
		}
	}
}
