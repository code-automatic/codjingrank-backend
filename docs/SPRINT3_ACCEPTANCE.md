# Sprint 3 Backend Acceptance

Updated 14/11/2026. Scope: workbook BE tasks **16, 18, 19, 20, 21**, assigned to
Trung Hiếu. FE17/22 belong to the separate frontend repository; QA23 remains
the assigned QA task. This report does not mark those tasks complete.

## Evidence by task

| Task | Delivered behavior | Implementation evidence |
|---|---|---|
| 16 — synchronous sample Run | Samples only; no submission/progress write; 102400-byte source validation; existing Gateway rate limiter; safe502/504 provider errors; Piston runtime milliseconds fixed | [Run](https://github.com/code-automatic/codjingrank-backend/commit/19e3db0), [provider response safety](https://github.com/code-automatic/codjingrank-backend/commit/8b59e45), `run_service/internal/{dto,service,handler}/*test.go` |
| 18 — ownership and polling | Owner/active DB-admin detail; outsider404; inactive403; Pending/Running/final polling; latest absence200/null; root/v1 submission reads | [ownership](https://github.com/code-automatic/codjingrank-backend/commit/4685349), [read contracts](https://github.com/code-automatic/codjingrank-backend/commit/5188080), `submission_service/internal/handler/{sprint3,polling,history_contract}_test.go` |
| 19 — worker recovery | Lost enqueue, stale Running, duplicate delivery, attempt fencing, live lease and retry exhaustion; original Redis list queue retained | [recovery](https://github.com/code-automatic/codjingrank-backend/commit/04cb6d8), repository/worker `recovery_test.go`, real-Piston smoke |
| 20 — Accepted progress/unlock | Verdict, distinct Accepted source and progress in one transaction; deduplicated source count, concurrency, rollback, best runtime and dynamic next-level unlock | [progress](https://github.com/code-automatic/codjingrank-backend/commit/1dc9078), `submission_service/internal/repository/progress_test.go` |
| 21 — history/stats | Stable pagination and AND filters; no list source; latest language filter; target-schema statistics, UTC activity/streak and progress-backed nullable rank without rating | [history/stats](https://github.com/code-automatic/codjingrank-backend/commit/5188080), repository `stats_test.go`, handler contract tests |

Integration review preserves original per-execution deadlines rather than adding
a new90s cap for an entire batch. Long valid test sets retain their existing
judge semantics. Piston HTTP calls keep the30s timeout and caller cancellation.

## Frontend handoff

Canonical Gateway routes (all require Bearer):

- `POST /api/v1/run`, `POST /api/v1/submit`.
- `GET /api/v1/submissions/{id}`: owner or current active DB admin; source here only.
- `GET /api/v1/me/submissions?page=1&size=20&problemId=...&languageId=...&status=...`.
- `GET /api/v1/me/submissions/{problemId}/latest?languageId=...`.
- `GET /api/v1/me/stats`.

History returns `data.items/page/size/total`, sorted by `createdAt DESC,id DESC`;
size is1–100. Filters combine with AND. Empty history is `items:[],total:0`;
no latest attempt is `data.submission:null`. Only detail/latest include source.
Malformed query values fail400. Stats has no `rank.rating`; `currentRank` is null
without qualifying progress. Activity counts all Accepted submissions by UTC
judging date, while progress/ranking counts distinct Accepted sources per
`(user,problem,language,source_hash)` and distinct solved problems.

Equivalent root aliases return the same target payloads. This intentionally
replaces the old root history shape, latest-empty404 and rating formula. Clients
must adopt the documented target fields; no parallel legacy payload is served.
UUID identities and the approved response envelope remain unchanged.

Contracts: [frontend guide](FRONTEND_API_CONTRACT.md), [full OpenAPI](openapi.yaml),
[learner OpenAPI](openapi_fe.yaml),
[published API docs](https://code-automatic.github.io/codjingrank-backend/).
Only implemented operations are marked integration-tested; other operations
remain targets and must stay mocked.

## Recovery and data guarantees

The worker reconciles at startup and every30s, scanning up to100 rows per batch.
Pending older than30s is re-enqueued. Running with a heartbeat older than2m is
reset/re-enqueued; live jobs refresh every15s. Three claimed attempts exhaust to
a safe Internal Error. Redis enqueue failure leaves a durable Pending row for
another scan. Recovery commits before enqueue, so a second enqueue failure or
crash can be retried later.

Only the current Running attempt can finalize. Duplicate delivery cannot claim
a Running/final submission; an old worker cannot overwrite a replacement
attempt. Recovery is at-least-once, not exactly-once remote execution after a
crash. The database fences duplicate effects and preserves terminal verdicts.

Accepted source/progress writes are atomic with verdict persistence. Repeated
identical source does not inflate Accepted counts. Non-Accepted outcomes retain
their submission history but do not write solved progress. Existing dynamic
unlock still requires every published problem in all earlier levels; empty
levels do not block. No AI or repository/GitHub call was added.

## Verification

Local evidence uses the dedicated project PostgreSQL16/Redis7 containers, plus
a separate UUID-named disposable Compose stack for the trusted Piston smoke:

- `gofmt -s -w .`, `make gofmt-check`, full DB/Redis `go test -count=1 ./...`.
- `go vet ./...`, `go build ./...`, race tests over changed Go packages and
  separate Redis recovery integration.
- Schema bootstrap twice and SQL constraint checks against a temporary database.
- Both OpenAPI minimal lints;19 captured successful HTTP responses validated
  against both full and learner contracts. Error paths have separate Go tests.
- `make docker-smoke`: all services ready; UUID Auth/refresh/catalog; sample-only
  Run; Pending observed in a real Redis list; intentional Redis job deletion;
  periodic recovery to real-Piston Accepted; root/v1 equivalence; history,
  latest/stats, progress, anti-spam429, outsider404 and Wrong Answer.
- Trusted direct Piston timeout/memory/network/output probes; the owned test
  containers and volumes are removed afterward. HAPI infrastructure is untouched.

CI now runs recovery separately from parallel queue tests and validates captured
HTTP payloads against both contracts. PR/check links are supplied at handoff.

## Review and remaining scope

Author review checked ownership/header spoofing, current admin/active state,
SQL filters/pagination, stale attempt races, terminal verdict protection,
same-source counting, transaction rollback, source/hidden-test projection and
root/v1 compatibility. A separate reviewer was not available in this session;
this is author review with automated and real-runtime evidence.

This is backend acceptance, not frontend E2E or production deployment. Full v1
Auth/profile, Forum, AI and leaderboard endpoints remain later work. AI provider,
rubric/quota/retry decisions are still open; AI will only run after persisted
Accepted and will not affect verdict or rank.

The existing Gateway limiter is per-IP/per-process (default120 requests/60s).
Piston response reads are capped at4MiB; this is not enforcement of each
problem's `output_limit_bytes`. Exact per-problem output/CPU policy and hostile
multi-tenant/load certification remain release gates. Memory metrics may be
null. The trusted sandbox probes do not certify production isolation.
