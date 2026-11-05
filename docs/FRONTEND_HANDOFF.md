# Frontend Integration Guide (v2.0)

Updated 2026-11-05. The current gateway routes are the source of truth for
what can be called today. The Package/Level, Forum, Leaderboard and AI routes
are target contracts in `docs/FRONTEND_API_CONTRACT.md`; Forum/Leaderboard/AI
are chưa implement and are not live yet.

## Base URL
- Development: `http://localhost:8080`

## API Documentation
- Swagger UI: `http://localhost:8080/docs/fe`
- OpenAPI Spec: `http://localhost:8080/docs/openapi.yaml`

## Auth Flow
1. POST /auth/register or POST /auth/login → get accessToken
2. Authorization: Bearer <accessToken> on protected routes

## Response Envelope
All responses use: `{ status, code, message, data }` for success, `{ status, code, message, error, details }` for errors.

## Business Statuses
- submission.status: Pending, Running, Accepted, Wrong Answer, etc.
- run.status: Accepted, Wrong Answer, Compilation Error, etc.
- aiReview.status (target): queued, running, completed, failed

Accepted is returned from the judge without waiting for AI. AI is requested
only after Accepted and its score is never used for ranking. Run and non-
Accepted verdicts do not request a review.

## Removed target flow

Do not build repository creation, GitHub connection, commit links, or webhook
screens. Those routes remain in the legacy OpenAPI snapshot only and are being
removed from the product.

## Common Routes
See Swagger UI for full API reference.
