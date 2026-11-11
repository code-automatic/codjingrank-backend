# Backend AI Context

Updated 2026-11-11. See [PROJECT_STATE](../PROJECT_STATE.md) for tested revision
and current branch; [DECISIONS](DECISIONS.md) for product authority.

## Implemented repository facts

Go monorepo (`net/http`, pgx, shared packages), Gateway :8080, Auth :8081,
Run :8082, Submission API :8083 and worker. One shared PostgreSQL target schema
uses UUIDs and Package → three Levels → Problem. `schema.sql` is a destructive
fresh-database bootstrap, not an old-data migration.

Repository/GitHub and Subject/Session/Lesson runtime flows were removed. Sprint 2
integrates UUID Auth/JWT/refresh persistence, learner `/api/v1` catalog and admin
catalog CRUD/publication, and UUID Run/Submission/worker SQL. Shared execution
checks require an active user, published package/problem, unlocked level and
assigned active language. Submission detail is owner-only (outsiders404).
Retired root problem/admin paths return410; unmounted legacy source is not an
active API. Root Auth works with UUID tokens; full v1 Auth/profile/history/stats,
Forum, AI and leaderboard remain pending. Old root stats still needs schema work.
See [Sprint 2 acceptance](SPRINT2_ACCEPTANCE.md) for exact runtime scope.

Run is synchronous sample execution via Piston. Submit persists Pending, uses
Redis LPUSH/BRPOP, then worker/Piston/persisted verdict. No Asynq rewrite.
Shared success/error envelopes and routing 404/405 handling are implemented;
HEAD/204 are bodyless where endpoints select204 (root logout still uses200).
Real PostgreSQL/Redis tests passed. Task6 now also has a trusted real-Piston
Docker smoke; see [deployment evidence](be/DEPLOYMENT.md). It does not certify
production sandbox security. Worker recovery is still pending.

## Target contracts for frontend mocks

- [Full OpenAPI](openapi.yaml): Gateway learner/admin routes.
- [Learner OpenAPI](openapi_fe.yaml): matching projection without admin-only
  schemas or hidden runtime data.
- [Frontend guide](FRONTEND_API_CONTRACT.md): workflow, schema gaps, limits,
  auth, compatibility and feature coverage. Runtime status is recorded per
  operation and in the acceptance report; unimplemented operations remain targets.
- [Response contract](fe/RESPONSE_CONTRACT.md): the approved envelope;
  `error` is a string, `details` always present, numeric code matches HTTP.

Use camelCase, UUIDs and UTC. Do not copy old SRS snake_case/category examples
or pretend stripping a prefix makes old integer DTOs match new schema.
Root/v1 aliases must share a tested contract or an explicit adapter plan.

## Invariants and open integration work

AI only starts after persisted Accepted; AI failures/score never change verdict,
progress or leaderboard. Public AI API is GET only. Ranking uses distinct solved
problems, then distinct Accepted solutions, with RANK ties and UTC first-solve
periods. Level unlock is dynamic, checks all prior levels; empty levels do not
block. Published judge inputs stay frozen even after archive; clone for changes.

Hidden tests, wrappers, internal/provider diagnostics and secrets never reach
learners. Owner/admin checks protect source and AI; outsiders get404. Gateway is
the only browser entrypoint. Keep current HS256 pending any separate JWT
migration; the transport design alone does not implement RS256/JWKS.

`users.avatar_url` and `users.bio` exist in the fresh schema (nullable, 2048/1000
characters). Go DTO/query/handler integration is pending. Password recovery
still needs reset-token persistence and email delivery.
AI provider/rubric/quota/retry policy remains open; dimensions are dynamic in
mock DTOs. Optional features are documented, not silently marked complete.
Use a dedicated project test database; never HAPI or another project's services.
