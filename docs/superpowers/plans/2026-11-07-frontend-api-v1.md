# Frontend API v1 Documentation Plan

Execute inline as a single agent, per owner instructions. This task implements
the requested documentation; it does not implement Go handlers or schema changes.

**Goal:** Make every F01–F13 learner/admin API mockable.
**Spec:** [design](../specs/2026-11-07-frontend-api-v1-design.md).
**Architecture:** One full public OpenAPI and a learner projection, with shared
schema names and matching examples; Markdown provides workflow and migration notes.
**Tools:** Existing Python YAML, JSON Schema and OpenAPI validators; no app dependency.

## Constraints and review focus

Preserve judge flow/envelope, UUID schema, Accepted-only AI and non-AI ranking.
Check hidden-test/provider leaks, missing review rows, dynamic level locks,
publication immutability, optional schema extensions and old/new ID compatibility.
Preserve unrelated AGENTS/workbook edits. No source or schema modifications.

## Tasks

- [x] Define concrete request/response schemas, common errors, pagination and examples.
- [x] Cover all learner and admin features, marking every operation target-only.
- [x] Update frontend guide/auth/SRS/context; remove obsolete API examples and routes.
- [x] Validate specifications, examples, projection, coverage and required security.
- [x] Self-review, record checks and integration limits in PROJECT_STATE.

No new automated behavior tests are needed for this documentation-only change;
specification and example validation are the acceptance checks. Keep document
commits focused; no publishing/merge is part of this new task without a request.

Validation: OpenAPI and all examples pass; Prism exercised94 full/42 learner
operations and11 named states. Go tests pass. Final self-review reconciled
SRS/current schema and corrected mock auth/state examples. No source/schema
changes; live integration not tested. Single-agent workflow per owner.
