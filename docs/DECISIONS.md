# Decisions

## Product v2.0 decisions — 2026-11-05

- Preserve Run/Submit/Judge/Piston/Redis behavior. Schema, repository queries
  and DTOs may change to Package/Level/Problem, progress, leaderboard, Forum
  and AI review.
- Remove repository creation and all GitHub App/OAuth/webhook/commit flows from
  the target product. Existing code and schema are legacy migration inputs.
- Trigger AI only after an Accepted verdict has been persisted. AI score,
  summary and findings are separate from verdict and leaderboard; Run and other
  verdicts never trigger it.
- Forum is a separate service. Leaderboard ranks by solved problem/Accepted
  counts and never reads AI score.
- `/api/v1` Forum/Leaderboard/AI routes are target contracts, not live routes;
  current root routes remain documented separately.
- Execution window is 06/11/2026–22/11/2026; feature freeze 19/11 21:00.

The sections below are historical implementation decisions from the legacy
Repository/GitHub flow. They explain current code only and must not be used to
design the v2 product.

## Infrastructure Decisions (legacy/current baseline)

### Piston (Code Execution Engine)

- **Local development**: Piston runs as a Docker container (`innogen-piston`) alongside Postgres and Redis.
- **Access URL**: `http://localhost:2000` from host machine, `http://piston:2000` from other Docker containers on the same network.
- **Image**: `ghcr.io/engineer-man/piston:latest`
- **Privileged mode**: Required for Piston's internal container-based isolation.
- **Volume**: `pistondata:/piston` persists installed language packages across restarts.
- **Decision**: Piston is NOT built into the `innogen-backend` monorepo. It is a separate external service, consumed via REST API.

### X-User-ID Temporary Authentication

- **Decision**: For STEP 5 (submission_service) and development convenience, a temporary `X-User-ID` HTTP header is used for authentication.
- **Implementation**: `shared/middleware/auth.go` provides `XUserID()` middleware that reads the integer user ID from the header and injects it into the request context using the same `UserIDKey` used by JWT `Auth()`.
- **Production**: Replace with JWT Bearer token authentication via `middleware.Auth(jwtSecret)` before deploying.
- **Routes using X-User-ID**: POST /submit, GET /submissions/{id}, GET /me/submissions, GET /me/submissions/{problemId}/latest

### Redis Queue for Async Judging

- **Decision**: Use Redis list-based queue (LPUSH/BRPOP) for submission job distribution.
- **Queue name**: `submission_jobs`
- **Payload**: `{"submissionId":"<uuid>"}` (JSON)
- **Delivery semantics**: At-most-once. Jobs removed from Redis on dequeue before processing.
- **Crash recovery**: MVP uses a placeholder reconciliation sweep. Production should implement periodic re-enqueue of stale Pending/Running submissions.
- **Library**: `github.com/redis/go-redis/v9`

### Shared Packages Extraction

- **Decision**: Extracted `piston` client, `judge` logic, status `constants`, and `languageutil` into `shared/` to avoid code duplication between run_service and submission_service worker.
- **Affected packages**:
  - `shared/piston/` — Piston API v2 HTTP client (exported types: Request, Response, Stage, Client)
  - `shared/judge/` — Output evaluation with TLE detection via timeLimitMs
  - `shared/constants/` — Status string constants (Accepted, WrongAnswer, etc.)
  - `shared/languageutil/` — File name resolution from language config
- **Run service compatibility**: run_service re-exports status constants from shared/constants for backward compatibility.

### Mock Commit Flow (STEP 7)

- **Decision**: Use mock commit SHA (40-char random hex) generated via `crypto/rand` for MVP. No real GitHub API calls are made.
- **Integration**: `submission_service` worker imports `repo_service` directly (same Go module, monorepo pattern — no HTTP overhead).
- **Repository model**: One repository per `(user_id, subject_id)` pair, created/updated via `INSERT ... ON CONFLICT`.
- **Path format**: `<subjectSlug>/Session-<NN>/Lesson-<NN>/Problem-<NN>-<problemSlug>/<fileName>` with zero-padded order indexes.
- **Repo name format**: `<subjectSlug>-RinnoGen`
- **Transaction safety**: Commit DML (upsert repo, update submission, insert commit) wrapped in a single `pgx` transaction.
- **Idempotency**: `CommitSubmission` checks `commit_sha` before proceeding — safe to call multiple times.
- **Packages**: `repo_service/repository/` and `repo_service/service/` are outside `internal/` to allow cross-service imports by `submission_service`. Pathbuilder remains in `internal/` as it's only used by repo_service.

### Time Limit Detection

### Per-User GitHub App Installations (STEP 8)

- **Decision**: Each user installs the GitHub App into their personal account or an organization. The owner information is stored in `github_accounts.github_owner` and `github_accounts.github_owner_type`.
- **Owner selection**:
  - `User` → POST `/user/repos` to create repos
  - `Organization` → POST `/orgs/{owner}/repos` to create repos
- **No global org**: `GITHUB_ORG_NAME` is completely removed. All owner data comes from the database.
- **Installation token**: Generated per-request via GitHub App JWT (RS256, 10-min expiry) exchanged for installation access token.
- **Unchanged file behavior**: Before committing, the service fetches the existing file from GitHub. If content is identical, the commit is skipped.
- **Secret handling**: Private key is read from file path (`GITHUB_PRIVATE_KEY_PATH`), never stored in env vars. All secrets are excluded from `.gitignore`. No private key, JWT, or token is ever logged.
- **Interface**: `GitHubClient` interface enables mock injection for tests without real GitHub credentials.
- **Worker integration**: Submission worker calls repo_service via HTTP POST to `/internal/commits/accepted-submission` (decoupled, no direct package dependency).

