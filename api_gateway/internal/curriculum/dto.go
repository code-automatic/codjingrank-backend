package curriculum

import (
	"innogen-backend/shared/models"
)

// SubjectDTO is the response DTO for a subject.
type SubjectDTO struct {
	ID         int     `json:"id"`
	Title      string  `json:"title"`
	Slug       string  `json:"slug"`
	Color      *string `json:"color"`
	LanguageID *int    `json:"languageId"`
}

// SessionDTO is the response DTO for a subject session.
type SessionDTO struct {
	ID          int     `json:"id"`
	SubjectID   int     `json:"subjectId"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	OrderIndex  int     `json:"orderIndex"`
}

// LessonDTO is the response DTO for a lesson.
type LessonDTO struct {
	ID               int     `json:"id"`
	SubjectSessionID int     `json:"subjectSessionId"`
	Title            string  `json:"title"`
	ContentMD        *string `json:"contentMd"`
	OrderIndex       int     `json:"orderIndex"`
}

// ProblemListItem is the response DTO for a problem within a lesson listing.
type ProblemListItem struct {
	ID             int     `json:"id"`
	Slug           string  `json:"slug"`
	Title          string  `json:"title"`
	Difficulty     string  `json:"difficulty"`
	OrderIndex     int     `json:"orderIndex"`
	AcceptanceRate float64 `json:"acceptanceRate"`
}

// --- Conversion helpers ---

func subjectToDTO(s models.Subject) SubjectDTO {
	return SubjectDTO{
		ID:         s.ID,
		Title:      s.Title,
		Slug:       s.Slug,
		Color:      s.Color,
		LanguageID: s.LanguageID,
	}
}

func sessionToDTO(s models.SubjectSession) SessionDTO {
	return SessionDTO{
		ID:          s.ID,
		SubjectID:   s.SubjectID,
		Title:       s.Title,
		Description: s.Description,
		OrderIndex:  s.OrderIndex,
	}
}

func lessonToDTO(l models.Lesson) LessonDTO {
	return LessonDTO{
		ID:               l.ID,
		SubjectSessionID: l.SubjectSessionID,
		Title:            l.Title,
		ContentMD:        l.ContentMD,
		OrderIndex:       l.OrderIndex,
	}
}
