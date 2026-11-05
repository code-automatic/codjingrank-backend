<!--
Sync Impact Report
- Version change: scaffold/unratified -> 1.0.0
- Modified principles: placeholder principles -> nine project principles covering judge behavior, execution security, boundaries, data integrity, AI isolation, testing, simplicity, observability, and migration discipline
- Added sections: Project Context, Development Workflow, Governance
- Removed sections: unresolved template placeholders and example-only guidance
- Follow-up TODOs: none; ratification date is the first adoption date of this constitution
-->

# Codjingrank Constitution

## Core Principles

### I. Preserve Existing Judge Behavior

The current judge flow is a compatibility contract. Run MUST execute sample tests
synchronously. Submit MUST preserve `Pending -> Redis -> worker -> Piston ->
verdict`. A schema, DTO, or service change MUST NOT silently change these
semantics. The project MUST keep the Redis list queue and existing worker model
unless a separately approved migration replaces them; target documents mentioning
Asynq do not authorize an Asynq rewrite.

### II. Security-First Code Execution

User code MUST run only through the isolated Piston execution path and MUST NOT
execute directly on the host or receive a Docker socket. Every execution MUST
enforce a time limit and appropriate CPU and memory limits, with an explicit
network policy. Hidden tests, credentials, provider keys, tokens, source that is
not needed by the caller, and internal service data MUST never be returned to a
learner or logged.

### III. Service and Contract Boundaries

The API Gateway is the sole public entrypoint for browser clients. Services MUST
publish explicit request, response, event, and error contracts using the shared
response envelope. Internal calls MUST use service authentication and stable
ownership boundaries. Any breaking route, DTO, event, or schema change MUST ship
with a migration or compatibility plan and updated contract tests.

### IV. Data Integrity and Recovery

`schema.sql` is the source of truth for the schema currently running. Database
changes MUST preserve required user, problem, and submission history or document
an explicit retention decision. Submission and queue operations MUST be
idempotent where retried, or MUST have a tested recovery strategy for crashes,
duplicate delivery, and stale jobs. Verdict persistence and related progress
updates MUST remain transactionally consistent.

### V. AI Review Isolation

The worker MAY request AI review only after an `Accepted` verdict has been
persisted. AI score, summary, and findings MUST be stored separately from the
judge verdict and MUST NOT affect leaderboard calculations. Run and every
non-Accepted verdict MUST skip AI review. AI timeout, provider failure, or
malformed output MUST remain an AI failure state and MUST never turn an Accepted
submission into a failed judge result.

### VI. Testable Changes

Contract changes MUST include contract tests. Worker, queue, or retry changes
MUST include integration or recovery coverage. Every bug fix MUST include a
regression test when the behavior is testable. Contributors MUST run `gofmt -s
-w .` and `go test ./...`; failures MUST be fixed or explicitly reported before
integration. Tests MUST cover authorization and data exposure at trust
boundaries, not only the happy path.

### VII. Simplicity and Reuse

Contributors MUST reuse existing Go packages, service boundaries, validators,
response helpers, and queue patterns before adding new abstractions. A new
dependency, wrapper, or service MUST have a concrete requirement that existing
code cannot satisfy. The smallest change that completes the behavior is the
default; a service MUST NOT be rewritten when a local behavior change is enough.

### VIII. Observable and Safe Operations

Services MUST emit structured logs through the existing logging conventions.
Logs MUST NOT contain secrets, tokens, or sensitive source code. Every request
and asynchronous job MUST be traceable by request ID, job ID, or submission ID,
with those identifiers carried across queue and service boundaries. Errors MUST
retain an actionable machine-readable code without exposing internal details.

### IX. Migration Discipline

Documentation MUST distinguish behavior that exists in the current code from
target behavior that is only planned. The current `schema.sql`, Go routes and
DTOs, and recorded product decisions are the baseline for implementation;
target SRS text is not proof that a feature exists. Legacy repository/GitHub
flow MUST NOT be deleted until its migration plan, retained-data policy, and
replacement tests are approved. New Package/Level, Forum, leaderboard, and AI
boundaries MUST be introduced incrementally with compatibility adapters where
needed.

## Project Context

This is a Go monorepo using `net/http`, pgx, PostgreSQL, Redis, and Piston. The
public service is `api_gateway`; `auth_service`, `run_service`, and
`submission_service` plus its worker own their current boundaries. `repo_service`
is legacy and is being removed from the target product. The target catalog is
Package/Topic -> Level -> Problem, while the running catalog remains
Subject -> Session -> Lesson -> Problem until migration is complete. Forum and
AI review services are target boundaries and are not assumed to be implemented.

## Development Workflow

Before changing an exported handler, service, repository, worker, or shared
package, contributors MUST inspect its callers and run GitNexus impact analysis
when the index is available. A stale index MUST be reported and refreshed before
relying on its graph results. Changes MUST update affected DTOs, routes, schema,
fixtures, tests, and frontend contracts together. Reviewers MUST verify that
legacy-versus-target claims are accurate and that no hidden test or secret crosses
an API boundary.

## Governance

This constitution is the highest-level project guidance for implementation and
review. `AGENTS.md` supplies repository workflow rules; `schema.sql`, current Go
routes/DTOs, and `docs/DECISIONS.md` supply implementation facts and product
decisions. When they conflict with target SRS text, the change MUST state the
migration required instead of assuming the target behavior already exists.

Amendments require a documented reason, an updated sync impact report, and a
semantic version bump. A MAJOR bump removes or reverses a principle, a MINOR bump
adds or materially expands a principle or governance rule, and a PATCH bump fixes
wording without changing obligations. Every implementation review MUST check
constitution compliance, test evidence, migration impact, and security impact.

**Version**: 1.0.0 | **Ratified**: 2026-11-06 | **Last Amended**: 2026-11-06
