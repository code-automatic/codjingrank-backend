# Project State

Updated: 2026-11-07
Branch: `fix/api-envelope`, based on `main` at `772398c`; runtime cleanup
`65630ab`, legacy artifacts `9d5ff9a`; schema baseline at `ff17e0b`.

## Objective and approved baseline

Prepare the standalone coding platform with its own fresh database. Keep one
shared PostgreSQL database, Package → Level → Problem, synchronous sample Run,
and Submit Pending → Redis LPUSH/BRPOP → worker → Piston → persisted verdict.
AI will run only after persisted Accepted, independently of verdict and ranking.
The baseline is [decisions](docs/DECISIONS.md),
[AI context](docs/AI_CONTEXT.md) and the
[approved schema design](docs/superpowers/specs/2026-11-06-database-schema-design.md).

## Current task: API envelope implementation

- Branch `fix/api-envelope` from `main` `772398c`; runtime commits `cf7a7a8`
  (shared router/bodyless responses) and `ee315aa` (HTTP service integration).
- Approved DECISIONS shape is unchanged. Generated 404/405 now use the JSON
  envelope, preserve Allow and route parameters, leave redirects/HTML intact.
  HEAD and 204 are bodyless. Admin authorization remains before dispatch.
- Boundary tests cover auth, rate limiting, panic and proxy transport failure;
  Wrong Answer stays a successful response with business status inside data.
- `go test -count=1 ./...`, `go vet ./...`, `go build ./...`, affected packages
  with `go test -race`, formatting and `git diff --check`: passed on runtime tree
  committed as `ee315aa`. No database/Redis/Piston integration run.
- Shared and integration staged GitNexus detection reports CRITICAL due to shared
  response callers/service entrypoints. Reviewed affected scope; judge unchanged.
- Independent reviewer attempted twice but disconnected. Primary code review
  completed; external review still needed before merging this branch.
- All commits in `feat/database-schema` and `chore/remove-legacy-code` already
  belong to main. `docs/api-envelope` has no unique commit; implementation and
  scoped envelope documentation are consolidated on `fix/api-envelope`.
- Merge approved PRs with a merge commit, preserving every original SHA; no
  squash/rebase. GitHub already allows merge commits. No push/PR/merge performed
  for the current branch; no repo-wide merge settings changed.
- Preserve unrelated docs edits and `outputs/`. Live UUID/schema integration,
  `/api/v1` routing and full backend readiness remain separate tasks.

## Completed

- Schema baseline committed in `0c78986`, guards in `5c01950`, repeatable SQL
  checks in `8567ef0`; 22 tables, 36 foreign keys. `ff17e0b` records its limits.
- Removed the repository/GitHub vertical: repo service, Auth OAuth/App methods,
  Gateway proxies, Accepted commit hook, config, submission metadata and SQL joins.
- Removed Subject/Session/Lesson handlers, DTOs, repositories and admin routes;
  retained language/problem/test/tag functionality for schema integration.
- Removed all seven legacy migrations, old seeds and legacy smoke scripts.
  New deployment starts from `schema.sql`; there is no old-data migration need.
- Removed the old Vite test UI including tracked dependencies/build output,
  its Make/CI targets and the old domain-specific Nginx configuration.
- Updated Compose/Make entrypoints and current README/deployment guidance.
  Historical audits/contracts are explicitly marked superseded for removed code.
- Added a Gateway contract test for removed routes and retained auth protection.

## Verification on cleanup working tree

- `go test -count=1 ./...`, `go vet ./...`, `go build ./...`: passed.
- Gateway removal test was also run against the original route file: failed for
  all nine tested retired GET/POST routes, then passed with the cleaned file.
- `gofmt`, `git diff --check`, both Compose `config --quiet`: passed.
- GitNexus impact checked 126 distinct function names. `Load` is CRITICAL
  (six service callers); only retired config fields were removed and callers
  updated. Ambiguous names/file lookups were cross-checked in source/build.
- GitNexus whole-diff detection reports CRITICAL due to deliberate large legacy
  deletion (including tracked UI dependencies). This is not a runtime safety proof.
- Run/judge/Piston/queue source and `schema.sql` were not modified in cleanup.
- No containers started, no database touched, no remote repositories changed.
  Earlier schema validation used unrelated infrastructure; do not repeat it.

## Integration limits and next work

- Retained Go IDs and SQL names are still incompatible with the target UUID
  schema. Passing Go tests proves compilation/unit behavior, not DB/API operation.
- Gateway `/api/v1`, new catalog APIs, Forum, AI service and target leaderboard
  are not implemented. No fake AI call replaces the removed commit hook.
- Adapt Auth/JWT/middleware and persistence to UUIDs/new schema, then catalog,
  Run/Submit consumers; add DB/worker integration and recovery tests.
- Keep current verdict mechanics while integrating accepted solutions/progress.
- Use a dedicated project PostgreSQL 16 test container, never HAPI or another
  project's service. `schema.sql` destructively bootstraps only a fresh database.
- Full Docker image build, live Piston/Redis judging and DB integration were not
  run in cleanup. SQL checks were not rerun because schema did not change.

## Existing unrelated work

Documentation changes in AGENTS, decisions/context/contracts/OpenAPI/SRS and
`outputs/` predated this task. Preserve them; do not bundle them into cleanup
commits. Cleanup status notes added to these documents remain reviewable alongside
those edits. No push or merge is authorized by this cleanup task.
