package leaderboard

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"innogen-backend/shared/middleware"
	"innogen-backend/shared/response"
)

var slugPattern = regexp.MustCompile("^[a-z0-9]+(-[a-z0-9]+)*$")

func Register(mux *http.ServeMux, db *pgxpool.Pool, secret string, log *slog.Logger) {
	auth := middleware.Auth(secret)
	list := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rank(w, r, db, log, false)
	})
	mux.HandleFunc("GET /api/v1/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		if len(r.Header.Values("Authorization")) != 0 {
			auth(list).ServeHTTP(w, r)
			return
		}
		list.ServeHTTP(w, r)
	})
	mux.Handle("GET /api/v1/leaderboard/me", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rank(w, r, db, log, true)
	})))
}

func bounds(scope string, now time.Time) (*time.Time, *time.Time) {
	now = now.UTC()
	switch scope {
	case "week":
		start := time.Date(now.Year(), now.Month(), now.Day()-(int(now.Weekday())+6)%7, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 0, 7)
		return &start, &end
	case "month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, 0)
		return &start, &end
	default:
		return nil, nil
	}
}

// Count and page use one snapshot, including when the requested page is empty.
const ranked = `WITH counts AS (
	SELECT u.id, u.username, u.avatar_url, count(*) solved, sum(pr.accepted_count) accepted
	FROM user_problem_progress pr
	JOIN users u ON u.id=pr.user_id AND u.is_active
	JOIN problems p ON p.id=pr.problem_id
	JOIN levels l ON l.id=p.level_id
	WHERE ($1::uuid IS NULL OR l.package_id=$1)
	  AND ($2::timestamptz IS NULL OR (pr.first_solved_at >= $2 AND pr.first_solved_at < $3::timestamptz))
	GROUP BY u.id
), ranked AS (
	SELECT *, rank() OVER (ORDER BY solved DESC, accepted DESC) position FROM counts
), entries AS (
	SELECT id, position, jsonb_build_object('rank',position,'userId',id,'username',username,
		'avatarUrl',avatar_url,'solvedProblemCount',solved,'acceptedSolutionCount',accepted) entry
	FROM ranked
) `

func rank(w http.ResponseWriter, r *http.Request, db *pgxpool.Pool, log *slog.Logger, mine bool) {
	q := r.URL.Query()
	scope, slug := q.Get("scope"), q.Get("package")
	if !q.Has("scope") {
		scope = "all"
	}
	page, size := 1, 50
	valid := scope == "all" || scope == "week" || scope == "month"
	if !mine {
		for key, dest := range map[string]*int{"page": &page, "size": &size} {
			if q.Has(key) {
				value, err := strconv.Atoi(q.Get(key))
				valid = valid && err == nil
				*dest = value
			}
		}
	}
	if !valid || page < 1 || page > 2147483647/100 || size < 1 || size > 100 ||
		(q.Has("package") && (len(slug) > 200 || !slugPattern.MatchString(slug))) {
		response.ErrorSimple(w, 400, "Invalid leaderboard filters")
		return
	}
	ctx := r.Context()
	fail := func(err error) {
		log.ErrorContext(ctx, "leaderboard request failed", "error", err)
		response.ErrorSimple(w, 500, "Internal server error")
	}
	user, signedIn := middleware.GetUserID(r)
	if !signedIn && (scope != "all" || page != 1 || size > 50) {
		response.ErrorSimple(w, 401, "Authorization required for this leaderboard")
		return
	}
	if signedIn {
		var active bool
		err := db.QueryRow(ctx, "SELECT is_active FROM users WHERE id=$1", user).Scan(&active)
		if errors.Is(err, pgx.ErrNoRows) {
			response.ErrorSimple(w, 401, "Unauthorized")
			return
		}
		if err != nil {
			fail(err)
			return
		}
		if !active {
			response.Error(w, 403, "Account is inactive", "ACCOUNT_INACTIVE")
			return
		}
	}
	var packageID, packageSlug *string
	if slug != "" {
		err := db.QueryRow(ctx, "SELECT id::text FROM packages WHERE slug=$1 AND status='published'", slug).Scan(&packageID)
		if errors.Is(err, pgx.ErrNoRows) {
			response.ErrorSimple(w, 404, "Package not found")
			return
		}
		if err != nil {
			fail(err)
			return
		}
		packageSlug = &slug
	}
	start, end := bounds(scope, time.Now())
	data := map[string]any{"scope": scope, "packageSlug": packageSlug, "periodStart": start, "periodEnd": end}
	if mine {
		var entry json.RawMessage
		err := db.QueryRow(ctx, ranked+"SELECT (SELECT entry FROM entries WHERE id=$4)", packageID, start, end, user).Scan(&entry)
		if err != nil {
			fail(err)
			return
		}
		data["entry"] = entry
		response.Success(w, 200, data, "Rank retrieved")
		return
	}
	var items json.RawMessage
	var total int64
	err := db.QueryRow(ctx, ranked+`SELECT (SELECT count(*) FROM entries),
		COALESCE(jsonb_agg(entry ORDER BY position,id),'[]')
		FROM (SELECT * FROM entries ORDER BY position,id LIMIT $4 OFFSET $5) page`,
		packageID, start, end, size, (page-1)*size).Scan(&total, &items)
	if err != nil {
		fail(err)
		return
	}
	data["items"], data["total"], data["page"], data["size"] = items, total, page, size
	response.Success(w, 200, data, "Leaderboard retrieved")
}
