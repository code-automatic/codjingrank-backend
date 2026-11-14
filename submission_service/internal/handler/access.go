package handler

import (
	"errors"
	"innogen-backend/shared/catalog"
	"innogen-backend/shared/middleware"
	"innogen-backend/shared/response"
	"net/http"
)

func (h *Handler) reader(w http.ResponseWriter, r *http.Request) (string, bool, bool) {
	uid, ok := middleware.GetUserID(r)
	if !ok {
		response.ErrorSimple(w, 401, "Unauthorized")
		return "", false, false
	}
	admin, err := h.svc.ReaderRole(r.Context(), uid)
	if err != nil {
		if errors.Is(err, catalog.ErrAccountInactive) {
			response.Error(w, 403, "Account is inactive", "ACCOUNT_INACTIVE")
		} else {
			response.ErrorSimple(w, 500, "Internal server error")
		}
		return "", false, false
	}
	return uid, admin, true
}
