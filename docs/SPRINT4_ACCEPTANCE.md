# Sprint 4 backend acceptance

Status: verified locally on `feat/sprint4-backend`, with uncommitted changes
on top of `bcbe9a1`. Provider tests use a local HTTP double.

## Scope and evidence

| Workbook task | Backend deliverable | Current evidence/status |
|---|---|---|
| 24 | AI schema and internal contract | [Schema](../schema.sql), [HTTP handlers](../ai_service/internal/review/http.go). Schema tests pass. AI read test covers owner access and outsider 404. Internal enqueue rejects missing credentials. |
| 25 | Queue after persisted Accepted | [Worker integration test](../submission_service/internal/worker/ai_test.go) observes Accepted in PostgreSQL at notification time. Wrong Answer and duplicate delivery do not notify. An enqueue 503 leaves the verdict intact. [AI integration test](../ai_service/internal/review/worker_test.go) verifies reconciliation and completion using a local provider. |
| 26 | Failure isolation and idempotency | The same test verifies the enablement cutoff, Wrong Answer exclusion, a single review after repeat reconciliation, failed review after three HTTP 503 attempts with Accepted intact, and no retry for HTTP 400. A concurrent worker cannot claim an active review, and a stale attempt cannot overwrite the recovered result. |
| 27 | Leaderboard | [DB and HTTP tests](../api_gateway/internal/leaderboard/http_test.go) cover UTC boundary, ties, rank outside page, empty page/rank, inactive users and guest limits. |
| 30 | Forum | [DB and HTTP tests](../forum_service/internal/forum/forum_test.go) cover posts, editing with a single database connection, comments, nesting, tombstones, like/unlike, view deduplication, solution gate and moderation. [Markdown tests](../forum_service/internal/forum/markdown_test.go) cover unsafe destinations and raw HTML escaping. |
| 32 (BE portion) | Contract and integration | 18 successful learner responses validated against both OpenAPI files and 5 moderation responses against the full spec. Startup smoke passed for Gateway, Forum and AI. CI includes response capture. |

## AI provider protocol and configuration

AI remains disabled unless `AI_PROVIDER_URL`, `AI_PROVIDER_MODEL`, and
`AI_ENABLED_SINCE` are set. `AI_INTERNAL_SECRET` protects the internal enqueue
route. `AI_PROVIDER_TOKEN` is optional. The cutoff uses RFC3339. The worker
uses `AI_SERVICE_URL` and the same internal secret. The service defaults to
port 8085; Forum defaults to 8084.

`AI_PROVIDER_URL` is the full OpenAI-compatible chat completions endpoint
(for example `https://opencode.ai/inference/go/openai/v1/chat/completions`), not
a base URL. The adapter sends `model`, `temperature: 0`,
`response_format: {"type":"json_object"}`, and system/user messages; the user
message carries the submitted source, language, and public problem statement as
untrusted data. The assistant message content must be a JSON object with
numeric `score` (0–100), nonempty `summary`, numeric `dimensions` object, and
optional `findings` containing severity, title, explanation, optional
suggestion, and optional line bounds. Responses are limited to 1 MiB and the
request timeout is 30 seconds. The adapter speaks the OpenAI-compatible chat
completions protocol; 429 and 5xx responses are retried, while other 4xx
responses are permanent. No live-provider acceptance has been performed.

## Reproduce verification

Use only the dedicated disposable Sprint 4 PostgreSQL/Redis instance. Set
`CATALOG_TEST_DATABASE_URL` and `CATALOG_TEST_REDIS_ADDR` in the environment.

```bash
make gofmt-check
go test -count=1 ./...
go test -race -count=1 ./ai_service/... ./forum_service/... ./api_gateway/internal/leaderboard ./submission_service/internal/worker
go vet ./...
go build ./...
docker compose config --quiet
make ci-openapi
```

Follow the capture commands in [backend CI](../.github/workflows/backend-ci.yml)
to reproduce schema validation. Schema bootstrap and constraint checks use
a temporary database created by `tests/schema/run.sh`.

Outstanding verification includes a real AI provider, frontend acceptance and
live notification from a real Piston judge. This change updates the fresh schema
bootstrap. An existing database needs an additive upgrade before deployment.
