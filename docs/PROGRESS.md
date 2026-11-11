# Implementation Progress

Updated: 2026-11-11. Local implementation is based on main `f10c4b1`;
not pushed, merged or deployed. Evidence: [Sprint 2 acceptance](SPRINT2_ACCEPTANCE.md).

## Sprint 2 backend

| Task | Scope | Status |
|---|---|---|
| 7 | Atomic package/three levels; dynamic progress/unlock reads | Implemented and DB-tested |
| 8 | UUID catalog repository and safe learner DTOs | Implemented and DB-tested |
| 9 | Gateway learner catalog API | Implemented and DB-tested |
| 10 | Admin catalog CRUD, publication, archive and clone | Implemented and DB-tested |
| 13 | Shared publication/language/unlock checks before Run/Submit | Implemented and integration-tested with Piston double |
| 15 | Catalog/auth/envelope/ownership contract tests | Implemented; 45 real HTTP payloads validated against OpenAPI |

UUID Auth/refresh persistence and Run/Submission/worker consumers were integrated
as prerequisites. Run remains synchronous; Submit keeps Pending → Redis list →
worker → Piston → verdict. No Asynq or AI rewrite.

## Pending

- FE tasks11/12/14: package/level filtering, problem/starter UI and editor/drafts;
  separate frontend repository is unavailable here.
- Accepted progress/solution writes, crash recovery, full v1 Auth/profile/history
  and stats, leaderboard, Forum and isolated post-Accepted AI review.
- Adversarial/load sandbox validation and frontend E2E. Task6 trusted real-Piston
  backend smoke is now available; see evidence below.
- Independent review did not complete because reviewer sessions disconnected;
  self-review and regression checks are not an external approval.

## Removed

Repository/GitHub runtime integration, old curriculum routes, migrations/seeds
and legacy test UI were removed. There is no legacy-data migration requirement:
this product uses a fresh standalone database. Retained unmounted legacy source
must not be mistaken for live endpoints.

## Sprint 1 task6 — Docker/CI baseline

Delivery configuration now bootstraps only empty PostgreSQL volumes, isolates
Compose names/ports, pins Piston and fails CI visibly on lint errors. The first
real full-stack smoke passed Accepted/Wrong Answer and ownership plus trusted
Piston timeout/memory/network/output probes. Final reproducibility verification
passed with automatic owned-resource cleanup; [deployment guide](be/DEPLOYMENT.md) lists commands and limits.
GitHub CI is configured locally, not remotely executed. Runtime unit conversion
(ms treated as seconds), per-problem output limits and production hardening
remain separate follow-up work.
