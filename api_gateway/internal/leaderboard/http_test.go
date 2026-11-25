package leaderboard

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"innogen-backend/shared/testutil"
)

func TestGuestRestrictions(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, nil, "test-secret", slog.Default())
	for path, want := range map[string]int{
		"/api/v1/leaderboard?scope=week": 401,
		"/api/v1/leaderboard?page=2":     401,
		"/api/v1/leaderboard?size=51":    401,
		"/api/v1/leaderboard?size=0":     400,
		"/api/v1/leaderboard/me":         401,
	} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != want {
			t.Errorf("%s: got %d, want %d", path, rec.Code, want)
		}
	}
}

func TestRankUsesDistinctProgressAndKeepsTiesAcrossPages(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.Background()
	var slug string
	if err := f.Pool.QueryRow(ctx, "SELECT slug FROM packages WHERE id=$1", f.PackageID).Scan(&slug); err != nil {
		t.Fatal(err)
	}
	ids := []string{f.UserID}
	for _, suffix := range []string{"tie", "third", "inactive", "empty"} {
		var id string
		err := f.Pool.QueryRow(ctx, `INSERT INTO users(email,username,password_hash,is_active)
			VALUES($1||'@example.org',$1,'test',$2) RETURNING id`, slug+"-"+suffix, suffix != "inactive").Scan(&id)
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
		t.Cleanup(func() {
			if _, err := f.Pool.Exec(ctx, "DELETE FROM user_problem_progress WHERE user_id=$1", id); err != nil {
				t.Error(err)
			}
			if _, err := f.Pool.Exec(ctx, "DELETE FROM users WHERE id=$1", id); err != nil {
				t.Error(err)
			}
		})
	}
	for i, count := range []int{3, 3, 2, 10} {
		if _, err := f.Pool.Exec(ctx, `INSERT INTO user_problem_progress
			(user_id,problem_id,first_solved_at,last_solved_at,accepted_count)
			VALUES($1,$2,now(),now(),$3)`, ids[i], f.ProblemID, count); err != nil {
			t.Fatal(err)
		}
	}
	mux := http.NewServeMux()
	Register(mux, f.Pool, "test-secret", slog.Default())
	read := func(path, user string, code int) map[string]json.RawMessage {
		t.Helper()
		req := httptest.NewRequest("GET", path, nil)
		if user != "" {
			token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"sub": user, "userId": user, "exp": time.Now().Add(time.Hour).Unix(),
			}).SignedString([]byte("test-secret"))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != code {
			t.Fatalf("%s: got %d, want %d: %s", path, rec.Code, code, rec.Body.String())
		}
		testutil.Capture(t, rec, req.Method, req.Pattern)
		var envelope struct{ Data map[string]json.RawMessage }
		if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		return envelope.Data
	}
	list := "/api/v1/leaderboard?package=" + slug
	mine := "/api/v1/leaderboard/me?package=" + slug
	data := read(list+"&size=2", "", 200)
	var entries []struct{ Rank int }
	if err := json.Unmarshal(data["items"], &entries); err != nil || len(entries) != 2 ||
		entries[0].Rank != 1 || entries[1].Rank != 1 || string(data["total"]) != "3" {
		t.Fatalf("expected two tied users and three active entries: %s %s", data["items"], data["total"])
	}
	data = read(mine+"&scope=week", ids[2], 200)
	var entry struct{ Rank, AcceptedSolutionCount int }
	if err := json.Unmarshal(data["entry"], &entry); err != nil || entry.Rank != 3 || entry.AcceptedSolutionCount != 2 {
		t.Fatalf("expected third user outside first page: %s", data["entry"])
	}
	if got := read(mine, ids[4], 200)["entry"]; string(got) != "null" {
		t.Fatalf("expected no rank: %s", got)
	}
	data = read(list+"&page=4", ids[0], 200)
	if string(data["items"]) != "[]" || string(data["total"]) != "3" {
		t.Fatalf("expected empty page with total retained: %s %s", data["items"], data["total"])
	}
}

func TestUTCWeekAcrossYear(t *testing.T) {
	now := time.Date(2027, 1, 4, 0, 30, 0, 0, time.FixedZone("UTC+7", 7*3600))
	start, end := bounds("week", now)
	if start.Format(time.RFC3339) != "2026-12-28T00:00:00Z" || end.Format(time.RFC3339) != "2027-01-04T00:00:00Z" {
		t.Fatalf("unexpected week: %v to %v", start, end)
	}
}
