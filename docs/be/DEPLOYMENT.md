# Docker and CI baseline

Updated 11/11/2026. Sprint 1 task6 uses the retained Gateway, Auth, Run,
Submission API, submission worker, PostgreSQL16, Redis7 and real Piston.
Current API availability is in [Sprint 2 acceptance](../SPRINT2_ACCEPTANCE.md).

## Requirements and isolation

Use Linux with pure cgroup v2, Docker Engine and Compose >=2.24.4 (`!reset` and
`!override` support). Piston is digest-pinned and requires privileged operation;
use a dedicated execution host. Do not expose its management API publicly.
Do not reuse HAPI or another project's database, containers or volumes.
Compose now uses project-scoped names; specify `-p` consistently. Existing
`innogen-*` containers/volumes are not automatically adopted or deleted.

```bash
# Disposable trusted end-to-end verification; requires internet for runtime install.
make docker-smoke
# Unit/integration tests need their dedicated CATALOG_TEST_* environment variables.
make ci-go
go vet ./...
go build ./...
make ci-openapi
make ci-docker
```

`docker-smoke` creates a unique project with random loopback Gateway/Piston
ports, no other published ports, a generated test JWT secret and fresh volumes.
It builds/starts all services, installs the exact Python3.10.0 seed runtime and
uses only trusted code fixtures. It checks UUID auth/refresh/catalog, sample Run,
Pending submission in a Redis list, worker judging and persisted verdict, plus
ownership and execution limit probes. It removes only its own project/volumes.
A failed installation or check exits nonzero; allow several minutes for builds
and downloads. Standard Go tests do not run this privileged smoke automatically.

## Local stack and fresh schema

```bash
# Set a local JWT_SECRET before starting; the Compose default is development only.
docker compose -p codjingrank-local up -d --build --wait
make piston-install  # Python3.10.0, default PISTON_BASE_URL=http://localhost:2000
# Stop without deleting persistent data:
docker compose -p codjingrank-local down
```

On an **empty PostgreSQL volume only**, the upstream entrypoint applies
`schema.sql` before its TCP readiness check passes. Existing volumes are never
reset or upgraded automatically. `schema.sql` is destructive clean bootstrap,
not a migration. Back up existing data and plan explicit schema changes; do not
use `down -v` to fix a production database. A partial bootstrap failure needs
inspection before any reset. Infrastructure-only `make compose-up` does not
install a Piston language. A healthy Piston API does not prove a runtime exists.
Installing an already installed language can return an error; inspect
`GET /api/v2/runtimes` before reinstalling.

Base Compose binds every diagnostic port to127.0.0.1. Production override
replaces Gateway ports and removes all other host ports instead of inheriting
base publications:

```bash
docker compose -p codjingrank-local -f docker-compose.yml \
  -f docker-compose.prod.yml config --quiet
```

This boundary is not full production readiness. Replace development DB/JWT
credentials, configure TLS/CORS/cookies, isolate internal services and apply
host controls before deployment. Existing Auth logout still returns200; target
v1 Auth/profile/history and worker recovery are separate incomplete features.

## CI scope

`.github/workflows/backend-ci.yml` checks Go formatting/tests/vet/build with
PostgreSQL/Redis, repeatable schema constraints, both OpenAPI files, production
port boundaries and all five Go Docker builds on pushes to main and PRs.
A separate real-Piston smoke job runs only on manual `workflow_dispatch` because
it requires privileged Docker and downloads an external language package.
`.github/workflows/api-docs.yml` remains the API reference publishing workflow.
`make ci-openapi` now propagates lint errors instead of printing a success-like
skip message. Local checks do not establish a successful GitHub Actions run;
no workflow was pushed or triggered by this task.

## Known execution limits

Piston disables job networking and sets wall/CPU timeout and fallback memory
limits; Go forwards per-problem wall timeout and memory. Output is globally
bounded by Piston's default buffer, **not yet per-problem `outputLimitBytes`**.
Per-problem CPU policy and concurrent-load/security testing remain separate work.
Piston `cpu_time` is milliseconds but the existing judge conversion treats it as
seconds, inflating displayed runtime; record/fix this as a separate regression.
These checks exercise trusted fixtures, not an adversarial sandbox certification.

## Task6 acceptance evidence — 11/11/2026

Local checks on `chore/docker-ci-runtime` (based on `3e2f9bf`): all five Go
images built; DB/Redis-enabled `go test -count=1 ./...`, `go vet ./...`,
`go build ./...`, schema bootstrap twice/constraints and both pinned Redocly
lints passed. Workflow passed actionlint1.7.7. Forced npx and curl exit7 probes
now fail their Make targets; both previously masked failures were removed.
Merged Compose JSON confirms project-scoped container names and production's
single loopback Gateway publication. No Go application source changed.

The trusted real-Piston smoke was run locally with a newly generated project,
fresh volumes and Python3.10.0: all services healthy, UUID auth/refresh and
catalog working, one synchronous sample Accepted without submission persistence,
Pending observed in Redis `submission_jobs` list, worker resumed to persist
Accepted for both hidden tests, independent wrong-code submission Wrong Answer,
and outsider detail404. Piston timeout/memory/network/output probes passed.
The stack and volumes were removed on exit; HAPI was not touched.

Independent reviewer session disconnected without a completed result. This is
an author-reviewed, locally validated baseline, not independent merge approval,
a successful remote Actions run or production security certification.
