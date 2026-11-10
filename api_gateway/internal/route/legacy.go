package route

import (
	"innogen-backend/shared/response"
	"net/http"
)

// RegisterRetiredCatalog closes legacy integer catalog routes instead of exposing a policy bypass.
func RegisterRetiredCatalog(mux *http.ServeMux) {
	gone := func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusGone, "Use the UUID /api/v1 catalog contract", "LEGACY_ROUTE_RETIRED")
	}
	for _, path := range []string{"/problems", "/problems/", "/admin", "/admin/"} {
		mux.HandleFunc(path, gone)
	}
}
