# Project State

Updated: 2026-11-05
Branch/commit: `main`, no commit exists yet

## Objective

Reconcile the new SRS in `code-automatic/srs` with this legacy backend and
provide an accurate frontend API handoff, including the missing leaderboard and
AI-review boundaries.

## Completed

- Read current routes, DTOs, services, worker, schema, OpenAPI and Docker setup.
- Used GitNexus repo context/query/context and Semble search to cross-check the
  code paths; GitNexus index is current but the repo has no HEAD.
- Added `docs/BACKEND_REALITY_AUDIT.md` with evidence-backed conflicts and a
  P0–P3 migration order.
- Added `docs/FRONTEND_API_CONTRACT.md` with current routes, payloads and
  proposed `/api/v1` Forum/Leaderboard/AI contracts.
- Corrected `docs/AI_CONTEXT.md` so future agents know Redis/GitHub already
  exist and do not confuse target SRS behavior with current behavior.
- Product decisions recorded: migrate to Topic/Package → Level; AI review is
  separate and does not affect leaderboard ranking; preserve current Run/
  Submit/Judge mechanics; remove repository/GitHub flow and replace its
  post-judge hook with an automatic internal AI service call only after Accepted;
  forum is a new service. Run and other verdicts do not trigger AI; AI failure
  does not change Accepted or ranking. Documentation only; no implementation yet.
- Imported the `code-automatic/srs` snapshot into `docs/srs/` and updated SRS,
  assignments, backlog, integration schedule, frontend handoff, OpenAPI
  legacy notes, architecture/deployment notes, progress and next steps.
- Authoritative documentation window is 06/11/2026–22/11/2026; feature freeze
  is 19/11 21:00. Repository/GitHub work is marked legacy/removal.

## Verification

- `go test ./...` passed after the final documentation edits; no code changed.
- Relative Markdown links, code-fence and whitespace scans passed. The only
  initial whitespace findings were pre-existing spaces in `docs/openapi.yaml`;
  touched lines were cleaned and the scan reran.
- No commit reference exists because the repository has no `HEAD`.
- `AGENTS.md` existed before this task and was not modified.
- `semble search 'leaderboard or AI code review implementation' . -k 10 --include-text-files`
  found no implementation outside the new proposed contract.
- `gitnexus_detect_changes` could not run because this repository has no `HEAD`.

## Next steps

Next implementation work: adapt schema/API while preserving Run/Submit/Judge,
remove repository/GitHub flow, add Forum and AI review. Provider, rubric,
quota, persistence and retry semantics remain to be decided before coding.
Do not delete remote GitHub repositories. Keep AI out of ranking calculations.
