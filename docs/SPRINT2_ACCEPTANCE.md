# Sprint 2 Backend Acceptance

Updated: 2026-11-11. Implementation tested at `2a6d440`, based on main
`f10c4b1`. Changes are local and stacked across scoped branches; no push, PR,
merge or deployment was performed for Sprint 2.

## Task results

| Task | Delivered | Evidence |
|---|---|---|
| 7 | Atomic Package + exactly three Levels; dynamic progress/unlock reads | Real PostgreSQL transaction and all-prior-level tests |
| 8 | UUID catalog repositories, safe learner DTOs | Real SQL; samples separated from hidden tests; no wrapper exposure |
| 9 | Gateway catalog endpoints | Auth, active account, pagination/filter validation and HTTP tests |
| 10 | Admin catalog CRUD, publication, archive and clone | DB role enforcement, publication prerequisites and frozen judge-input tests |
| 13 | Publication/language/unlock checks before Run/Submit | Shared access checks, UUID execution consumers; Piston HTTP double |
| 15 | Catalog/auth/envelope/ownership contract checks | Real DB/Redis tests; 45 actual successful HTTP responses validated against OpenAPI |

BE owner: Trung Hiếu. FE tasks11 (Hoàng Huy),12 (Quốc Thịnh),14 (Thanh Tùng)
remain pending in the separate unavailable frontend repository. These results
complete the scoped BE work, not every target endpoint or production gate.

## Runtime availability for frontend

| Gateway routes | Current state |
|---|---|
| GET `/api/v1/languages`, `/tags` | Implemented; optional supplied token must be valid |
| GET `/api/v1/packages`, `/packages/{slug}`, `/packages/{slug}/levels` | Implemented; active authenticated user |
| GET `/api/v1/problems`, `/problems/{slug}` and `/languages`, `/template`, `/samples` children | Implemented; locked content denied; learner-safe projection |
| `/api/v1/admin/packages`, `/levels`, `/problems`, `/languages`, `/tags` families | Implemented methods follow OpenAPI; includes lifecycle, clone, associations, templates and tests |
| POST `/api/v1/run`, `/api/v1/submit` | UUID integrated, access-checked; partial target contract, Piston double only |
| Root `/auth/register`, `/auth/login`, `/auth/refresh`, `/auth/logout`, `/auth/me` | UUID persistence/JWT/refresh integrated; not the full v1 Auth DTO/policy |
| Root `/submissions/{id}` | Owner-only detail; outsider404; target v1 detail/history pending |
| Root `/problems` and `/admin` families | Retired with410 to prevent legacy bypass |
| Other target v1 operations | Pending; continue using mocks |

Paths abbreviated in a row retain its stated prefix. Full target schemas remain
in [OpenAPI](openapi.yaml) and the exact [learner projection](openapi_fe.yaml).
`x-implementation-status` distinguishes DB-tested catalog, partially integrated
Run/Submit and unimplemented targets. It is not a deployment indicator or proof
that every error branch was schema-validated. The 45 captures cover successful
payloads; separate Go assertions exercise negative/access/lifecycle cases.
Published GitHub Pages remains at the previously published revision until these
local changes are reviewed and published.

## Preserved behavior and fixes

- Run executes samples synchronously and creates no submission. Submit persists
  Pending → Redis LPUSH/BRPOP → worker → Piston → verdict. No Asynq rewrite.
- UUID Auth/JWT/refresh and execution SQL were necessary catalog prerequisites.
  Integer JWTs are rejected; refresh rotation is transactional and replay-tested.
- Worker claims Pending atomically; duplicate final jobs do not execute again.
- Strict JSON rejects client execution overrides. Provider/hidden-test diagnostics
  are generic; function-mode compiler output does not expose wrapper code.
- Timeout and memory values are forwarded to Piston; Run aggregation preserves
  Time Limit Exceeded. The regression failed before the fix and passed after it.
- Published judge configuration stays frozen after archive; clone creates a draft.
  Level unlock reads all preceding levels; empty levels do not block.

## Reproducible verification

Use a dedicated disposable PostgreSQL16 database bootstrapped from `schema.sql`
and Redis7. Never bootstrap another project's database. Local tests used only
`codjingrank-sprint2-postgres` (127.0.0.1:55432) and
`codjingrank-sprint2-redis` (127.0.0.1:56379); HAPI was not touched.

```bash
# Set CATALOG_TEST_DATABASE_URL and CATALOG_TEST_REDIS_ADDR for owned test services.
gofmt -s -w .
go test -count=1 ./...
go vet ./...
go build ./...
SCHEMA_TEST_CONTAINER=codjingrank-sprint2-postgres SCHEMA_TEST_USER=innogen \
  bash tests/schema/run.sh
# Use a fresh capture file: capture appends.
CATALOG_CONTRACT_CAPTURE=/tmp/new-sprint2-responses.jsonl \
  go test -count=1 ./api_gateway/internal/admin ./api_gateway/internal/catalog
python3 tests/sprint2_contracts.py /tmp/new-sprint2-responses.jsonl
```

The DB/Redis environment must also be exported for the Go test commands;
without it integration tests skip. The Python validator requires PyYAML and
jsonschema as test tooling. Full DB/Redis tests, vet/build, bootstrap twice plus
SQL constraints and 45 captured payload validations passed at the tested revision. Both OpenAPI
files also passed Redocly1.34.5 minimal lint (localhost server warnings); the
learner operation projection remains exact.

## Commits and review

Original commits retained in order: `93adfce`, `ee9e5ad`, `416481c`, `7a89927`,
`de32a15`, `7877093`, `5e4a236`, `6fbad06`, `5f34b38`, `35f7906`, `b6698ed`,
`2a6d440`. Each is a scoped implementation/test/fix step. Current documentation
branch includes all of them; main is unchanged. Preserve their SHAs during review
and eventual merge commits.

GitNexus CLI was refreshed; MCP results missed a source-confirmed Run caller, so
source references, diff review and tests remain necessary. Two independent
reviewer sessions disconnected without a completed review; author self-review
found and fixed the recorded issues, but independent approval is still pending.

## Remaining limits

- Accepted solution/progress writes belong to the next sprint. Tests seed progress
  rows; they do not establish worker-driven unlock end to end.
- Worker crash recovery is absent; Running jobs can remain stuck after a crash.
- Full target v1 Auth/profile/password policy and submission history/stats remain
  incomplete. Legacy root stats SQL is still incompatible with the new hierarchy.
- Run provider failures currently produce an Internal Error result rather than
  the complete target502/504 transport contract. Do not mark Run fully compliant.
- No live Piston, full-stack or frontend E2E. Actual CPU/memory/network/output
  sandbox enforcement is unverified; request-forwarding tests cannot prove it.
- Forum, leaderboard and AI review are pending. AI must run only after persisted
  Accepted, independently of verdict/ranking. No AI replacement hook was added.
