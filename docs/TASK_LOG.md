## Task Log

### 2026-05-15 — Piston Infrastructure

- Added `piston` service to `docker-compose.yml` (ghcr.io/engineer-man/piston:latest, port 2000)
- Added `compose-up`, `compose-down`, `logs-piston` targets to `Makefile`
- Documented Piston access URLs in `docs/DECISIONS.md` and `docs/NEXT_STEPS.md`
- Validation: `docker compose up -d postgres redis piston` starts all three infrastructure services

### 2026-05-15 — Submission Service API (STEP 5)

- Implemented POST /submit (create Pending submission)
- Implemented GET /submissions/{id} (get submission by UUID with auth)
- Implemented GET /me/submissions (list user's submissions, no code in response)
- Implemented GET /me/submissions/{problemId}/latest (latest for user+problem)
- Added `Submission` model to shared/models
- Added `XUserID()` middleware for temporary header-based auth
- Added anti-spam detection (429 response on 10s cooldown trigger)
- Added UUID format validation on path parameters
- Added pagination limit (50) on list endpoint
- Files: dto.go, repository.go, service.go, handler.go, route.go, cmd/api/main.go
- Validated: go build, go vet, go test all pass
- Commit: pending

### 2026-05-15 — Redis Queue and Judge Worker (STEP 6)

- Extracted shared packages: `shared/piston/`, `shared/judge/`, `shared/constants/`, `shared/languageutil/`
- Migrated run_service to use shared packages
- Created Redis-backed job queue (`submission_service/internal/queue/`) using `go-redis/v9`
- Implemented async worker (`submission_service/internal/worker/`) with full judging flow:
  - Dequeue from Redis → load submission → mark Running → execute all test cases via Piston → judge results → update submission
  - Status priority: CompilationError > InternalError > TLE > RuntimeError > WrongAnswer > Accepted
  - Per-test TLE detection via problem's time_limit_ms + Piston run_timeout
- Updated POST /submit to enqueue submission for async judging
- Added `UpdateSubmissionStatus` and `UpdateSubmissionResult` repository methods
- Updated handler to use proper sentinel error for spam cooldown
- Deleted old run_service/internal/piston/ and run_service/internal/judge/ (moved to shared/)
- Validated: go build, go vet, go test (18/18 pass) all clean
- Commit: pending

### 2026-05-15 — Repo Service Mock Commit (STEP 7)

- Implemented `repo_service`: pathbuilder, repository, service, handler, route, dto
- Implemented `BuildRepoName`, `BuildFilePath`, `GenerateCommitSHA` in pathbuilder
- Implemented `CommitSubmission` flow: curriculum lookup → path building → repo upsert → submission update → commit insert
- Added idempotency guard: skips if commit_sha already set
- Wrapped commit DML in pgx transaction for atomicity
- Added ownership check on ListCommits (verifies user owns repository)
- Integrated submission worker: after Accepted, triggers `CommitSubmission` (best-effort)
- Moved `repo_service/repository` and `repo_service/service` out of `internal/` for cross-service imports
- Added `Repository` and `SubmissionCommit` models to shared/models
- Added 4 unit tests for pathbuilder (BuildRepoName, BuildFilePath, zero-padding, GenerateCommitSHA)
- Validated: go build, go vet, go test all pass
- Commit: pending

### 2026-05-15 — GitHub App Implementation (STEP 8)

- Implemented real GitHub App client with JWT generation (RS256) and REST API calls
- Added `GitHubClient` interface with `MockClient` for testing
- Implemented `GetInstallationToken`, `EnsureRepo`, `GetFileContent`, `CreateOrUpdateFile`
- Added per-user GitHub owner support: `github_accounts.github_owner`, `github_accounts.github_owner_type`
- Added `repositories.github_owner` column (denormalized)
- Removed `GITHUB_ORG_NAME`; owner comes from `github_accounts` table
- Added `GITHUB_DEFAULT_BRANCH` and `GITHUB_API_BASE_URL` config
- Added `GithubAccount` model to shared/models
- Updated `CommitSubmission` to use real GitHub flow: curriculum→language→account→token→repo→file→commit→DB
- Added `POST /internal/commits/accepted-submission` endpoint (internal)
- Worker now calls repo_service via HTTP instead of direct package import
- Unchanged file detection: skips commit if content matches existing file
- Content-diff optimization: only pushes to GitHub if code changed
- Schema migration: `migrations/001_add_github_owner.sql` for existing DBs
- Security: `.gitignore` covers `*.pem`, `*.key`, `secrets/`; secrets never logged
- Validated: go build, go vet, go test (22/22 pass) all clean
- Commit: pending

### 2026-05-15 — API Gateway Proxy (STEP 9)

- Implemented reverse proxy for all backend services using `net/http/httputil`
- Added JWT validation at the gateway edge for all protected routes
- Gateway injects `X-User-ID`, `X-User-Email`, `X-User-Role` from JWT claims into backend requests
- Strips incoming `X-User-*` and `Authorization` headers from public requests (security)
- Kept existing curriculum/problem direct-DB routes working
- Public routes: curriculum, /auth/login (no auth required)
- Protected routes: /auth/me, /run, /submit, /submissions/*, /me/*, /repositories/*
- Internal routes NOT exposed through gateway
- Added proxy Transport timeout (30s ResponseHeaderTimeout)
- auth_service switched from JWT auth to X-User-ID (gateway validates JWT)
- run_service added XUserID middleware
- Added `*_SERVICE_URL` config variables for flexible deployments
- Worker now uses `REPO_SERVICE_URL` config (fixes Docker/K8s deployment)
- Validated: go build, go vet, go test (22/22 pass)
- Commit: pending

### 2026-05-15 — Local Seed Data and E2E Validation (STEP 10)

- Created `seeds/dev_seed.sql` with admin user, JavaScript language, subject, session, lesson, problem, test cases
- Created `scripts/hash_password.go` for bcrypt hash generation
- Created `scripts/e2e_mvp.sh` — full end-to-end validation: health → login → subjects → run → submit → poll → verify
- Added `make seed-dev` (apply schema + seed) and `make e2e` (run E2E script)
- Updated `README.md` with quick start, architecture overview, and API table
- Updated project memory docs
- Validated: go build, go vet, go test (22/22 pass)
- Commit: pending

### 2026-05-17 — Admin Content Management APIs (STEP 11)

- Implemented 30 admin CRUD endpoints under `/admin/*` prefix in api_gateway
- Created `AdminRepository` with 31 methods for all curriculum/problem entities
- Added `AdminAuth` middleware (JWT validation + admin role check)
- Resources: languages, subjects, sessions, lessons, problems, lesson-problem links, test cases, tags, problem-tag links
- Protected by JWT Bearer token + role=admin check (401/403)
- Admin routes isolated via StripPrefix + sub-mux, no collision with public routes
- Pagination support on problem list
- PostgreSQL error mapping: 23505→409, 23503→404, 23514→400, 23502→400
- Enum validation for difficulty (Easy/Medium/Hard) and visibility (sample/hidden)
- Files: middleware.go, dto.go, repository.go, handler.go, route.go + main.go wiring
- Validated: go build, go vet, go test (22/22 pass)
- Commit: pending

### 2026-05-17 — GitHub App Webhook Handling (STEP 12)

- Implemented `POST /webhooks/github` endpoint in repo_service (no JWT, HMAC-SHA256 verification)
- Created `repo_service/internal/webhook/` package: verifier.go, dto.go, handler.go, service.go
- Added `github_installations` table for webhook data (schema.sql + migration)
- Added `status` columns to `github_accounts` and `repositories` tables
- Handles events: installation (created/deleted/suspend/unsuspend), installation_repositories (added/removed), repository (renamed/deleted/archived)
- Webhook signature verification using X-Hub-Signature-256 with constant-time HMAC comparison
- Body size limit (1MB), 413 on overflow
- Schema migration: `migrations/002_step12_github_webhooks.sql`
- Added `GITHUB_WEBHOOK_SECRET` config
- Added `GithubInstallation` model to shared/models
- Validated: go build, go vet, go test (22/22 pass)
- Commit: pending

### 2026-05-17 — Local Test UI (STEP 13)

- Created `tools/test_ui/` — Vite + React + TypeScript test UI
- Features: Dashboard (health checks), Auth (login/logout), Curriculum browser, Runner, Submissions (submit + poll), Admin CRUD, GitHub webhook notes
- API client with base URL config (localStorage) and Bearer token auto-attach
- Created `scripts/send_github_webhook_test.sh` for HMAC-signed webhook testing
- Added `make test-ui-install`, `make test-ui`, `make test-ui-build` targets
- Validated: npm install + npm run build pass
- Commit: pending

### 2026-05-19 — GitHub Connect UI Fix
- Added GET /github/connection and POST /github/installations/link endpoints in repo_service
- Added gateway proxy routes for both
- Created GithubConnect.tsx page with connection check, callback handling, and re-linking
- Fixed missing InstallationID in connection response
- Added duplicate installation check before linking
- Validated: go build, go test, npm build all pass

### 2026-05-20 — GitHub Commit E2E Visibility (STEP 14)

- Added `commitUrl` to submission read responses (LEFT JOIN submission_commits)
- Updated Test UI Submissions page with structured commit metadata (repoPath, commitSha, clickable commitUrl)
- Updated GitHub Connect page with Accepted-commit flow explanation
- Created `scripts/e2e_github.sh` for E2E GitHub commit validation
- Added `make e2e-github` target
- Validated: go build, go test, npm build all pass

### 2026-05-20 — GitHub OAuth Account Linking (STEP 16)
- Added OAuth flow: start URL → GitHub authorize → callback → exchange code → get user → store identity
- Added 4 new columns to github_accounts: github_noreply_email, commit_author_name, oauth_connected_at, oauth_status
- Created repo_service/internal/oauth package: state, client, service, handler, dto
- Added GET /github/oauth/start-url, GET /github/oauth/callback, GET /github/account, POST /github/oauth/disconnect
- Updated test UI with OAuth Account Identity section (username, avatar, email, commit author)
- OAuth callback redirects to frontend with ?oauth=connected or ?oauth=error

### Stabilize GitHub Connection Flow
- Fixed LinkGithubInstallation to extract owner from GitHub API GET /app/installations/{id}
- Added GetInstallation to GitHubClient interface
- Merged UI useEffect hooks for reliable OAuth callback handling
