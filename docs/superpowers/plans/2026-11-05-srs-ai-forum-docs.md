# SRS and AI/Forum Documentation Update Plan

> **For agentic workers:** Documentation-only execution; no product code or schema migration is implemented in this task.

**Goal:** Import the SRS source documents and align the SRS, team plan, backlog, schedule, and frontend API contract with the approved flow: preserve Run/Submit/Judge, remove Repository/GitHub, run AI review only after Accepted, and add Forum.

**Architecture:** Keep the current Go services and judge mechanics as the implementation baseline. Replace the post-judge repository commit integration with an internal AI service call after an Accepted result; add Forum as a separate service boundary. Treat `/api/v1` and new schema names as the target contract, not live routes until implemented.

**Tech Stack:** Markdown, GitHub source repository `code-automatic/srs`, current Go/Redis/PostgreSQL/Piston backend references.

**Spec:** `docs/srs/01_SRS_Coding_Judge_Platform.md`, plus the related imported documents in `docs/srs/`.

## Global Constraints

- Do not modify the existing `AGENTS.md`.
- Do not change backend code, migrations, or remote GitHub repositories.
- Preserve current Run, Submit, Redis queue, worker, Piston, and verdict mechanics in the target description.
- AI review is automatic only after `Accepted`; its score never affects verdict or leaderboard.
- The schedule starts 06/11/2026 and ends 22/11/2026.

## Review Focus

- Every Repository/GitHub requirement is removed or marked legacy/removal.
- No frontend contract advertises unimplemented AI, Forum, or Leaderboard routes as live.
- All imported schedule dates and milestone ranges fit 06/11–22/11.
- Existing API paths and response envelope remain documented separately from the target `/api/v1` contract.

---

### Task 1: Import SRS source

**Files:** Create `docs/srs/01...04...md` and `docs/srs/SOURCE.md`.

- [x] Download the private `master` snapshot and preserve the four source filenames.
- [x] Record repository URL, source branch, commit, and import date.

### Task 2: Rewrite target SRS boundaries

**Files:** Modify `docs/srs/01_SRS_Coding_Judge_Platform.md`.

- [x] State current mechanics are retained.
- [x] Remove Repository/GitHub from scope and replace the post-judge hook with AI-after-Accepted.
- [x] Add Forum and AI review requirements and persistence/API boundaries.
- [x] Reconcile schema and route names with the target Package/Level model.

### Task 3: Sync team documents and schedule

**Files:** Modify `docs/srs/02_PHAN_CONG_NHIEM_VU.md`, `03_BACKLOG_THEO_SPRINT.md`, `04_KE_HOACH_TICH_HOP.md`.

- [x] Reassign repository work to AI service and Forum.
- [x] Remove GitHub/repository stories and add AI/forum stories.
- [x] Rewrite the execution window as 06/11/2026–22/11/2026 with explicit milestones.

### Task 4: Sync frontend and project context

**Files:** Modify `docs/FRONTEND_API_CONTRACT.md`, `docs/AI_CONTEXT.md`, `PROJECT_STATE.md`.

- [x] Mark current routes versus target `/api/v1` routes.
- [x] Document automatic AI review after Accepted and Forum routes.
- [x] Remove frontend dependency on Repository/GitHub.

### Task 5: Verify documentation

- [x] Check Markdown links, date range, forbidden live repo claims, and trailing whitespace.
- [x] Run `go test ./...` only as a regression check; no code changes are expected.
