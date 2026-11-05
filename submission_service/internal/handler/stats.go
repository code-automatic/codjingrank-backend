package handler

import (
	"net/http"

	"innogen-backend/shared/middleware"
	"innogen-backend/shared/response"
)

// GetStats handles GET /me/stats.
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		response.ErrorSimple(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	stats, err := h.svc.GetUserStats(r.Context(), userID)
	if err != nil {
		h.log.Error("failed to get user stats", "error", err.Error())
		response.ErrorSimple(w, http.StatusInternalServerError, "Failed to get user stats")
		return
	}

	response.Success(w, http.StatusOK, stats, "User stats retrieved successfully")
}
