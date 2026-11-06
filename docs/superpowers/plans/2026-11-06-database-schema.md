# Database Schema Implementation Plan

> **For agentic workers:** Execute this plan task-by-task in the current worktree. Keep each task in a focused commit.

**Goal:** Replace the legacy bootstrap schema with a clean shared PostgreSQL schema for the standalone Coding Judge Platform.

**Architecture:** One PostgreSQL database owns identity, Package/Level/Problem catalog, judging, progress, forum, leaderboard queries, and AI review. The schema preserves Run/Submit behavior at the data boundary but removes the old curriculum and repository/GitHub tables.

**Tech Stack:** PostgreSQL 16, `pgcrypto`, Go `pgx` consumers, Redis list queue, Piston.

**Spec:** `docs/superpowers/specs/2026-11-06-database-schema-design.md`

## Global Constraints

- The database is new and empty; no legacy data migration or compatibility view is required.
- Run remains synchronous and does not persist a Run record.
- Submit inserts `Pending`, then Redis/worker/Piston updates the same submission to a final verdict.
- Every verdict is retained; only Accepted creates idempotent progress and AI review eligibility.
- AI score is never used by leaderboard queries.
- Repository, GitHub, Asynq, and separate Auth DB tables are out of scope.
- Hidden test input/output and AI provider credentials must stay out of learner-facing data paths.
- Use UUID primary keys and UTC-aware timestamps.

## Review Focus

- A worker retry must not increment progress or Accepted counts twice.
- A duplicate Accepted source must remain a valid submission without violating uniqueness.
- A stale Running claim must be recoverable using persisted claim metadata.
- Published problems must not allow hidden test mutation through ordinary admin writes.
- Public queries must have no path to return hidden test data or AI provider errors.
- Deleting a user, problem, or submission must respect retention and foreign-key policy.

---

### Task 1: Replace bootstrap and identity schema

**Files:**
- Modify: `schema.sql`

**Interfaces:**
- Produces `users`, `refresh_tokens`, enums, timestamp helper, and the clean reset/bootstrap boundary consumed by all later tables.

- [ ] Add `pgcrypto`, target enum types, `set_updated_at`, and destructive local reset for only target objects.
- [ ] Create UUID `users` with unique lower-case email/username indexes, password hash, role, active state, and timestamps.
- [ ] Create rotating `refresh_tokens` with token hash uniqueness, expiry, revocation, and replacement self-reference.
- [ ] Validate the section with `psql` against a temporary PostgreSQL database.
- [ ] Commit as `feat: add standalone identity schema`.

### Task 2: Add Package/Level/Problem catalog

**Files:**
- Modify: `schema.sql`

**Interfaces:**
- Consumes identity/bootstrap types from Task 1 and produces catalog tables referenced by judging and forum tables.

- [ ] Create packages, levels, problems, tags, problem_tags, languages, problem_languages, and code_templates.
- [ ] Enforce exactly three level numbers per package with checks and unique package ordering.
- [ ] Create separate public `problem_samples` and hidden `problem_test_cases` tables.
- [ ] Add publication/status checks, immutable published test protection, foreign keys, and catalog indexes.
- [ ] Validate catalog inserts and forbidden level/order values with SQL assertions.
- [ ] Commit as `feat: add package level problem catalog`.

### Task 3: Add submissions and idempotent progress

**Files:**
- Modify: `schema.sql`

**Interfaces:**
- Consumes catalog tables and produces the submission UUID/status fields expected by Submit API and worker recovery.

- [ ] Create submission status enum and `submissions` with source hash, verdict metrics, queue claim metadata, and foreign keys.
- [ ] Add status transition and anti-spam constraints/triggers without changing the synchronous Run behavior.
- [ ] Create `submission_test_results`, `accepted_solutions`, and `user_problem_progress`.
- [ ] Add an all-time leaderboard view plus indexes for all/week/month query predicates.
- [ ] Validate duplicate Accepted handling and worker retry idempotency with SQL assertions.
- [ ] Commit as `feat: add judge persistence and progress schema`.

### Task 4: Add forum, AI review, and optional daily challenge

**Files:**
- Modify: `schema.sql`

**Interfaces:**
- Consumes users, problems, languages, submissions, and accepted solutions; produces persistence for Forum and post-Accepted AI review.

- [ ] Create forum post/comment/like enums and tables with ownership, one-level replies, moderation status, and counters.
- [ ] Create `ai_reviews` and `ai_review_findings` with one review per Accepted submission and independent failure status.
- [ ] Create `daily_challenges` as an optional catalog feature without coupling it to judge state.
- [ ] Add indexes and cascade/restrict policies matching the approved design.
- [ ] Validate forum like uniqueness, comment parent consistency at service boundary, and AI review uniqueness.
- [ ] Commit as `feat: add forum and ai review persistence`.

### Task 5: Update state and run final verification

**Files:**
- Modify: `PROJECT_STATE.md`
- Verify: `schema.sql`, `docs/superpowers/specs/2026-11-06-database-schema-design.md`

- [ ] Run PostgreSQL parse/bootstrap validation against a temporary database.
- [ ] Run `git diff --check`, `go test ./...`, and inspect `git diff --stat`.
- [ ] Confirm no repository/GitHub table or old Subject/Session/Lesson table remains in `schema.sql`.
- [ ] Record branch, commits, checks, and known source-code follow-up work in `PROJECT_STATE.md`.
- [ ] Commit as `docs: record standalone schema verification`.


## Execution handoff

- Baseline committed as `0c78986`; publication fixes as `5c01950`; repeatable
  PostgreSQL test runner as `8567ef0`. The initial four schema sections were
  committed together rather than the planned per-domain commits; history has
  not been rewritten.
- PostgreSQL bootstrap and focused constraints pass via `tests/schema/run.sh`.
- `go test ./...` and `git diff --check` pass; no Go sources were modified.
- Service-level duplicate-progress handling, recovery, comment validation and
  period leaderboard queries are not implemented by this schema-only task.
- Historical docs/API edits and outputs remain outside these schema commits.
