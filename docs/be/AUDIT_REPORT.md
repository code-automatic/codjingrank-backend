# Audit Report

**Date:** 2026-11-05
**Source:** GitNexus knowledge graph (2,099 symbols, 6,525 relationships, 177 flows) + manual code review

> This audit describes the current legacy backend. The v2 target preserves the
> judge mechanics, removes Repository/GitHub flow, and adds Forum, Leaderboard
> and post-Accepted AI review; see `docs/BACKEND_REALITY_AUDIT.md`.

## HIGH Priority Issues

### 1. CI Go version mismatch
- **File:** `.github/workflows/ci.yml:17` says `go-version: '1.23'`
- **File:** `go.mod:3` says `go 1.26.2`
- **File:** `Dockerfile:2` uses `golang:1.26-alpine`
- **Impact:** CI runs on Go 1.23, may pass when code requires 1.26 features. Docker builds on 1.26.
- **Fix:** Update ci.yml to `go-version: '1.23'` (Docker Hub max). Or use `1.26` if available.

### 2. Legacy Repository/GitHub surface
- **Files:** `repo_service/`, GitHub routes and commit metadata
- **Fact:** These still exist in the current code but are outside the v2 target.
- **Fix:** Remove them during the approved schema/API migration after deciding how to retain historical data.

### 3. Makefile PHONY line too long
- **File:** `Makefile:3` — single .PHONY line with 40+ targets, 300+ characters
- **Fix:** Split into multiple .PHONY lines grouped by category.

## MEDIUM Priority Issues

### 4. OpenAPI spec — missing `/admin/*` routes in full spec
- `docs/openapi.yaml` documents only public/frontend routes. Admin CRUD routes (`/admin/languages`, `/admin/subjects`, etc.) are NOT documented.
- The full spec should include all 30 admin endpoints.

### 5. OpenAPI spec — `/auth/register` documented in FE spec but route registers via gateway as public
- `docs/openapi_fe.yaml` includes POST /auth/register ✅
- `docs/openapi.yaml` (full) also includes it ✅
- `docs/be/DEPLOYMENT.md` doesn't mention registration

### 6. Docker compose — no healthcheck on Go services
- **File:** `docker-compose.yml` — go services have `depends_on` but no `healthcheck` blocks
- Alpine runtime images lack `curl`/`wget` for HTTP healthchecks
- **Impact:** `depends_on` without `condition: service_healthy` means services start before dependencies are ready
- **Fix:** Add `condition: service_healthy` to Go service depends_on, or accept race conditions locally.

### 7. Secret-scan in CI only checks git-ls-files
- **File:** `.github/workflows/ci.yml:68` — only checks tracked files
- **Impact:** Does NOT scan commit content for accidentally staged secrets
- **Fix:** Add `grep -rE 'BEGIN.*PRIVATE KEY|ghp_|ghs_' --exclude-dir={.git,node_modules,vendor,dist}` or use trufflehog/git-secrets.

## LOW Priority Issues

### 8. Deprecated functions still used
- **File:** `shared/response/response.go` — `JSON()` is marked deprecated but still used by internal services
- Callers in run_service, submission_service, repo_service, auth_service still call `response.JSON()`
- No user-facing impact since JSON wraps to Success

### 9. README API docs link to `/docs/fe` — 404 if gateway not running
- `/docs/fe` served from api_gateway; cannot open if gateway is down

### 10. E2E script uses python3 for JSON parsing
- `scripts/e2e_mvp.sh` and `scripts/e2e_github.sh` depend on `python3`
- If `jq` is preferred but not enforced, script fails silently

### 11. No ARCHITECTURE.md link from README.md
- README has Docker + API Docs sections but no link to `docs/be/ARCHITECTURE.md`

## Verified Correct

| Check | Status |
|-------|--------|
| Response envelope: status/code/message/data | ✅ Consistent |
| Business status inside data (submission.status, run.status) | ✅ Correct |
| Legacy GitHub OAuth routes vs implementation | ✅ Matches current code; removed from v2 target |
| Legacy webhook route (no JWT) vs implementation | ✅ Matches current code; removed from v2 target |
| Docker service names match Makefile targets | ✅ Consistent |
| CI all 5 jobs defined | ✅ |
| .dockerignore covers secrets | ✅ |
| PROGRESS steps 1-16 match TASK_LOG | ✅ (except stale 17) |
| Gateway middleware stack (7 layers) | ✅ Correct |
| Admin routes isolated via StripPrefix | ✅ Correct |
