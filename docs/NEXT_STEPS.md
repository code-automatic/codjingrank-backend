# Next Steps — v2.0

Updated: 2026-11-14. Delivery deadline: 22/11/2026.
Baseline: [decisions](DECISIONS.md); current evidence:
[Sprint 3 BE acceptance](SPRINT3_ACCEPTANCE.md).

## Complete Sprint 3 integration

- Publish the scoped BE task16/18/19/20/21 commits, run required PR checks and
  merge with a merge commit preserving every original SHA.
- FE17/22 and QA23 remain separately assigned. The backend report does not
  certify frontend integration or the team's QA acceptance.
- FE should use the learner OpenAPI implementation flags and runtime matrix;
  retain mocks for target-only operations.

## Next backend work

1. Complete the full target v1 Auth/profile contract. Root UUID Auth/refresh
   already works; do not assume every target operation has a v1 alias.
2. Implement leaderboard endpoints from distinct solved problems/Accepted
   sources, excluding AI score.
3. Finalize provider/rubric/quota/retry choices before implementing isolated
   AI review after persisted Accepted. AI failure must never change verdict.
4. Implement Forum/moderation and remaining target routes with contract tests.

Accepted progress/unlock, worker recovery, ownership and v1 submission
history/latest/stats are implemented; see the acceptance report for tests,
compatibility changes and recovery guarantees.

## Release gates

- Exact per-problem CPU/output policy and adversarial/load sandbox validation
  remain open. Piston's global output bound is not per-problem enforcement.
- Backend CI and trusted real-Piston smoke already ran remotely for task6;
  rerun on Sprint 3's merged revision and publish the actual run link.
- Complete frontend E2E and team QA; trusted backend probes alone do not certify
  hostile multi-tenant production isolation.
- Reserve 20–21/11 for regression/security/demo checks; 22/11 for submission.

Use only dedicated project test infrastructure. `schema.sql` is destructive
bootstrap for fresh disposable databases, never an existing user's database.
