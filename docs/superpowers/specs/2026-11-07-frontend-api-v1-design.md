# Frontend API v1 Contract Design

Status: documentation work requested by the owner; runtime implementation excluded.

## Goal and baseline

Provide a mockable contract for every F01–F13 learner/admin screen using the
approved envelope, one fresh shared UUID database, Package → three Levels →
Problem, synchronous sample Run and asynchronous Submit. AI starts only after
persisted Accepted and never changes verdict or ranking. Decisions and schema
outrank obsolete examples in the source SRS.

## Approach

Use `docs/openapi.yaml` as the complete public Gateway contract, including admin;
`docs/openapi_fe.yaml` is its learner-only projection for the existing docs UI.
Both are self-contained OpenAPI 3.0.3, with concrete payloads and examples.
Update the integration guide, SRS contract appendix and project context together.
Keeping the legacy OpenAPI would retain retired routes; introducing a third
canonical API file would leave competing sources. Neither is appropriate here.

## Contract choices

- CamelCase JSON, UUID identities, UTC timestamps, page/size/total pagination.
- Keep Run/Submit request `code`, Run response `status/passed/total/results`, and
  Submit response `data.submission`; new IDs are an explicit compatibility break.
- Owner/admin access for submissions and AI; outsiders receive 404. Hidden
  judge data and provider diagnostics never appear in learner schemas.
- Non-Accepted AI GET returns 409; Accepted without a review row returns 200
  with `review: null` so the post-commit scheduling gap is represented honestly.
- Catalog level state is dynamic. Published judge inputs are immutable even
  after archive; clone creates a draft. No rejudge or manual AI trigger.
- Complete optional SRS password recovery/profile/daily APIs as target contracts;
  flag reset-token persistence as a schema extension. Do not alter schema in
  this documentation task. Follow-up schema correction on 2026-11-07 adds
  nullable avatar_url/bio; Go persistence integration is still pending.
- Fill transport details (limits, nullability, empty states, error/status mapping)
  as documented contract design choices, not previously ratified product decisions.

## Verification

Validate OpenAPI, every reference and operation ID, examples against schemas,
learner projection boundaries, coverage F01–F13, security and bodyless 204s.
Document mock commands; do not claim endpoint execution or backend readiness.
