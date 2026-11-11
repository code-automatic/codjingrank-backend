# Docker/CI Runtime Baseline Implementation Plan

**Goal:** Complete Sprint 1 task6 with reproducible startup and real Run/Submit
through PostgreSQL, Redis, worker and Piston, retaining Redis LPUSH/BRPOP.
**Spec:** Task6 in the assigned workbook; docs/DECISIONS.md and constitution.
**Architecture:** Keep the existing Compose services and Go entrypoints. Use a
separate Compose project with disposable volumes; bootstrap only its empty DB.
No application feature or queue rewrite. Execute inline with executing-plans.

## Scope and review focus

- Never touch HAPI, existing databases, unrelated AGENTS/workbook edits or remote
  Git. Test setup must fail closed when a named project already exists.
- Keep credentials/source out of Docker build context. Production exposes only
  Gateway on loopback; Piston requires a dedicated trusted execution host.
- Fresh schema bootstrap must precede consumers. Existing volumes are never
  destructively bootstrapped on restart. Piston runtime install must fail visibly.
- CI must fail on test/lint/build errors; validate both OpenAPI projections.
- A healthy HTTP response alone is insufficient: assert sample Run and persisted
  asynchronous Accepted plus a negative verdict using a real runtime.

## Ordered tasks

1. [x] Correct Compose startup, isolation and build inputs. Reproduce prod port
   merging and missing schema readiness; fix only delivery configuration. Check
   merged JSON exposes no internal ports; build all five retained Go services.
2. [x] Add a bounded disposable smoke script/fixture covering health, UUID auth,
   refresh, safe catalog, sample Run, Submit/poll/worker and ownership against
   actual Piston. Start with one installed Python runtime; report enforced limits
   separately from readiness. Clean up only resources owned by this run.
3. [x] Make ci-openapi propagate failures, add Go/schema/Redis/OpenAPI/Docker CI.
   Prove forced npx failure returns nonzero; execute the local equivalents.
4. [x] Record task6 acceptance, limitations and commands in deployment guidance
   and PROJECT_STATE.md. Review diff, run GitNexus staged detection and commit
   focused steps. No unsolicited push/PR/merge/deployment.

Ruling: Include delivery fixes exposed by the task6 audit; do not expand into
full auth/profile, worker crash recovery, AI or production sandbox certification.

## Verification result

Delivery implementation `9da0ab8`: all five images built and full trusted
real-Piston smoke passed twice after fixture/readiness fixes. Final script passed
Accepted/Wrong Answer, Pending queue observation, ownership and resource probes;
its disposable stack/volumes were removed. Go DB/Redis suite, vet/build,
repeatable schema constraints, both OpenAPI lints and actionlint1.7.7 passed.
Forced lint/runtime-install failures propagate nonzero. GitNexus staged detection
reported LOW, no application symbols changed. Reviewer session disconnected;
final review is author self-review, no independent approval. See deployment guide.

Ruling: Keep existing runtime-ms conversion and per-problem output policy as
explicit separate application fixes; task6 verifies infrastructure and trusted
execution, not production correctness/certification of every judge metric.
