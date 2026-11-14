# Frontend API Contract — `/api/v1`

**Contract version:** 2.1.0 · **Status:** target contract with partial local implementation; not a deployed API.

**Published reference:** https://code-automatic.github.io/codjingrank-backend/
([learner](https://code-automatic.github.io/codjingrank-backend/learner.html),
[full/admin](https://code-automatic.github.io/codjingrank-backend/full.html)).
GitHub Pages hosts documentation only; send mock requests to local Prism.

Every path in the tables below is relative to `/api/v1`.

The complete machine-readable contract is [`openapi_fe.yaml`](openapi_fe.yaml) for learner screens and [`openapi.yaml`](openapi.yaml) for the full Gateway, including admin. Both files use OpenAPI 3.0.3, UUID identifiers, camelCase JSON, UTC RFC3339 timestamps, and the approved response envelope. Do not infer runtime availability from this document: the Gateway now registers v1 catalog/admin catalog and Run/Submit. Full v1 Auth/profile/history and other features remain pending. See [Sprint 2 acceptance](SPRINT2_ACCEPTANCE.md) before switching a screen from mocks. Root Auth currently provides UUID tokens for integration; its DTO is not the full v1 Auth contract.

## Rules shared by every endpoint

- Success: `{ "status":"success", "code":200, "message":"...", "data":{} }`.
- Error: `{ "status":"error", "code":400, "message":"...", "error":"VALIDATION_ERROR", "details":null }`. `code` equals HTTP status; `error` is a stable string; `details` is always present and nullable.
- Frontend branches on HTTP status and `error`, never on `message`. `204 No Content` has no body.
- Public base URL is the Gateway. New frontend code uses `/api/v1`; it does not call Auth, Run, Submission, Forum, or AI services directly.
- Protected calls send `Authorization: Bearer <accessToken>` and `credentials: include`. The refresh token is an HttpOnly cookie; the client never reads or stores it. Gateway strips client `X-User-*` headers before forwarding.
- All path IDs are UUIDs. Query pagination is `page` (default 1), `size` (default 20, maximum 100), and response data contains `items`, `page`, `size`, `total` where the endpoint is paginated.
- User-facing payloads never contain password hashes, refresh tokens, hidden test input/output, wrapper code, service tokens, provider keys, raw provider errors, or internal stack traces.
- `Run` is synchronous and executes samples only. `Submit` persists `Pending`, queues Redis work, and is polled until a terminal verdict. AI is a separate read model and never changes verdict, progress, or ranking.

## Authentication and profile

| Method | Path | Auth | Response data |
|---|---|---|---|
| POST | `/auth/register` | Public | `accessToken`, `user`; 201 |
| POST | `/auth/login` | Public | `accessToken`, `user`; 200 |
| POST | `/auth/refresh` | Refresh cookie | Rotated `accessToken`, `user`; 200 |
| POST | `/auth/logout` | Optional refresh cookie | 204 and cleared cookie |
| POST | `/auth/forgot-password` | Public | 202, always generic; Should/schema extension |
| POST | `/auth/reset-password` | Public | 204; Should/schema extension |
| GET | `/users/me` | Bearer | Current `user` |
| PATCH | `/users/me` | Bearer | Updated `user`; Should; schema-backed avatar/bio, Go integration pending |
| PUT | `/users/me/password` | Bearer | 204; revokes refresh tokens |

Registration target policy is 8–72 characters with letters and numbers. The current Go DTO still accepts six characters; this is a migration gap, not a mock-server license to loosen the target contract.

## Landing, languages, catalog, and progress

| Method | Path | Auth | Response data |
|---|---|---|---|
| GET | `/public/landing` | Public | Aggregate stats, active languages, featured packages, top five |
| GET | `/languages` | Public | Active language `items` |
| GET | `/tags` | Public | Tag `items`; Should |
| GET | `/packages` | Bearer | Published package page |
| GET | `/packages/{slug}` | Bearer | Package plus exactly three level states |
| GET | `/packages/{slug}/levels` | Bearer | Dynamic unlock/progress state |
| GET | `/problems` | Bearer | Filtered published problem page |
| GET | `/problems/{slug}` | Bearer | Detail and public samples; 403 `LEVEL_LOCKED` when locked |
| GET | `/problems/{slug}/languages` | Bearer | Supported active languages |
| GET | `/problems/{slug}/template?languageId=` | Bearer | Starter code and filename only |
| GET | `/problems/{slug}/samples` | Bearer | Public sample list |
| GET | `/daily-challenge` | Bearer | UTC challenge; Could |

Canonical catalog names are `packages`, `levels`, and `problems`. The old `subjects`, `sessions`, `lessons`, `categories`, and integer IDs are not part of the new contract.

## Run, submit, history, and AI review

| Method | Path | Auth | Response data |
|---|---|---|---|
| POST | `/run` | Bearer | Synchronous `status`, `passed`, `total`, sample `results`; 200 even for Wrong Answer/CE |
| POST | `/submit` | Bearer | `submission` with UUID and `Pending`; 201 |
| GET | `/submissions/{id}` | Owner/Admin | Submission detail and source; 404 for another user |
| GET | `/me/submissions` | Bearer | Paginated all-verdict history, no source in list |
| GET | `/me/submissions/{problemId}/latest` | Bearer | Latest submission or `submission: null` |
| GET | `/me/stats` | Bearer | Streak, Accepted progress, activity, current rank |
| GET | `/submissions/{id}/test-results` | Owner/Admin | Safe per-test ordinal/verdict/runtime; Could |
| GET | `/submissions/{id}/ai-review` | Owner/Admin | `review: null` before scheduling, then queued/running/completed/failed |

`POST /run` accepts `{problemId, languageId, code}` and never creates a submission. `POST /submit` accepts the same body and returns a durable submission ID; the client polls every 1–2 seconds while `Pending` or `Running`. `GET /submissions/{id}/ai-review` returns 409 `AI_REVIEW_NOT_ELIGIBLE` for non-Accepted submissions. An Accepted submission may briefly return `review: null` while the post-commit AI job is queued.

Leaderboard ranking is independent of AI:

| Method | Path | Auth | Response data |
|---|---|---|---|
| GET | `/leaderboard?scope=all\|week\|month&package=` | Optional Bearer | Page of `rank`, public author, solved and Accepted counts |
| GET | `/leaderboard/me?scope=...&package=` | Bearer | Current entry or `entry: null` |

`all` is public; week/month and pages beyond the guest first page require Bearer. Ties use SQL `RANK()` after solved distinct problems and Accepted solution count. AI score is never an input.

## Forum

| Method | Path | Auth |
|---|---|---|
| GET | `/forum/posts` | Optional Bearer |
| POST | `/forum/posts` | Bearer |
| GET | `/forum/posts/{id}` | Optional Bearer |
| PATCH | `/forum/posts/{id}` | Owner |
| DELETE | `/forum/posts/{id}` | Owner; 204 |
| GET | `/forum/posts/{id}/comments` | Public |
| POST | `/forum/posts/{id}/comments` | Bearer |
| PATCH | `/forum/comments/{id}` | Owner |
| DELETE | `/forum/comments/{id}` | Owner; 204 |
| PUT | `/forum/posts/{id}/like` | Bearer |
| DELETE | `/forum/posts/{id}/like` | Bearer; returns updated like state |

Post creation uses `title`, `contentMd`, `postType`, optional `problemId`, `languageId`, and `codeSnippet`. `solution` posts require Accepted progress for the selected problem. Comments are one level deep. Content is Markdown/plain text sanitized by the server; the frontend must still render through a sanitizer. Like/unlike is idempotent.

## Admin contract

All admin routes require `role=admin`; the complete request and response schemas are in `openapi.yaml`.

| Area | Endpoints |
|---|---|
| Users | `GET /admin/users`, `GET/PATCH /admin/users/{id}` |
| Packages/levels | `GET/POST /admin/packages`, `GET/PATCH/DELETE /admin/packages/{id}`, publish/archive actions, `GET /admin/packages/{id}/levels`, `PATCH /admin/levels/{id}` |
| Problems | `GET/POST /admin/problems`, `GET/PATCH/DELETE /admin/problems/{id}`, publish/archive/clone actions |
| Problem metadata | `GET/PUT /admin/problems/{id}/tags`, `GET/PUT /admin/problems/{id}/languages`, `GET/PUT/DELETE /admin/problems/{id}/templates/{languageId}` |
| Samples/tests | `GET/POST /admin/problems/{id}/samples`, `PATCH/DELETE .../samples/{itemId}`, `GET/POST /admin/problems/{id}/test-cases`, `PATCH/DELETE .../test-cases/{itemId}` |
| Shared catalog | `GET/POST /admin/languages`, `GET/PATCH /admin/languages/{id}`; `GET/POST /admin/tags`, `PATCH/DELETE /admin/tags/{id}` |
| Moderation | `GET /admin/forum/posts`, `GET/PATCH /admin/forum/posts/{id}`, `GET /admin/forum/posts/{id}/comments`, `PATCH /admin/forum/comments/{id}` |
| Dashboard/optional | `GET /admin/stats`, `GET/PUT /admin/daily-challenges` |

Published judge inputs are immutable, including after archive. Use the clone endpoint to create a new draft. `executeCode`, points, hidden expected output, wrapper code, and arbitrary unknown fields are rejected or remain admin-only; they are never learner payloads.

## Errors and mock usage

Common machine codes are defined as reusable responses in both OpenAPI files: `BAD_REQUEST`, `VALIDATION_ERROR`, `UNPROCESSABLE_REQUEST`, `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `METHOD_NOT_ALLOWED`, `CONFLICT`, `REQUEST_TOO_LARGE`, `RATE_LIMITED`, `UPSTREAM_ERROR`, `TIMEOUT`, and `INTERNAL_ERROR`. Endpoint descriptions add `LEVEL_LOCKED`, `LANGUAGE_NOT_SUPPORTED`, `SUBMISSION_COOLDOWN`, `AI_REVIEW_NOT_ELIGIBLE`, `SOLUTION_NOT_ALLOWED`, `INVALID_COMMENT_PARENT`, `PUBLISH_REQUIREMENTS_NOT_MET`, and `PROBLEM_PUBLISHED_LOCKED` where applicable.

For a mock server, use the learner spec for learner screens and the full spec for admin screens. Example:

```bash
npx @stoplight/prism-cli mock docs/openapi_fe.yaml -p 4010
# or
npx @stoplight/prism-cli mock docs/openapi.yaml -p 4010
```

The mock examples are deterministic and deliberately use fake UUIDs/tokens. Select named examples with a request header such as `Prefer: example=accepted`, `Prefer: example=failed`, or `Prefer: example=empty`; select HTTP errors with `Prefer: code=403`. The mock is stateless: it does not advance Pending jobs, persist writes, validate real JWTs, calculate ranks, or enforce ownership/business policies. Those require backend integration tests.

A mock returning 200 does not prove a backend endpoint exists. The Gateway implementation must add `/api/v1` registration, UUID DTOs, ownership checks, cookie-path compatibility, contract tests, and DB/service integration before frontend switches from mock to live.

## Source of truth

1. Product/security decisions: [`DECISIONS.md`](DECISIONS.md)
2. Runtime schema: [`schema.sql`](../schema.sql)
3. Full contract: [`openapi.yaml`](openapi.yaml)
4. Learner contract: [`openapi_fe.yaml`](openapi_fe.yaml)
5. Current-vs-target notes: [`AI_CONTEXT.md`](AI_CONTEXT.md) and [`BACKEND_REALITY_AUDIT.md`](BACKEND_REALITY_AUDIT.md)

Do not use the historical root-route OpenAPI entries or GitHub/repository endpoints for the new frontend.

## Transport design details and schema gaps

These are contract-level design choices made to complete mock coverage on 07/11/2026; they do not claim prior product approval or implementation:

- `GET /packages` and all editor/catalog progression endpoints require Bearer. Guests use landing, languages, tags, public Forum reads and the first all-time leaderboard page. A supplied invalid optional token fails with 401; it is not silently treated as a guest.
- Semantic field validation uses 400 `VALIDATION_ERROR`, matching current shared helpers. Malformed JSON is 400 `BAD_REQUEST`; unsupported language and invalid comment parent are 422 with their endpoint-specific machine codes. Invalid UUID/query/type is 400 `VALIDATION_ERROR`. Unknown JSON fields are rejected rather than ignored. OpenAPI descriptions are the per-endpoint authority for business error codes.
- Registration allows username `[A-Za-z0-9_]{3,50}` and bcrypt passwords of 8–72 characters with a hard 72-byte UTF-8 limit. These exact bounds are transport choices; server validation must enforce bytes as well as JSON Schema character limits. Run/Submit `code` is nonblank and at most 102400 UTF-8 bytes. Preserve source bytes exactly. Forum title/content/comment/code limits are in their request schemas; all text writes reject whitespace-only values.
- Query defaults: `scope=all`, post `sort=newest`, pagination 1/20 (leaderboard 1/50). History is newest `createdAt,id` first; language/tag lists sort by slug/id. Unlisted sort values are invalid, not SQL expressions. All filters combine with AND; missing filters mean no restriction.
- Week starts Monday 00:00 UTC; month starts at the first day 00:00 UTC. Bounds are half-open and returned as `periodStart/periodEnd`; all-time has null bounds. Period ranking counts problems FIRST solved in that period and sums their current distinct Accepted counts. Stable userId ordering does not break equal rank.
- A problem's title/statement may receive editorial corrections after publication; samples, hidden tests, supported languages, templates, execution mode, limits, level and semantic judge behavior are frozen. Admin must clone for changes. Metadata tag links can change. Packages have exactly three levels, created atomically with the package; only level display names are editable.
- Deactivate users/languages instead of deleting history. No language-delete endpoint, no level create/delete, no hard user delete, no rejudge, no comment likes, no reports/tags for Forum and no public AI trigger/retry. Deleted comment content becomes a tombstone; it is not returned to learner or moderation views.
- `avatarUrl` maps to `users.avatar_url` (nullable, maximum 2048 characters); `bio` maps to `users.bio` (nullable, maximum 1000 characters). Omitting a PATCH field leaves it unchanged; null clears it. URI validation belongs to the API; Go DTO/query/handler integration is still pending. Password-recovery endpoints require a new hashed single-use expiring-token store plus email delivery; they remain `Should`, not available runtime routes. Comment moderation uses existing `is_deleted` (no invented hidden state). AI rubric keys/provider/quota/retry remain service-policy decisions; the client renders dimensions dynamically and reads generic safe failures.
- Current HS256 JWT implementation is unchanged. This transport contract does not silently implement the old SRS RS256/JWKS proposal. All admin endpoints enforce current active admin role; protected target routes enforce active account state. Server-side revocation enforcement for existing access tokens requires integration; never claim a refresh-cookie clear alone revokes a JWT.

## Compatibility and implementation gates

This standalone product has no old database to migrate. UUID Auth/catalog/execution and Sprint 3 submission history/latest/stats consumers are integrated. Full v1 Auth/profile still needs implementation. Legacy integer JWTs are rejected. Root/v1 Run, Submit and submission reads share their tested target payloads; removed curriculum/GitHub routes are never restored as aliases. See [Sprint 3 acceptance](SPRINT3_ACCEPTANCE.md).

Gateway strips `/api/v1` only when the corresponding upstream understands the target payload. The pending `GET /users/me` integration is intended to rewrite to Auth `/auth/me`; other mappings follow the OpenAPI `x-owner`. Cookie Path=/ must cover root/v1 refresh/logout during compatibility; remove old Path=/auth cookie, preserve all Set-Cookie headers, validate Origin against the configured allowlist and use SameSite according to deployment. Cross-site cookie deployment requires a CSRF design before enablement. Internal Auth lookup and AI job delivery are not browser endpoints and are intentionally absent from public OpenAPI.

### Compatibility checklist (updated 14/11/2026)

| Earlier root behavior / remaining gap | Target v1 | Current status / integration requirement |
|---|---|---|
| Integer user/problem/language IDs | UUID strings everywhere | Auth/JWT/catalog/Run/Submit/history/stats integrated; full profile/Auth and frontend types still require alignment |
| Logout200 envelope | Logout204 bodyless/idempotent | Root logout remains200; bodyless204 and target v1 Auth route remain pending |
| History data.submissions without page contract | data.items/page/size/total | Implemented at root/v1; clients must read items, support pagination and omit list-source assumptions |
| Latest absence404 | 200 data.submission=null | Implemented at root/v1; do not use404 as the empty-state signal |
| Legacy stats rank.rating | Target progress/rank without rating | Implemented at root/v1; remove rating assumptions; currentRank is nullable and AI score stays separate |
| Root public problem detail | Target catalog requires Bearer/level checks | Implemented v1 authentication/level checks; root problem family retired410 |
| Root problem test-cases (sample projection) | `/problems/{slug}/samples` | Implemented slug route; root family retired410; hidden tests admin-only |
| Six-character current registration | Target8+ with letter/digit | Validation change needs contract/regression tests |

### Feature coverage

| SRS | Mock surface |
|---|---|
| F01 | Auth/profile/password recovery |
| F02 | Landing |
| F03 | Languages |
| F04 | Package/Level/Problem catalog |
| F05 | Starter templates |
| F06 | Synchronous Run |
| F07 | Pending Submit/polling |
| F08 | Level progression |
| F09 | All-verdict history/statistics |
| F10 | Solved-problem leaderboard |
| F11 | Forum posts/comments/likes |
| F12 | Admin users/catalog/moderation/dashboard |
| F13 | Accepted-only AI read |

The full specification contains 94 operations (52 admin); the learner projection contains 42. Every operation has concrete schemas, security, status/error responses and examples. `Should`/`Could` coverage does not promise delivery before the 22/11 deadline.

## Validation evidence — 07/11/2026

Both specs passed OpenAPI 3.0.3 validation. All references, unique operation IDs,
path parameters, F01–F13 coverage, admin security, bodyless204 responses and
example payload schemas were checked. The learner operations and reachable
schemas match the full contract exactly; no admin/hidden/provider schema leaks.
Prism5.16.0 returned declared success envelopes for all94 full and42 learner
operations; 11 named lifecycle/empty examples also passed. A missing Bearer on
catalog returns401 in the mock. These checks validate mockability, not backend
behavior. `go test -count=1 ./...` passed without Go/schema edits.
