# Coding Judge Platform Database Design

**Status:** Approved; schema baseline implemented, service integration pending  
**Branch:** `feat/database-schema`

## Intent

Define the first PostgreSQL schema for the new standalone platform. The
database is empty and shared by all services. This is a clean target schema,
not a migration of the old `subjects → sessions → lessons → problems` model.

## Goals

- Keep the existing judge behavior: synchronous sample Run and asynchronous
  Submit through Redis, worker, Piston, and a final verdict.
- Model `Package → Level → Problem` directly.
- Keep authentication, judge data, forum, progress, leaderboard, and AI review
  in one database with internal foreign keys.
- Preserve every submission verdict while making Accepted progress idempotent.
- Prevent hidden test data and AI provider data from crossing public boundaries.

## Non-goals

- Repository, GitHub App, OAuth, webhook, commit, or repository metadata tables.
- Asynq, a second queue implementation, or a separate Auth database.
- Contest scoring, custom Run input, rejudge history, or AI-based ranking.

## Data model

### Identity and access

- `users`: UUID identity, email, username, password hash, role,
  active state, and timestamps.
- `refresh_tokens`: hashed rotating refresh tokens, expiry, revocation, device
  metadata, and self-reference to the token it replaced.

All service-owned user references use the same primary key and foreign keys.
Password hashes are stored; plaintext passwords and access tokens are never
stored.

### Catalog

- `packages`: topic/package name, slug, description, publication state, and
  ordering.
- `levels`: exactly three ordered levels per package (`easy`, `medium`,
  `hard`), with a unique `(package_id, level_no)` and `(package_id, slug)`.
- `problems`: one level owner, slug, statement, execution mode, limits,
  publication state, and timestamps.
- `tags` and `problem_tags`: optional problem filtering.
- `languages`, `problem_languages`, and `code_templates`: active runtimes and
  per-problem starter/wrapper code.
- `problem_samples`: public Run inputs and expected outputs.
- `problem_test_cases`: hidden Submit inputs and expected outputs; accessible
  only to judge/admin code paths.

Publication and immutable-test rules are enforced by service logic and checks
where PostgreSQL can express them. Samples and hidden tests are separate
tables so learner queries cannot accidentally return hidden fields.

### Judge and progress

- `submissions`: every attempt, beginning at `Pending`, then `Running`, and a
  final verdict. It stores source code, SHA-256 source hash, pass/total counts,
  runtime/memory, error, attempt count, and judge timestamps.
- `submission_test_results`: per-test result metadata. Hidden input and
  expected output are not returned by learner handlers.
- `accepted_solutions`: one row per distinct Accepted
  `(user_id, problem_id, language_id, source_hash)`. A submission references
  this row when it becomes Accepted. This separates “every attempt” from
  idempotent Accepted progress and avoids rejecting duplicate submissions.
- `user_problem_progress`: one row per user/problem, with first/last solve,
  Accepted solution count, and best runtime/memory.

Worker recovery uses submission status, `attempt_count`, `claimed_at`,
`last_error`, and timestamps. Reprocessing the same submission cannot create a
second `accepted_solutions` row or increment progress twice.

### Forum

- `forum_posts`: author, optional problem/language, post type, moderation
  status, title, Markdown content, optional code snippet, and counters.
- `forum_comments`: one-level replies through a nullable `parent_id`; the
  service validates that parent and post agree.
- `forum_post_likes`: composite primary key `(post_id, user_id)`.

Forum content is owned by users and soft-deleted/moderated through status.

### AI review

- `ai_reviews`: one review lifecycle per Accepted submission, with
  `queued/running/completed/failed`, score, summary, model, prompt version,
  provider request ID, error, and timestamps.
- `ai_review_findings`: severity, line range, title, explanation, suggestion,
  and stable ordering for one review.

The worker inserts or updates the review only after the Accepted verdict is
committed. AI fields are never part of verdict or leaderboard calculations.

## Constraints and derived data

- Primary keys use UUIDs for users and externally visible domain records; level
  numbers and display-order fields use integers where ordering is intrinsic.
- Foreign keys use `ON DELETE RESTRICT` for published catalog and submissions;
  forum content and review findings cascade only with their parent.
- Unique slugs and `(package_id, level_no)` prevent ambiguous catalog URLs.
- A partial/derived uniqueness rule is represented by `accepted_solutions`,
  rather than a partial unique constraint on mutable submission status.
- Leaderboard is a view/query over `user_problem_progress`, ranking distinct
  solved problems first and Accepted solution count second. AI score is absent.
- Indexes cover queue recovery (`status`, `claimed_at`), user history,
  problem history, public catalog slugs, and leaderboard ordering.

## Transaction boundaries

1. Submit API inserts `Pending` submission and returns its UUID; Redis enqueue
   happens after the insert.
2. Worker claims the submission and records `Running` with a claim timestamp.
3. Worker persists the final verdict and test summary.
4. For Accepted, one transaction inserts/gets `accepted_solutions`, upserts
   `user_problem_progress`, and commits before enqueueing AI review.
5. AI failure updates only `ai_reviews`; the submission remains Accepted.

## Security and retention

- Hidden test input/output, starter wrappers, service tokens, provider keys,
  and raw internal errors stay out of learner responses.
- Source code is retained with submissions for review; retention/deletion
  policy is a later product decision and must not weaken ownership checks.
- All writes use parameterized SQL. Timestamps are UTC-aware where supported.

## Approaches considered

1. **Legacy-compatible migration:** rejected because this project is standalone
   with an empty database and no requirement to preserve old hierarchy.
2. **Separate Auth and Judge databases:** rejected for now; it adds service
   synchronization and removes useful internal foreign keys without a current
   scaling or compliance need.
3. **One clean shared database:** selected because it matches the agreed
   architecture, simplifies transactions and operations, and keeps all domain
   invariants enforceable in one schema.

## Implementation decisions

- Use UUID primary keys for `users` and all domain records. JWT claims and Go
  DTOs use the same UUID representation.
- Keep `schema.sql` as the first executable deliverable; add a migration tool
  only when incremental deployment is required.
- Implement leaderboard scopes as `all`, `week`, and `month`, using UTC
  boundaries and first solves in the selected period. AI quota and cleanup
  schedules remain service configuration, outside this schema design.


## Verification and integration boundary

Run the PostgreSQL checks using:

```bash
SCHEMA_TEST_CONTAINER=<postgres-container> SCHEMA_TEST_USER=<admin-role> bash tests/schema/run.sh
```

The runner creates a fresh randomly named test database and removes that same
database on exit. `schema.sql` drops target objects and their data when run;
it is a reset/bootstrap script, not a non-destructive migration.

The schema does not yet make legacy Go consumers compatible with UUIDs or new
catalog tables. Do not apply historical migrations or seeds to this database.
Progress transactions, queue recovery, same-post one-level comment validation,
forum counters, access grants and learner DTO projection are service integration
work. The all-time leaderboard view exists; week/month/package queries still
need implementation. Sequential SQL checks do not prove concurrent correctness.
