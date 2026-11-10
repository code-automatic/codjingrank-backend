# Frontend Handoff — v2.1

**API docs:** https://code-automatic.github.io/codjingrank-backend/

- [Learner reference](https://code-automatic.github.io/codjingrank-backend/learner.html): 42 endpoints.
- [Full reference](https://code-automatic.github.io/codjingrank-backend/full.html): 94 endpoints including 52 admin.
- The public site publishes only generated API references and OpenAPI contracts,
  with synthetic examples. Backend source remains private. Pages is not an API server.
- Publication is automatic after contract/site changes merge into `main` via
  `.github/workflows/api-docs.yml`; check the successful deployment before using
  the public copy as the latest revision.

The frontend mock must be generated from [`openapi_fe.yaml`](openapi_fe.yaml). Admin mock screens use [`openapi.yaml`](openapi.yaml). The route inventory and product rules are summarized in [`FRONTEND_API_CONTRACT.md`](FRONTEND_API_CONTRACT.md).

```bash
npx @stoplight/prism-cli mock docs/openapi_fe.yaml -p 4010
```

All operations are marked `x-implementation-status: target-not-implemented`; a mock response is not evidence that the Go Gateway route is live. Use `/api/v1`, UUIDs, camelCase and the shared response envelope. Keep access tokens in memory, send `credentials: include`, and never send `X-User-*` headers.

The frontend must treat Run as synchronous sample execution, Submit as Pending → polling → verdict, and AI review as a separate Accepted-only read model. AI score does not affect leaderboard ranking. Do not build repository creation, GitHub connection, Subject/Session/Lesson navigation, or legacy root routes into the new client.
