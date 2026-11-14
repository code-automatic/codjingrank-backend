package handler

import (
	"innogen-backend/shared/middleware"
	"innogen-backend/submission_service/internal/dto"
	"math"
	"net/url"
	"strconv"
)

func historyQuery(v url.Values) (dto.HistoryQuery, string) {
	q := dto.HistoryQuery{Page: 1, Size: 20, ProblemID: v.Get("problemId"), LanguageID: v.Get("languageId"), Status: v.Get("status")}
	for _, x := range []struct {
		name string
		dest *int
	}{{"page", &q.Page}, {"size", &q.Size}} {
		if v.Has(x.name) {
			n, e := strconv.Atoi(v.Get(x.name))
			if e != nil || n < 1 {
				return q, x.name
			}
			*x.dest = n
		}
	}
	if q.Size > 100 {
		return q, "size"
	}
	if q.Page-1 > math.MaxInt/q.Size {
		return q, "page"
	}
	for _, name := range []string{"problemId", "languageId"} {
		if v.Has(name) && !middleware.ValidUUID(v.Get(name)) {
			return q, name
		}
	}
	if v.Has("status") {
		switch q.Status {
		case "Pending", "Running", "Accepted", "Wrong Answer", "Time Limit Exceeded", "Memory Limit Exceeded", "Compilation Error", "Runtime Error", "Internal Error":
		default:
			return q, "status"
		}
	}
	return q, ""
}
