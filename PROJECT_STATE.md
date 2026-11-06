# Project State

Updated: 2026-11-06
Branch/commit: `feat/database-schema`, based on `main` at `b6ddc18`

## Objective

Reconcile the new SRS in `code-automatic/srs` with this legacy backend and
provide an accurate frontend API handoff. Current task: establish the clean
shared database baseline for the standalone product.

## Completed

- Read current routes, DTOs, services, worker, schema, OpenAPI and Docker setup.
- Used GitNexus repo context/query/context and Semble search to cross-check the
  code paths; GitNexus index is current but the repo has no HEAD.
- Added `docs/BACKEND_REALITY_AUDIT.md` with evidence-backed conflicts and a
  P0–P3 migration order.
- Added `docs/FRONTEND_API_CONTRACT.md` with current routes, payloads and
  proposed `/api/v1` Forum/Leaderboard/AI contracts.
- Added the Gateway v1 façade route matrix and explicit upstream rewrites:
  current auth/catalog/run/submit routes are mapped only where handlers exist;
  Package, Forum, Leaderboard and AI routes remain target-only.
- Reconciled SRS Run/Submit requirements and workflow diagrams with the live
  synchronous Run and submission UUID polling behavior; legacy `run_id`/`job_id`
  examples are labeled as non-live.
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

## In progress

- Database design is drafted in
  `docs/superpowers/specs/2026-11-06-database-schema-design.md`.
- The agreed architecture is one clean shared PostgreSQL database for the
  standalone project; old hierarchy and repository/GitHub tables are excluded.
- Commit `0c78986` replaces `schema.sql` with the approved clean target schema. Go
  handlers/DTOs still need a separate integration task; source behavior has not
  been rewritten in this task.

## Verification

- `go test ./...` passed after the final documentation edits; no code changed.
- `git diff --check` and YAML parsing for `docs/openapi.yaml` and
  `docs/openapi_fe.yaml` passed after the v1 contract edits. `make ci-openapi`
  skipped lint because Redocly is not installed.
- Relative Markdown links, code-fence and whitespace scans passed. The only
  initial whitespace findings were pre-existing spaces in `docs/openapi.yaml`;
  touched lines were cleaned and the scan reran.
- Schema commits: `0c78986` (baseline), `5c01950` (publication guards),
  `8567ef0` (repeatable SQL checks). No push or merge performed.
- PostgreSQL 16 bootstrap verified: 22 tables, 36 foreign keys, no legacy
  hierarchy/repository tables. The test runner creates and drops only its own
  randomly named temporary database.
- `SCHEMA_TEST_CONTAINER=hapi_db_container SCHEMA_TEST_USER=warmdev bash
  tests/schema/run.sh` passed at `8567ef0`: bootstrap twice, incomplete
  publication rejection, published/archived test immutability, Accepted
  uniqueness, non-Accepted AI rejection and AI failure isolation.
- Earlier inline DO-block checks lacked `docker exec -i` and did not execute;
  their invariant claims are superseded by the repeatable test above.
- `go test ./...` passed at `8567ef0` (many results cached). This does not prove
  legacy Go queries work against the target database.
- Review performed by the primary agent; no independent review performed.
- Existing documentation changes and `outputs/` predate this schema task and
  remain separate from the schema commits.
- `AGENTS.md` existed before this task and now includes branch/commit discipline.
- `semble search 'leaderboard or AI code review implementation' . -k 10 --include-text-files`
  found no implementation outside the new proposed contract.
- `gitnexus_detect_changes(scope=all, base_ref=main)` reported low risk and no
  affected execution processes for the schema-only change.

## Next steps

Next implementation work: adapt Go schema consumers/API while preserving
Run/Submit/Judge, then implement Forum and AI review services. Provider,
rubric, quota and retry service semantics remain separate follow-up work.
Do not delete remote GitHub repositories. Keep AI out of ranking calculations.

## Integration limits

- `schema.sql` is a destructive clean bootstrap, not a data-preserving migration.
  Never run it on an existing data-bearing application database.
- Existing Go DTOs, SQL queries, seeds, historical migrations and API examples
  still contain integer IDs and legacy names. Do not apply old migrations or
  seeds to this target database. No live service integration was tested.
- Progress upserts, recovery/re-enqueue, comment parent validation, hidden-data
  projection, counter maintenance and transaction coordination remain service
  responsibilities; the schema alone does not implement those features.
- Leaderboard view covers all time; week/month/package query logic is follow-up.
- SQL publication protections were checked sequentially; concurrent admin and
  worker operations and database access roles have not been validated.
