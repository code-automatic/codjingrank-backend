# Frontend Handoff — v2.1

The frontend mock must be generated from [`openapi_fe.yaml`](openapi_fe.yaml). Admin mock screens use [`openapi.yaml`](openapi.yaml). The route inventory and product rules are summarized in [`FRONTEND_API_CONTRACT.md`](FRONTEND_API_CONTRACT.md).

```bash
npx @stoplight/prism-cli mock docs/openapi_fe.yaml -p 4010
```

All operations are marked `x-implementation-status: target-not-implemented`; a mock response is not evidence that the Go Gateway route is live. Use `/api/v1`, UUIDs, camelCase and the shared response envelope. Keep access tokens in memory, send `credentials: include`, and never send `X-User-*` headers.

The frontend must treat Run as synchronous sample execution, Submit as Pending → polling → verdict, and AI review as a separate Accepted-only read model. AI score does not affect leaderboard ranking. Do not build repository creation, GitHub connection, Subject/Session/Lesson navigation, or legacy root routes into the new client.
