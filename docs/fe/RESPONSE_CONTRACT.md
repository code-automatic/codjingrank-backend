# Response Contract

Confirmed on 2026-11-07 against the
[API response envelope decision](../DECISIONS.md#api-response-envelope-decision--confirmed-2026-11-07).
This contract applies to JSON API responses, including the planned `/api/v1`.
The existing `shared/response` helpers implement this shape; this document covers router-generated 404/405 and existing middleware/proxy failures through
contract tests. Live database integration remains pending. `204 No Content` has no
body; documentation/file responses are outside the JSON envelope.

## Success

Success responses will have an HTTP 2xx status code and the following structure:

```json
{
  "status": "success",
  "code": 200,
  "message": "Human-readable endpoint-specific message here",
  "data": { "items": [] }
}
```

- **`status`**: Always exactly `"success"`.
- **`code`**: Matches the HTTP status code (e.g., 200, 201).
- **`message`**: A human-readable message describing the outcome. Examples include "Problem detail retrieved successfully" or "Submission created successfully". The frontend can use this string directly to show in a toast or snackbar message to the user. **Note: The frontend must NOT branch business logic based on the exact text of `message`, as it may change to improve phrasing.**
- **`data`**: The payload containing the requested resource or result. Shape is endpoint-specific. Keys are `camelCase`.

## Error

Error responses will have an HTTP 4xx or 5xx status code and the following structure:

```json
{
  "status": "error",
  "code": 404,
  "message": "Problem not found",
  "error": "NOT_FOUND",
  "details": null
}
```

- **`status`**: Always exactly `"error"`.
- **`code`**: Matches the HTTP status code (e.g., 400, 404, 500).
- **`message`**: A safe English message that can be displayed to the user.
- **`error`**: A machine-readable, stable string code representing the error type (e.g., `"VALIDATION_ERROR"`, `"NOT_FOUND"`, `"UNAUTHORIZED"`). Business logic on the frontend should branch on this field if needed.
- **`details`**: Always present; `null` when no safe context is available.
  Current validation errors use `{ "field": "email" }`. No raw SQL, stack
  traces, hidden tests, tokens or provider secrets may appear here.
- **`error` is a string**, not `{code, message}`. The numeric `code` remains
  the HTTP status; it is not the machine-readable application error code.

## Business states and payloads

Run verdicts, submission states and AI review states belong inside `data`.
For example, successfully retrieving a submission with `status: "Wrong Answer"`
still uses envelope `status: "success"`. AI failure does not change Accepted.
Endpoint payload and pagination rules remain endpoint-specific; this confirmation
adds no `meta`, `requestId` or other top-level fields. Wrap a payload only once.

## Frontend Logic Guidelines
- **HTTP status code must match `body.code`**. Use the HTTP status code for transport-level handling (e.g., retries, generic error boundaries).
- Use `status` to definitively verify a successful response envelope.
- Use the `body.error` field to programmatically handle specific machine-readable app errors (e.g., redirecting to login on `"UNAUTHORIZED"`).
- Only use `message` for display purposes (toasts, alerts, inline errors), never for conditional logic.
