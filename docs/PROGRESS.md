# Implementation Progress

Updated: 2026-11-14. Sprint 2 and Sprint 1 task6 merged through PR7/8 at
`0a7afa5`. Sprint 3 backend is implemented on `test/sprint3-acceptance`;
publication status and actual checks belong in its acceptance report.

## Sprint 2 backend

| Task | Scope | Status |
|---|---|---|
| 7 | Atomic package/three levels; dynamic progress/unlock reads | Merged, DB-tested |
| 8 | UUID catalog repository and safe learner DTOs | Merged, DB-tested |
| 9 | Gateway learner catalog API | Merged, DB-tested |
| 10 | Admin catalog CRUD, publication, archive and clone | Merged, DB-tested |
| 13 | Shared publication/language/unlock checks before Run/Submit | Merged, integration-tested |
| 15 | Catalog/auth/envelope/ownership contract tests | Merged; 45 real HTTP payloads validated |

Evidence: [Sprint 2 acceptance](SPRINT2_ACCEPTANCE.md).

## Sprint 3 backend

| Task | Scope | Status |
|---|---|---|
| 16 | Synchronous sample Run; size/timeout/provider safety | Implemented |
| 18 | Active owner/admin detail and polling; root/v1 routes | Implemented |
| 19 | Redis-list recovery, leases, retries and attempt fencing | Implemented |
| 20 | Transactional Accepted source/progress and dynamic unlock | Implemented |
| 21 | Filtered history, latest and UTC statistics | Implemented |

Evidence, per-task commits, compatibility changes and exact verification limits:
[Sprint 3 acceptance](SPRINT3_ACCEPTANCE.md). Run stays synchronous; Submit
keeps Pending → Redis LPUSH/BRPOP → worker → Piston → persisted verdict.
No Asynq, repository/GitHub or AI flow was added.

## Sprint 1 task6 — Docker/CI baseline

Merged in PR8. Dedicated Compose bootstrap, port isolation, pinned Piston and
Go/schema/OpenAPI/Docker CI are integrated. All three remote baseline jobs,
including trusted real-Piston runtime smoke, passed on `0a7afa5`:
[Actions run](https://github.com/code-automatic/codjingrank-backend/actions/runs/42952389102).
See [deployment guide](be/DEPLOYMENT.md) for commands and limits.
Sprint 3 fixes runtime milliseconds and extends smoke to lost-job recovery,
progress and submission read contracts.

## Remaining scope

- FE11/12/14 and FE17/22 belong to the separate frontend repository; QA23 is
  the assigned team QA task. Backend acceptance does not mark them complete.
- Full target v1 Auth/profile, leaderboard endpoints, Forum and isolated AI review.
- Exact per-problem CPU/output policy, adversarial/load certification and FE E2E.
- Author review and automated/runtime checks are recorded transparently;
  they do not substitute for an independent review approval.

## Removed

Repository/GitHub runtime integration, old curriculum routes, migrations/seeds
and legacy test UI were removed. This product uses a fresh standalone database;
there is no legacy-data migration requirement. Unmounted legacy source is not
an implemented public endpoint.
