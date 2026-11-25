# Sprint 4 backend design

Status: approach 1 implemented locally. Baseline: main `bcbe9a1`.

## Goal and scope

Deliver workbook BE tasks 24, 25, 26, 27 and 30, plus the backend portion of
task 32. The source is the student-ID workbook, Task rows 25–33. FE tasks
28/29/31 and frontend acceptance remain in the separate frontend repository.
Reuse `schema.sql`, `docs/DECISIONS.md`, `docs/openapi.yaml` and its learner
projection. Existing target HTTP payloads and access rules are authoritative.

## Approaches

1. Recommended proposal: configure an AI HTTP provider adapter; verify with a
   deterministic local HTTP test double before live credentials/model selection.
   This delivers integration evidence without claiming live AI quality.
2. Integrate a selected live provider/model immediately. This additionally needs
   user selection, credentials, spending limits and a live verification window.

The local implementation uses approach 1. Live provider/model selection is
pending.
Use separate Go Forum and AI services, the shared PostgreSQL database, existing
Gateway/proxy patterns, pgx, stdlib HTTP and the approved response envelope.
Keep the Submission Redis list queue and all judge/progress semantics intact.

## AI: tasks 24–26

- Reuse `ai_reviews` and `ai_review_findings`; add only columns/constraints
  required for bounded attempts and worker fencing. One review per submission
  is protected by the existing unique constraint. Persist model/prompt version
  internally; never expose them or provider diagnostics to learners.
- The internal enqueue contract accepts only a submission UUID, authenticates
  the service caller, re-reads persisted status, and rejects non-Accepted work.
  It never accepts user-supplied source, verdict, score or owner claims.
- Notify AI only after the judge transaction commits Accepted. Notification
  failure must not roll back or replace verdict/progress. AI periodically
  reconciles persisted Accepted submissions missing reviews so a crash or lost
  notification does not lose work. Notification is bounded and must not block
  further judging for the provider call duration.
- AI workers claim durable queued rows atomically. Retry/recovery is bounded,
  and completion is fenced against stale attempts. Remote execution can repeat
  after an ambiguous timeout; only the current attempt may persist a result.
- Proposed initial policy, subject to approval: one concurrent provider request,
  30-second request timeout, at most three attempts; retry transient network,
  rate-limit and server failures only. Validate response sizes, JSON, score
  0–100, finite dimension values, findings and line ranges before storing.
  Terminal failure stores a generic public message separately from verdict.
- Proposed rubric v1: correctness of explanation, readability, maintainability,
  and time/space efficiency; total score 0–100. This is advisory feedback, never
  a replacement judge. Exact prompt weights are builder-owned, versioned and
  documented; provider quality is not certified by a test double.
- Review only submission source, language and public problem statement; exclude
  hidden tests, driver/wrapper code, user identity and credentials. Treat source
  and statement as untrusted prompt data. No provider tools/code execution.
- GET `/api/v1/submissions/{id}/ai-review` follows OpenAPI: active owner/admin,
  outsiders404, non-Accepted409, Accepted without a row200 `review:null`.
  queued/running/completed/failed fields follow the existing nullability contract.
  No public enqueue/retry/rejudge endpoint. No runtime calls without explicit
  provider configuration. Test doubles are test-only and never fabricate a
  completed review in deployed code.
- Quota proposal: bounded concurrency/attempts and no public retry; a live
  deployment's monetary quota must be configured at the provider before live
  acceptance. Do not backfill historical submissions into paid AI implicitly:
  use an explicit enablement cutoff for reconciliation and notifications.

## Leaderboard: task 27

- Gateway owns GET leaderboard and `/me` using existing progress/accepted counts.
  Rank by solved problems then distinct Accepted solutions; SQL RANK ties have
  gaps, stable output ordering uses rank/user UUID. Never query AI scores.
- Support all/week/month and package slug. UTC week begins Monday; period
  membership uses firstSolvedAt with a half-open interval. Accepted count is
  the current distinct-solution total for qualifying problems.
- Guests: all scope, first page, at most50. Validate supplied tokens even on
  optional-auth routes. Require active authenticated users for other pages,
  period scopes and `/me`; return null entry if no qualifying progress.
- Read username/avatar directly from the shared users table. Omit email and
  inactive accounts. Reuse existing package visibility rules.
- Initial implementation can query PostgreSQL directly: contract permits a
  cache lag of up to60s but does not require stale results. Introduce cache
  only if measured performance needs it; preserve guest/auth response isolation.

## Forum: task 30

- A separate `forum_service` owns posts, comments, likes and moderation; Gateway
  exposes the documented learner/admin routes. Use the existing schema.
- Public reads include published posts only; invalid supplied JWT is rejected.
  Writes require an active user. Owner checks and active database admin checks
  are performed server-side; forged headers never grant privileges.
- Match existing pagination, search/filter/sort and DTO fields. Read solution
  posts publicly; creating or converting a post to solution requires persisted
  Accepted progress for its selected problem. Validate related problem/language
  visibility and membership as required by the contract.
- Parent comments must belong to the same post and be top-level; reject deeper
  nesting422. Deleted comments are tombstones with no retained content exposed.
  Soft-delete posts; hidden/deleted posts are excluded from learner reads.
- Like/unlike is idempotent; transactional counters and row locks prevent
  negative/drifting counts under concurrent operations.
- Preserve Markdown as Markdown while escaping raw HTML and rejecting unsafe
  link/image schemes at the server boundary. Do not implement sanitization
  using regex. If the existing dependencies cannot safely parse Markdown, use
  a narrowly scoped maintained parser/sanitizer with malicious-input tests.
  Frontend rendering still needs its own sanitizer and E2E verification.

## Integration, ownership and acceptance

The primary owns contracts, schema integration and PROJECT_STATE.md. Planned
specialist boundaries after approval: coder-be AI; coder-be Forum; primary
leaderboard and cross-service wiring. Assign shared config/schema ownership
explicitly before parallel changes. Reviewer and tester remain read-only;
devops handles Docker/CI wiring; docs updates only assigned documentation.

Acceptance requires real disposable PostgreSQL/Redis integration, AI HTTP
double tests for success/timeout/bad response/retry/stale completion, Accepted
eligibility and lost-notification recovery, concurrency and ownership tests,
UTC/tie/pagination leaderboard cases, Forum nesting/likes/moderation/XSS cases,
and actual HTTP response validation against both OpenAPI projections. Run
affected race tests, full Go tests, vet/build, schema checks and service startup
smoke. Update implementation flags only for verified operations. Keep existing
Run/Submit tests and sandbox probes green.

Use scoped branches and focused commits, preserve unrelated AGENTS/workbook
edits. Publishing/merging and live-provider acceptance need their applicable
authorization; this design does not silently authorize deployment or spending.

## Exclusions and remaining decisions

No frontend implementation, full v1 Auth/profile expansion, password recovery,
new judge algorithm, GitHub integration, AI-based ranking, extra Forum features
or production sandbox certification. Preserve existing results/history.

See [Sprint 4 acceptance](../../SPRINT4_ACCEPTANCE.md) for the implemented
adapter, local evidence and remaining live-provider verification.
