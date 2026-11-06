# API Envelope Implementation Plan

> Execute inline with `executing-plans`; the user approved the DECISIONS envelope and requested implementation.

**Goal:** Apply the approved JSON envelope consistently without changing endpoint payloads, pagination, judging or `/api/v1` routing.

**Spec:** [DECISIONS](../../DECISIONS.md#api-response-envelope-decision--confirmed-2026-11-07), [response contract](../../fe/RESPONSE_CONTRACT.md).

**Architecture:** Reuse `shared/response`. Add `response.Router(*http.ServeMux) http.Handler` to replace only generated 404/405 bodies, preserving Allow, redirect behavior, route matching, HEAD and documentation responses. Wire all four HTTP entrypoints and the nested admin mux. No response buffering or double wrapping of upstream payloads.

## Constraints and review focus

- Preserve success/error field sets, numeric HTTP code, string machine error and nullable details.
- Retain 204/HEAD bodyless behavior; do not convert redirects or documentation into JSON.
- Preserve auth before admin dispatch; retain Allow for 405 and PathValue extraction.
- Test Gateway proxy transport failure and middleware errors without DB/Redis/Piston.
- Existing uncommitted documentation and outputs must not enter unrelated commits.
- PR integration uses merge commits, preserving branch commit SHAs; no squash/rebase.

## Tasks

### 1. Shared response/router contract

Files: `shared/response/response.go`, new `shared/response/router.go`, response tests.

- [x] Add exact JSON key/type assertions for success, errors and validation; verify business verdict stays inside data.
- [x] Add failing tests for generated 404/405, Allow, HEAD, redirects, PathValue, untouched HTML and Success(204).
- [x] Run affected tests to record red, implement minimal router handling and 204 guard, rerun green.
- [x] Commit shared behavior and tests separately.

### 2. Service integration and boundary verification

Files: Gateway/Auth/Run/Submission HTTP mains, admin route; Gateway proxy/middleware/admin tests.

- [x] Wrap root and admin mux dispatch with `response.Router`.
- [x] Test admin routing/auth order, auth failure, rate limit, panic, CORS 204 and failed proxy transport.
- [x] Run `gofmt -s -w .`, `go test ./...`, `go vet ./...`, `go build ./...`, `git diff --check`.
- [x] Requested read-only review twice; reviewer transport failed both attempts. Primary reviewed dispatch/auth/HEAD/redirect behavior; no independent review available.
- [x] Commit integration separately; record evidence and remaining DB integration limits in PROJECT_STATE.

### 3. Documentation and handoff

- [x] Commit only envelope-specific documentation hunks; preserve prior changes.
- [x] Record merge-commit preference in AGENTS. Check existing branches against main.
- [x] No push/merge until authorized for this branch; report PR readiness and blockers.
