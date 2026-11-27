package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"innogen-backend/shared/audit"
	"innogen-backend/shared/testutil"
)

func adminTestToken(t *testing.T, userID, email, role string) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID, "userId": userID, "email": email, "role": role,
		"exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func adminCall(t *testing.T, mux *http.ServeMux, token, method, path, body string, want int) (map[string]any, []byte) {
	t.Helper()
	req := httptest.NewRequest(method, "/api/v1/admin"+path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != want {
		t.Fatalf("%s %s: got %d want %d: %s", method, path, rec.Code, want, rec.Body.String())
	}
	if rec.Body.Len() == 0 {
		return nil, rec.Body.Bytes()
	}
	var env map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("%s %s: invalid envelope: %v", method, path, err)
	}
	return env, rec.Body.Bytes()
}

func TestAdminV1UsersStatsAuditRoutes(t *testing.T) {
	mux := http.NewServeMux()
	RegisterV1(mux, nil, "test-secret", slog.New(slog.NewTextHandler(io.Discard, nil)))
	const id = "00000000-0000-4000-8000-000000000000"
	for _, tc := range []struct{ method, path, pattern string }{
		{"GET", "/api/v1/admin/users", "GET /api/v1/admin/users"},
		{"GET", "/api/v1/admin/users/" + id, "GET /api/v1/admin/users/{id}"},
		{"PATCH", "/api/v1/admin/users/" + id, "PATCH /api/v1/admin/users/{id}"},
		{"GET", "/api/v1/admin/stats", "GET /api/v1/admin/stats"},
		{"GET", "/api/v1/admin/audit-logs", "GET /api/v1/admin/audit-logs"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		_, pattern := mux.Handler(req)
		if pattern != tc.pattern {
			t.Fatalf("%s %s mapped to %q, want %q", tc.method, tc.path, pattern, tc.pattern)
		}
	}
	req := httptest.NewRequest("GET", "/api/v1/admin/users", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != 401 {
		t.Fatalf("admin users without token: got %d want 401", rec.Code)
	}
}

func TestAdminUsersStatsAudit(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.Background()
	var adminEmail string
	if err := f.Pool.QueryRow(ctx, `UPDATE users SET role='admin' WHERE id=$1 RETURNING email`, f.UserID).Scan(&adminEmail); err != nil {
		t.Fatal(err)
	}

	var suffix, targetID, targetEmail, targetUsername string
	if err := f.Pool.QueryRow(ctx, `SELECT gen_random_uuid()::text`).Scan(&suffix); err != nil {
		t.Fatal(err)
	}
	targetEmail = "admin-target-" + suffix + "@example.org"
	targetUsername = "admin-target-" + suffix
	if err := f.Pool.QueryRow(ctx, `INSERT INTO users(email,username,password_hash,full_name,role) VALUES($1,$2,'hash','Target User','student') RETURNING id`, targetEmail, targetUsername).Scan(&targetID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := f.Pool.Exec(ctx, `DELETE FROM audit_logs WHERE resource_id=$1 OR action='test.audit.write'`, targetID); err != nil {
			t.Error(err)
		}
		if _, err := f.Pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, targetID); err != nil {
			t.Error(err)
		}
	})

	if _, err := f.Pool.Exec(ctx, `INSERT INTO submissions(user_id,problem_id,language_id,source_code,source_hash,status)
		VALUES($1,$2,$3,'print(1)',$4,'Accepted')`, f.UserID, f.ProblemID, f.LanguageID, strings.Repeat("a", 64)); err != nil {
		t.Fatal(err)
	}

	adminToken := adminTestToken(t, f.UserID, adminEmail, "admin")
	studentToken := adminTestToken(t, targetID, targetEmail, "student")
	mux := http.NewServeMux()
	RegisterV1(mux, f.Pool, "test-secret", slog.New(slog.NewTextHandler(io.Discard, nil)))

	listPath := "/users?q=" + url.QueryEscape(targetUsername) + "&role=student&isActive=true&page=1&size=1"
	env, raw := adminCall(t, mux, adminToken, "GET", listPath, "", 200)
	if bytes.Contains(raw, []byte("password")) || bytes.Contains(raw, []byte("refresh")) {
		t.Fatalf("admin user list leaks credentials: %s", raw)
	}
	data := env["data"].(map[string]any)
	items := data["items"].([]any)
	if len(items) != 1 || data["total"] != float64(1) || data["page"] != float64(1) || data["size"] != float64(1) {
		t.Fatalf("unexpected user page: %v", data)
	}
	user := items[0].(map[string]any)
	if user["id"] != targetID || user["role"] != "student" || user["fullName"] != "Target User" {
		t.Fatalf("unexpected user: %v", user)
	}

	env, _ = adminCall(t, mux, adminToken, "GET", "/users/"+targetID, "", 200)
	if got := env["data"].(map[string]any)["user"].(map[string]any)["email"]; got != targetEmail {
		t.Fatalf("wrong user detail: %v", got)
	}
	adminCall(t, mux, adminToken, "GET", "/users/00000000-0000-4000-8000-000000000000", "", 404)
	adminCall(t, mux, adminToken, "GET", "/users/not-a-uuid", "", 400)
	adminCall(t, mux, adminToken, "GET", "/users?size=101", "", 400)
	adminCall(t, mux, adminToken, "GET", "/users?page=0", "", 400)
	adminCall(t, mux, adminToken, "GET", "/users?role=superadmin", "", 400)
	adminCall(t, mux, adminToken, "GET", "/users?isActive=maybe", "", 400)
	for _, wildcard := range []string{"%25", "_"} {
		env, _ = adminCall(t, mux, adminToken, "GET", "/users?q="+wildcard+"&size=100", "", 200)
		if total := env["data"].(map[string]any)["total"].(float64); total != 0 {
			t.Fatalf("wildcard q=%q returned %v users", wildcard, total)
		}
	}

	refreshHash := "admin-refresh-" + suffix
	if _, err := f.Pool.Exec(ctx, `INSERT INTO refresh_tokens(user_id,token_hash,expires_at) VALUES($1,$2,now()+interval '1 day')`, targetID, refreshHash); err != nil {
		t.Fatal(err)
	}

	env, _ = adminCall(t, mux, adminToken, "PATCH", "/users/"+targetID, `{"role":"teacher","isActive":false,"fullName":"Updated User"}`, 200)
	patched := env["data"].(map[string]any)["user"].(map[string]any)
	if patched["role"] != "teacher" || patched["isActive"] != false || patched["fullName"] != "Updated User" {
		t.Fatalf("patch not applied: %v", patched)
	}
	var revokedAt *time.Time
	if err := f.Pool.QueryRow(ctx, `SELECT revoked_at FROM refresh_tokens WHERE token_hash=$1`, refreshHash).Scan(&revokedAt); err != nil {
		t.Fatal(err)
	}
	if revokedAt == nil {
		t.Fatal("refresh token was not revoked when the user was deactivated")
	}
	adminCall(t, mux, adminToken, "PATCH", "/users/"+targetID, `{"role":"superadmin"}`, 400)
	adminCall(t, mux, adminToken, "PATCH", "/users/"+targetID, `{"isActive":"true"}`, 400)
	adminCall(t, mux, adminToken, "PATCH", "/users/"+targetID, `{"email":"other@example.org"}`, 400)
	adminCall(t, mux, adminToken, "PATCH", "/users/"+targetID, `{}`, 400)
	adminCall(t, mux, adminToken, "PATCH", "/users/00000000-0000-4000-8000-000000000000", `{"role":"teacher"}`, 404)

	env, _ = adminCall(t, mux, adminToken, "PATCH", "/users/"+targetID, `{"fullName":null}`, 200)
	if got := env["data"].(map[string]any)["user"].(map[string]any)["fullName"]; got != nil {
		t.Fatalf("fullName not cleared: %v", got)
	}

	for _, action := range []string{"user.role_changed", "user.locked", "user.updated"} {
		env, _ = adminCall(t, mux, adminToken, "GET", "/audit-logs?action="+action+"&actorEmail="+url.QueryEscape(adminEmail)+"&size=100", "", 200)
		data = env["data"].(map[string]any)
		if data["total"].(float64) < 1 {
			t.Fatalf("missing audit action %s: %v", action, data)
		}
		for _, item := range data["items"].([]any) {
			entry := item.(map[string]any)
			if entry["action"] != action || entry["resourceId"] != targetID {
				t.Fatalf("unexpected audit entry: %v", entry)
			}
		}
	}
	env, _ = adminCall(t, mux, adminToken, "GET", "/audit-logs?action=does.not.exist", "", 200)
	if env["data"].(map[string]any)["total"] != float64(0) {
		t.Fatalf("action filter leaked entries: %v", env["data"])
	}
	adminCall(t, mux, adminToken, "GET", "/audit-logs?size=101", "", 400)

	// An admin cannot change their own role or active status, but may edit their name.
	adminCall(t, mux, adminToken, "PATCH", "/users/"+f.UserID, `{"role":"student"}`, 409)
	adminCall(t, mux, adminToken, "PATCH", "/users/"+f.UserID, `{"isActive":false}`, 409)
	adminCall(t, mux, adminToken, "PATCH", "/users/"+f.UserID, `{"fullName":"Self Updated"}`, 200)

	// Reactivation must not revive tokens revoked by the earlier deactivation.
	adminCall(t, mux, adminToken, "PATCH", "/users/"+targetID, `{"isActive":true}`, 200)
	if err := f.Pool.QueryRow(ctx, `SELECT revoked_at FROM refresh_tokens WHERE token_hash=$1`, refreshHash).Scan(&revokedAt); err != nil {
		t.Fatal(err)
	}
	if revokedAt == nil {
		t.Fatal("refresh token was resurrected after reactivation")
	}

	if err := audit.Write(ctx, f.Pool, audit.AuditEntry{
		ActorID: f.UserID, ActorEmail: adminEmail, Action: "test.audit.write",
		ResourceType: "test", ResourceID: targetID, Detail: map[string]any{"ok": true},
	}); err != nil {
		t.Fatal(err)
	}
	env, _ = adminCall(t, mux, adminToken, "GET", "/audit-logs?action=test.audit.write", "", 200)
	data = env["data"].(map[string]any)
	entry := data["items"].([]any)[0].(map[string]any)
	if data["total"] != float64(1) || entry["detail"].(map[string]any)["ok"] != true {
		t.Fatalf("direct audit write not listed: %v", data)
	}

	// Six distinct solvers on the fixture problem make the global "most solved"
	// leader deterministic: no other test gives one problem more than one solver.
	var solverIDs []string
	for i := 0; i < 6; i++ {
		username := "solver-" + suffix + "-" + strconv.Itoa(i)
		var userID, submissionID string
		if err := f.Pool.QueryRow(ctx, `INSERT INTO users(email,username,password_hash)
			VALUES($1,$2,'hash') RETURNING id`, username+"@example.org", username).Scan(&userID); err != nil {
			t.Fatal(err)
		}
		if err := f.Pool.QueryRow(ctx, `INSERT INTO submissions(user_id,problem_id,language_id,source_code,source_hash,status)
			VALUES($1,$2,$3,'print(1)',$4,'Accepted') RETURNING id`,
			userID, f.ProblemID, f.LanguageID, strings.Repeat("b", 64)).Scan(&submissionID); err != nil {
			t.Fatal(err)
		}
		if _, err := f.Pool.Exec(ctx, `INSERT INTO accepted_solutions(submission_id,user_id,problem_id,language_id,source_hash)
			VALUES($1,$2,$3,$4,$5)`, submissionID, userID, f.ProblemID, f.LanguageID, strings.Repeat("b", 64)); err != nil {
			t.Fatal(err)
		}
		solverIDs = append(solverIDs, userID)
	}
	t.Cleanup(func() {
		for _, q := range []string{
			`DELETE FROM accepted_solutions WHERE user_id = ANY($1::uuid[])`,
			`DELETE FROM submissions WHERE user_id = ANY($1::uuid[])`,
			`DELETE FROM users WHERE id = ANY($1::uuid[])`,
		} {
			if _, err := f.Pool.Exec(ctx, q, solverIDs); err != nil {
				t.Error(err)
			}
		}
	})

	// Other packages run their own fixtures against the same database, so global
	// counters can move while this request is in flight. Bound each counter by
	// snapshots taken on either side of the call instead of comparing it with a
	// query issued after the response.
	snapshot := func() map[string]int64 {
		t.Helper()
		var c struct {
			users, active, packages, publishedProblems            int64
			submissions, accepted, usersWithSolutions, forumPosts int64
		}
		if err := f.Pool.QueryRow(ctx, `SELECT
			(SELECT count(*) FROM users), (SELECT count(*) FROM users WHERE is_active),
			(SELECT count(*) FROM packages),
			(SELECT count(*) FROM problems WHERE status='published'),
			(SELECT count(*) FROM submissions), (SELECT count(*) FROM submissions WHERE status='Accepted'),
			(SELECT count(DISTINCT user_id) FROM accepted_solutions),
			(SELECT count(*) FROM forum_posts)`).Scan(
			&c.users, &c.active, &c.packages, &c.publishedProblems,
			&c.submissions, &c.accepted, &c.usersWithSolutions, &c.forumPosts,
		); err != nil {
			t.Fatal(err)
		}
		return map[string]int64{
			"users": c.users, "activeUsers": c.active,
			"packages": c.packages, "publishedProblems": c.publishedProblems,
			"submissions": c.submissions, "acceptedSubmissions": c.accepted,
			"usersWithSolutions": c.usersWithSolutions, "forumPosts": c.forumPosts,
		}
	}
	before := snapshot()
	env, _ = adminCall(t, mux, adminToken, "GET", "/stats", "", 200)
	after := snapshot()
	stats := env["data"].(map[string]any)
	for key, hi := range after {
		got, ok := stats[key].(float64)
		if !ok {
			t.Fatalf("stats %s missing or not a number: %v", key, stats[key])
		}
		lo := before[key]
		if lo > hi {
			lo, hi = hi, lo
		}
		if int64(got) < lo || int64(got) > hi {
			t.Fatalf("stats %s: got %v outside concurrently observed range [%d,%d]", key, got, lo, hi)
		}
	}
	// AdminStats is additionalProperties:false in docs/openapi.yaml.
	allowed := map[string]bool{
		"users": true, "activeUsers": true, "packages": true, "publishedProblems": true,
		"submissions": true, "acceptedSubmissions": true, "usersWithSolutions": true,
		"forumPosts": true, "mostSolvedProblems": true,
	}
	if len(stats) != len(allowed) {
		t.Fatalf("stats has fields outside the OpenAPI schema: %v", stats)
	}
	for key := range stats {
		if !allowed[key] {
			t.Fatalf("unexpected stats field %q", key)
		}
	}
	if users, solutions := int64(stats["users"].(float64)), int64(stats["usersWithSolutions"].(float64)); users < 2 || solutions < 6 {
		t.Fatalf("stats is missing this test's fixtures: %v", stats)
	}
	if stats["activeUsers"].(float64) > stats["users"].(float64) ||
		stats["acceptedSubmissions"].(float64) > stats["submissions"].(float64) ||
		stats["usersWithSolutions"].(float64) > stats["users"].(float64) {
		t.Fatalf("stats counters are inconsistent: %v", stats)
	}

	solved := stats["mostSolvedProblems"].([]any)
	if len(solved) > 5 {
		t.Fatalf("mostSolvedProblems has %d rows, want at most 5", len(solved))
	}
	previous := -1.0
	foundLeader := false
	seen := make(map[string]bool, len(solved))
	for i, item := range solved {
		row := item.(map[string]any)
		id, _ := row["problemId"].(string)
		slug, _ := row["slug"].(string)
		title, _ := row["title"].(string)
		count, ok := row["solvedUserCount"].(float64)
		if !ok || len(id) != 36 || strings.Count(id, "-") != 4 || slug == "" || title == "" || count < 1 {
			t.Fatalf("mostSolvedProblems[%d] is malformed: %v", i, row)
		}
		if seen[id] {
			t.Fatalf("mostSolvedProblems repeats problem %s", id)
		}
		seen[id] = true
		if count < previous {
			t.Fatalf("mostSolvedProblems is not ordered by solver count: %v", solved)
		}
		previous = count
		if id == f.ProblemID {
			foundLeader = true
			if count < 6 {
				t.Fatalf("fixture problem solver count: got %v want at least 6", count)
			}
		}
	}
	if !foundLeader {
		t.Fatalf("fixture problem with six solvers missing from mostSolvedProblems: %v", solved)
	}

	studentToken = adminTestToken(t, targetID, targetEmail, "teacher")
	adminCall(t, mux, studentToken, "GET", "/users", "", 403)
	adminCall(t, mux, studentToken, "GET", "/users/"+targetID, "", 403)
	adminCall(t, mux, studentToken, "PATCH", "/users/"+targetID, `{"isActive":true}`, 403)
	adminCall(t, mux, studentToken, "GET", "/stats", "", 403)
	adminCall(t, mux, studentToken, "GET", "/audit-logs", "", 403)
}

