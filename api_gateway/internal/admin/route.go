package admin

import (
	"net/http"
)

// RegisterAdminRoutes registers all admin CRUD routes on the main mux under /admin/ prefix.
func RegisterAdminRoutes(mux *http.ServeMux, h *AdminHandler, jwtSecret string) {
	adminMW := AdminAuth(jwtSecret)
	adminMux := http.NewServeMux()

	// Languages
	adminMux.HandleFunc("POST /languages", h.CreateLanguage)
	adminMux.HandleFunc("PUT /languages/{id}", h.UpdateLanguage)
	adminMux.HandleFunc("GET /languages", h.ListLanguages)
	adminMux.HandleFunc("GET /languages/{id}", h.GetLanguage)

	// Subjects
	adminMux.HandleFunc("POST /subjects", h.CreateSubject)
	adminMux.HandleFunc("PUT /subjects/{id}", h.UpdateSubject)
	adminMux.HandleFunc("DELETE /subjects/{id}", h.DeleteSubject)
	adminMux.HandleFunc("GET /subjects", h.ListSubjects)
	adminMux.HandleFunc("GET /subjects/{id}", h.GetSubject)

	// Sessions
	adminMux.HandleFunc("POST /subjects/{subjectId}/sessions", h.CreateSession)
	adminMux.HandleFunc("PUT /sessions/{id}", h.UpdateSession)
	adminMux.HandleFunc("DELETE /sessions/{id}", h.DeleteSession)

	// Lessons
	adminMux.HandleFunc("POST /sessions/{sessionId}/lessons", h.CreateLesson)
	adminMux.HandleFunc("PUT /lessons/{id}", h.UpdateLesson)
	adminMux.HandleFunc("DELETE /lessons/{id}", h.DeleteLesson)

	// Problems
	adminMux.HandleFunc("POST /problems", h.CreateProblem)
	adminMux.HandleFunc("PUT /problems/{id}", h.UpdateProblem)
	adminMux.HandleFunc("DELETE /problems/{id}", h.DeleteProblem)
	adminMux.HandleFunc("GET /problems", h.ListProblems)
	adminMux.HandleFunc("GET /problems/{id}", h.GetProblem)

	// Lesson-Problems
	adminMux.HandleFunc("POST /lessons/{lessonId}/problems", h.CreateLessonProblem)
	adminMux.HandleFunc("DELETE /lessons/{lessonId}/problems/{problemId}", h.DeleteLessonProblem)

	// Test Cases
	adminMux.HandleFunc("POST /problems/{problemId}/test-cases", h.CreateTestCase)
	adminMux.HandleFunc("PUT /test-cases/{id}", h.UpdateTestCase)
	adminMux.HandleFunc("DELETE /test-cases/{id}", h.DeleteTestCase)
	adminMux.HandleFunc("GET /problems/{problemId}/test-cases", h.ListTestCases)

	// Tags
	adminMux.HandleFunc("POST /tags", h.CreateTag)
	adminMux.HandleFunc("GET /tags", h.ListTags)

	// Problem-Tags
	adminMux.HandleFunc("POST /problems/{problemId}/tags", h.CreateProblemTag)
	adminMux.HandleFunc("DELETE /problems/{problemId}/tags/{tagId}", h.DeleteProblemTag)

	mux.Handle("/admin/", http.StripPrefix("/admin", adminMW(adminMux)))
}
