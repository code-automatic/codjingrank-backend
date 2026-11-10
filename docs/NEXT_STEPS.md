# Next Steps — v2.0

Updated: 2026-11-11. Delivery deadline: 22/11/2026.
Baseline: [decisions](DECISIONS.md); current evidence:
[Sprint 2 acceptance](SPRINT2_ACCEPTANCE.md).

## Finish Sprint 2 integration

- Review the local stacked feature commits before opening PRs. Preserve original
  commits with merge commits; do not mark local changes deployed.
- FE tasks11/12/14 remain in the separate frontend repository. Use the learner
  OpenAPI and runtime matrix; keep unimplemented endpoints on mocks.
- Root Auth issues UUID tokens, but the full target v1 Auth/profile contract is
  pending. Implement it explicitly rather than adding an untested prefix alias.

## Next backend work

1. Persist Accepted solutions/progress transactionally and idempotently, then
   prove level unlock from actual worker results. Current progress tests seed
   progress rows; they do not prove this missing write path.
2. Add worker crash recovery/retry strategy; duplicate final jobs are guarded,
   but Running jobs can remain stuck after a crash.
3. Complete v1 submission detail/history/stats and owner/admin read policy.
   The root stats query still needs UUID-schema integration.
4. Implement leaderboard from distinct solved problems/Accepted solutions;
   isolate AI review after persisted Accepted. AI errors never change verdict.
5. Implement Forum/moderation and remaining target routes with contract tests.

## Release gates

- Validate real Piston timeout, CPU/memory, network and output restrictions;
  passing HTTP-double tests only proves request forwarding and judge handling.
- Resolve Docker/CI audit findings, run full-stack E2E and frontend integration.
- Finish provider/rubric/quota/retry choices before AI implementation.
- Reserve 20–21/11 for regression/security/demo checks; 22/11 for submission.

Use only dedicated project test infrastructure. `schema.sql` is destructive
bootstrap for fresh disposable databases, never an existing user's database.