func TestAdminLastActiveAdminProtected(t *testing.T) {
	f := testutil.Catalog(t)
	ctx := context.Background()

	var suffix, targetID string
	if err := f.Pool.QueryRow(ctx, `SELECT gen_random_uuid()::text`).Scan(&suffix); err != nil {
		t.Fatal(err)
	}
	if err := f.Pool.QueryRow(ctx, `INSERT INTO users(email,username,password_hash,role,is_active)
		VALUES($1,$2,'hash','admin',true) RETURNING id`,
		"last-admin-"+suffix+"@example.org", "last-admin-"+suffix).Scan(&targetID); err != nil {
		t.Fatal(err)
	}

	// Park any other active admins so the target is the only one.
	var parked []string
	rows, err := f.Pool.Query(ctx, `UPDATE users SET is_active=false WHERE role='admin' AND is_active AND id<>$1 RETURNING id::text`, targetID)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			t.Fatal(err)
		}
		parked = append(parked, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, id := range parked {
			if _, err := f.Pool.Exec(ctx, `UPDATE users SET is_active=true WHERE id=$1`, id); err != nil {
				t.Error(err)
			}
		}
		if _, err := f.Pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, targetID); err != nil {
			t.Error(err)
		}
	})

	repo := NewAdminRepository(f.Pool)
	role := "student"
	if _, err := repo.UpdateUser(ctx, targetID, UpdateUserRequest{Role: &role}); !errors.Is(err, ErrLastAdminProtected) {
		t.Fatalf("demote last active admin: got %v want ErrLastAdminProtected", err)
	}
	inactive := false
	if _, err := repo.UpdateUser(ctx, targetID, UpdateUserRequest{IsActive: &inactive}); !errors.Is(err, ErrLastAdminProtected) {
		t.Fatalf("deactivate last active admin: got %v want ErrLastAdminProtected", err)
	}
	name := "Still Admin"
	user, err := repo.UpdateUser(ctx, targetID, UpdateUserRequest{FullName: &name})
	if err != nil || user == nil || user.FullName == nil || *user.FullName != name {
		t.Fatalf("renaming last admin should be allowed: user=%v err=%v", user, err)
	}
}