### Time Limit Detection

- **Decision**: Piston `run_timeout` field is set from `problems.time_limit_ms`. If a process is killed by signal AND a time limit was configured, it's classified as Time Limit Exceeded instead of Runtime Error. Without a configured limit, signals are treated as Runtime Error (backward compatible with run_service).

### API Gateway as Single Entrypoint (STEP 9)

- **Decision**: The API Gateway (:8080) is the single public entrypoint. All browser/frontend requests go through it.
- **JWT validation at edge**: Gateway validates JWT Bearer tokens and injects `X-User-ID`, `X-User-Email`, `X-User-Role` headers before forwarding to backend services.
- **Backend auth simplification**: Backend services use `XUserID()` middleware (reads headers), not JWT validation. The gateway is the sole JWT validator.
- **Header stripping**: Incoming `X-User-*` and `Authorization` headers are stripped from public requests to prevent header injection.
- **Internal routes**: `/internal/*` routes (e.g., repo_service commit endpoint) are NOT registered in the gateway — accessible only via direct service-to-service communication.
- **Proxy timeout**: 30-second `ResponseHeaderTimeout` on all reverse proxies to prevent goroutine leaks.
- **Service URLs**: Configurable via `*_SERVICE_URL` env vars for flexible deployment (localhost, Docker, K8s).

### Dev Seed Data and E2E Validation (STEP 10)

- **Decision**: Use a standalone Go script (`scripts/hash_password.go`) with build tag `ignore` for bcrypt hash generation. The hash is baked into `seeds/dev_seed.sql` so no runtime Go dependency is needed for seeding.
- **Seed strategy**: All INSERTs use `ON CONFLICT DO NOTHING` so the seed is idempotent — safe to run multiple times.
- **Test cases**: One sample (visible to users) and one hidden (not exposed) to validate both paths.
- **E2E script**: Bash-based using curl + jq. Validates the full flow: login → curriculum → run → submit → poll → verify Accepted.
- **GitHub seed**: Commented out by default. Real GitHub App installation data must be filled in manually.

### Admin Content Management (STEP 11)

- **Decision**: Admin APIs live in `api_gateway/internal/admin/` rather than a separate service. The gateway already owns curriculum/problem data access.
- **Auth model**: JWT Bearer token with role=admin claim required. Non-admin tokens get 403.
- **Route isolation**: Admin routes use `http.StripPrefix("/admin", ...)` + private sub-mux to prevent collisions with public curriculum/problem routes.
- **Partial updates**: PUT endpoints use pointer types (`*string`, `*int`, `*bool`) to distinguish "not provided" from "set to zero/empty".
- **Test case visibility**: Admin endpoints return all test cases (sample + hidden). Public endpoints hardcode `visibility=sample`.
- **Tag immutability**: Tags support create and list only for MVP. Update/delete deferred.

### GitHub App Webhook Handling (STEP 12)

- **Decision**: Webhooks are handled by `repo_service` at `POST /webhooks/github` with HMAC-SHA256 signature verification. No JWT required.
- **Signature verification**: Uses `X-Hub-Signature-256` header, constant-time `hmac.Equal` comparison. Empty secret → all requests rejected.
- **Installation webhooks**: Stored in `github_installations` table before being linked to a local user. Allows webhook data to arrive before user connects their GitHub account.
- **Soft deletes**: Installations, accounts, and repositories use `status` columns (`active`, `suspended`, `deleted`, `archived`) — never hard-deleted. Preserves history.
- **Idempotency**: All writes use `INSERT ... ON CONFLICT DO UPDATE` or `UPDATE WHERE`. Re-delivered webhooks produce no side effects.

### GitHub App Connection State (Fix)

- **Decision**: GitHub connected state source of truth is `github_accounts` table in PostgreSQL. Frontend never relies on localStorage for connection status.
- **Callback flow**: GitHub redirects to `/github/callback?installation_id=XXX`. Frontend calls `POST /github/installations/link` to link the installation to the current user.
- **Race handling**: Webhook may populate `github_installations` before or after the link API is called. If the installation doesn't exist yet, the API returns 404 with a retry message.
- **Duplicate prevention**: `LinkGithubInstallation` checks if installation is already linked to another user before upserting.

### GitHub OAuth Account Linking (STEP 16)

- GitHub OAuth identifies the commit author. GitHub App installation provides write permission.
- OAuth account and installation owner are separate — for org installs, commit author is the OAuth-linked user.
- Noreply email: {github_user_id}+{username}@users.noreply.github.com
- OAuth access token is NOT stored after use.
- State uses signed JWT with 10-min expiry binding to the current RinnoGen user.

### API Response Contract
- All responses use envelope: `{status, code, message, data}` for success, `{status, code, message, error, details}` for errors
- `code` is the HTTP status code number
- `error` is the machine-readable code (BAD_REQUEST, UNAUTHORIZED, etc.)
- Business statuses (submission status, run result) stay inside `data`

### GitHub App Installation Lifecycle (Fix)
- `installation_id` is NEVER manually provided by users. It comes from GitHub callback or webhook.
- `github_installations` stores installation state independently from user link.
- `github_accounts` links RinnoGen user to OAuth identity + active installation.
- Uninstall is handled by `installation.deleted` webhook as soft deletion (status='deleted', uninstalled_at set).
- Historical submissions/commits are preserved after uninstall.
- `installed_at`/`uninstalled_at` track lifecycle timestamps.
