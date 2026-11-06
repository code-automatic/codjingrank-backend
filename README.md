# Codjingrank Backend

Go monorepo for the standalone coding practice platform. Product baseline:
[decisions](docs/DECISIONS.md), [SRS](docs/srs/01_SRS_Coding_Judge_Platform.md)
and [frontend contract](docs/FRONTEND_API_CONTRACT.md).

## Structure and current status

- `api_gateway/`: public entrypoint, auth proxies, problem and admin handlers (:8080).
- `auth_service/`: password authentication, JWT and refresh cookies (:8081).
- `run_service/`: synchronous sample execution using Piston (:8082).
- `submission_service/`: Pending submission API (:8083), Redis list queue and worker.
- `shared/`: configuration, models, response envelope, middleware and judge helpers.
- `schema.sql`: clean shared PostgreSQL 16 bootstrap for Package → Level → Problem.
- `tests/schema/`: bootstrap and constraint checks.

**Integration is pending:** surviving Go DTOs and SQL still use legacy integer IDs
and column names. They compile but are not yet compatible with the new UUID schema.
Do not treat the stack as an operational new-product release.
The `/api/v1` façade, Forum, AI review and target leaderboard are not implemented.

Repository/GitHub integration, Subject/Session/Lesson endpoints, old migrations,
old seeds and the legacy Vite test UI have been removed. No replacement AI call
has been added; the worker currently stops after persisting its verdict.

## Development commands

```bash
make fmt                     # gofmt
make test                    # Go unit/contract tests
make run-gateway             # start gateway
make run-auth                # start authentication service
make run-runner              # start synchronous runner
make run-submission          # start submission API
make run-submission-worker   # start queue consumer
make compose-up              # PostgreSQL, Redis and Piston infrastructure
```

Config comes from environment variables in `shared/config/config.go`.
Compose definitions exist for the retained services; starting them does not apply
schema or complete schema integration. Judge behavior remains synchronous Run and
Submit → Redis LPUSH/BRPOP → worker → Piston → persisted verdict.

## Database checks

Only use a dedicated PostgreSQL 16 test container:

```bash
SCHEMA_TEST_CONTAINER=codjingrank-schema-test SCHEMA_TEST_USER=postgres \
  bash tests/schema/run.sh
```

The runner creates and drops its own temporary database. `schema.sql` is a
**destructive clean bootstrap**, never a migration for an existing database.
Historical migrations and demo seeds must not be applied to the new product.

## Frontend mock contract

Use [learner OpenAPI](docs/openapi_fe.yaml) for learner screens and
[full Gateway/admin OpenAPI](docs/openapi.yaml) for admin screens. The
[frontend contract](docs/FRONTEND_API_CONTRACT.md) covers F01–F13, permissions,
examples and schema-integration gaps. These are target `/api/v1` contracts;
current Go routes do not implement them yet.

```bash
npx @stoplight/prism-cli mock docs/openapi_fe.yaml -p 4010
```
