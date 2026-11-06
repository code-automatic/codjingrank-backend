package admin

import (
	"net/http"

	"innogen-backend/shared/response"
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

	// Problems
	adminMux.HandleFunc("POST /problems", h.CreateProblem)
	adminMux.HandleFunc("PUT /problems/{id}", h.UpdateProblem)
	adminMux.HandleFunc("DELETE /problems/{id}", h.DeleteProblem)
	adminMux.HandleFunc("GET /problems", h.ListProblems)
	adminMux.HandleFunc("GET /problems/{id}", h.GetProblem)

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

	mux.Handle("/admin/", http.StripPrefix("/admin", adminMW(response.Router(adminMux))))
}
