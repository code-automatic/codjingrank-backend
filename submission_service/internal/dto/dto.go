package dto

import (
	"innogen-backend/shared/middleware"
	"strings"
	"time"

	"innogen-backend/shared/models"
)

// SubmitRequest is the JSON body for POST /submit.
type SubmitRequest struct {
	ProblemID  string `json:"problemId"`
	LanguageID string `json:"languageId"`
	Code       string `json:"code"`
}

// Validate returns a field name and an error message if invalid, or empty strings if valid.
func (r *SubmitRequest) Validate() (string, string) {
	if !middleware.ValidUUID(r.ProblemID) {
		return "problemId", "Invalid problem ID"
	}
	if !middleware.ValidUUID(r.LanguageID) {
		return "languageId", "Invalid language ID"
	}
	if strings.TrimSpace(r.Code) == "" {
		return "code", "Code must not be empty"
	}
	if len(r.Code) > 102400 {
		return "code", "Code exceeds maximum length"
	}
	return "", ""
}

// SubmitResponse wraps a submission for POST /submit response.
type SubmitResponse struct {
	Submission *models.Submission `json:"submission"`
}

// GetSubmissionResponse wraps a submission for GET /submissions/{id} response.
type GetSubmissionResponse struct {
	Submission *models.Submission `json:"submission"`
}

// SubmissionListItem is a submission without the code field, for list endpoints.
type SubmissionListItem struct {
	ID             string  `json:"id"`
	UserID         string  `json:"userId"`
	ProblemID      string  `json:"problemId"`
	LanguageID     string  `json:"languageId"`
	ProblemSlug    string  `json:"problemSlug"`
	ProblemTitle   string  `json:"problemTitle"`
	LanguageSlug   string  `json:"languageSlug"`
	Status         string  `json:"status"`
	RuntimeMs      *int    `json:"runtimeMs"`
	MemoryKb       *int    `json:"memoryKb"`
	ErrorMessage   *string `json:"errorMessage"`
	PassCount      int     `json:"passCount"`
	TotalTestcases int     `json:"totalTestcases"`
	CreatedAt      string  `json:"createdAt"`
	JudgedAt       *string `json:"judgedAt"`
}

// ToSubmissionListItem converts a models.Submission to a SubmissionListItem.
func ToSubmissionListItem(s *models.Submission) SubmissionListItem {
	item := SubmissionListItem{
		ID:             s.ID,
		UserID:         s.UserID,
		ProblemID:      s.ProblemID,
		LanguageID:     s.LanguageID,
		ProblemSlug:    s.ProblemSlug,
		ProblemTitle:   s.ProblemTitle,
		LanguageSlug:   s.LanguageSlug,
		Status:         s.Status,
		RuntimeMs:      s.RuntimeMs,
		MemoryKb:       s.MemoryKb,
		ErrorMessage:   s.ErrorMessage,
		PassCount:      s.PassCount,
		TotalTestcases: s.TotalTestcases,
		CreatedAt:      s.CreatedAt.UTC().Format(time.RFC3339),
	}
	if s.JudgedAt != nil {
		js := s.JudgedAt.UTC().Format(time.RFC3339)
		item.JudgedAt = &js
	}
	return item
}

// ListSubmissionsResponse wraps submission list for GET /me/submissions.
type ListSubmissionsResponse struct {
	Items []SubmissionListItem `json:"items"`
	Page  int                  `json:"page"`
	Size  int                  `json:"size"`
	Total int                  `json:"total"`
}

type HistoryQuery struct {
	Page, Size                    int
	ProblemID, LanguageID, Status string
}

// LatestSubmissionResponse wraps a submission for GET /me/submissions/{problemId}/latest.
type LatestSubmissionResponse struct {
	Submission *models.Submission `json:"submission"`
}

// UserStatsResponse wraps the user statistics for GET /me/stats.
type UserStatsResponse struct {
	Streak      Streak      `json:"streak"`
	SolvedCount SolvedCount `json:"solvedCount"`
	Rank        Rank        `json:"rank"`
	Activity    []Activity  `json:"activity"`
}

type Streak struct {
	CurrentStreak  int     `json:"currentStreak"`
	MaxStreak      int     `json:"maxStreak"`
	LastActiveDate *string `json:"lastActiveDate"`
}

type SolvedCount struct {
	Total  int `json:"total"`
	Easy   int `json:"easy"`
	Medium int `json:"medium"`
	Hard   int `json:"hard"`
}

type Rank struct {
	CurrentRank *int `json:"currentRank"`
	TotalUsers  int  `json:"totalUsers"`
}

type Activity struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}
