# RinnoGen / Innogen Backend AI Context

## Current repository

This is a Go monorepo for a LeetCode-like programming practice platform. The
current production-shaped catalog is:

`Subject -> Session -> Lesson -> Problem`

The planned product catalog is `Topic/Package -> Level (Easy/Medium/Hard) ->
Problem`; that is a migration target, not the current schema.

## Services and runtime

- `api_gateway` (`:8080`): public entry point, curriculum/problem handlers,
  reverse proxies, CORS, request ID, rate limit.
- `auth_service` (`:8081`): bcrypt users, HS256 access JWT, rotated HttpOnly
  refresh cookie, GitHub App connection.
- `run_service` (`:8082`): synchronous sample execution through Piston.
- `submission_service` (`:8083`) + worker: creates Pending submissions, Redis
  list queue, hidden-test judging and verdict updates. The old best-effort repo
  commit still exists in code; its planned replacement is an automatic internal
  AI review call only after persisting an Accepted verdict.
- `repo_service` (`:8084`): legacy repositories, GitHub App/OAuth, commits and
  webhook; this flow is scheduled for removal.
- Shared code is under `shared/`; PostgreSQL schema source is `schema.sql`.

Redis and GitHub integration already exist. Do not describe them as future
work. The current queue is Redis `LPUSH/BRPOP`, not Asynq.

## Current API facts

- Gateway paths have no `/api/v1` prefix: `/subjects`, `/problems/{slug}`,
  `/run`, `/submit`, `/me/stats` and `/me/submissions`.
- Responses use `{status, code, message, data}` or
  `{status, code, message, error, details}` from `shared/response`.
- IDs are integer for users/problems/languages; submission IDs are UUIDs.
- A Run response is synchronous. Submit inserts `Pending` before judging and
  persists every final verdict, including failures.
- There is no leaderboard list API, AI code review, forum, or level-unlock
  implementation yet. `/me/stats` computes only the current user's summary/rank.
- Target behavior keeps judge verdicts and execution mechanics, changes schema
  for Package/Level/progress/forum/AI review, removes repository/GitHub logic,
  and keeps AI score separate from leaderboard ranking.
- Confirmed: Run and non-Accepted submissions never trigger AI review. AI
  failure does not change the judge verdict or ranking. Repository/GitHub flow
  removal is approved as target scope, but implementation has not been authorized.
- Documentation plan window: 2026-11-06 through 2026-11-22. This is a target
  schedule, not evidence that implementation has started or will fit available
  staffing.
- Planned new boundaries are `forum_service` for posts/comments/likes and
  `ai_service` for code-quality review. Neither service exists yet; do not claim
  their routes are live.

## Working rules

- Use `net/http`, pgx, and existing shared packages; do not invent a new service
  or silently replace the architecture.
- Keep logs/errors in English, Go names camelCase, and run `gofmt` plus
  `go test ./...` after code changes.
- Treat `schema.sql`, Go routes/DTOs, and `docs/FRONTEND_API_CONTRACT.md` as
  current contract. Treat the SRS in `code-automatic/srs` as proposed target
  behavior; check `docs/BACKEND_REALITY_AUDIT.md` before implementing it.
