# Local verification

- `make ci-go`: format check and Go tests.
- `make ci-docker`: build retained service images.
- `make ci-openapi`: existing optional OpenAPI lint (currently masks failures).
- `make ci`: the above commands.

The legacy Vite UI build target was removed with `tools/test_ui`.
No GitHub Actions workflow is currently checked in; local targets do not prove
an automated pipeline exists. Go tests do not prove new-schema integration.

## Baseline audit — 2026-11-08

Checked on `main` `68b0244` with a Docker-capable local environment:

- `make ci-go`, `go vet ./...`, `go build ./...`, `gofmt` check and `git diff --check` passed.
- `docker compose config --quiet` and the production override parse passed.
- `docker compose build` built all retained images successfully.
- `make ci-openapi` completed, but its exit status is not a reliable lint gate because it masks failures. See the findings below.
- `gh api .../actions/workflows` returned zero workflows. There is no GitHub Actions CI pipeline in this repository.

Known blockers for calling the Docker/CI baseline production-ready:

1. `docker-compose.prod.yml` uses `ports: []`, which does not remove ports inherited from the base file. The rendered production config still publishes Postgres, Redis, Piston, Auth, Run and Submission; Gateway has both the public bind and the localhost override. Replace inherited ports with a network-only design before production deployment.
2. Piston uses `privileged: true` and the `latest` image tag. This requires a separate sandbox-hardening decision and pinned image validation.
3. `make ci-openapi` masks any Redocly failure with `|| echo` and exits zero. CI must fail on lint errors once the lint tool is installed or invoked through a pinned tool runner.
4. No `.dockerignore` is present while the builder uses `COPY . .`; local environment files, Git metadata and outputs enter the Docker build context. Exclude secrets and unrelated local artifacts before using remote builders or shared caches.
5. No schema bootstrap is wired into Compose startup. The stack can start without applying `schema.sql`; schema integration remains a separate readiness gate.

This audit does not start the full stack, mutate the HAPI containers, or claim live Postgres/Redis/Piston/judge integration.

Resume verification: Go checks and Compose rendering were repeated successfully. All five built application images declare non-root `appuser`. Existing images/build logs establish image build only, not full-stack runtime readiness.
