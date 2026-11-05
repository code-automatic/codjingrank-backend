package admin

import (
	"net/http"

	"innogen-backend/shared/middleware"
	"innogen-backend/shared/response"
)

// AdminAuth returns middleware that validates JWT and checks for admin role.
// Returns 401 for missing/invalid JWT, 403 for non-admin users.
func AdminAuth(jwtSecret string) func(http.Handler) http.Handler {
	authMW := middleware.Auth(jwtSecret)
	return func(next http.Handler) http.Handler {
		return authMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := middleware.GetUserRole(r)
			if !ok || role != "admin" {
				response.ErrorSimple(w, http.StatusForbidden, "Admin access required")
				return
			}
			next.ServeHTTP(w, r)
		}))
	}
}
