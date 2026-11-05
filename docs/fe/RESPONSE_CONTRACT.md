# Response Contract

The frontend and backend use a standard JSON envelope for all API responses.

## Success

Success responses will have an HTTP 2xx status code and the following structure:

```json
{
  "status": "success",
  "code": 200,
  "message": "Human-readable endpoint-specific message here",
  "data": { ... }
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
- **`message`**: A human-readable message that can be displayed to the user.
- **`error`**: A machine-readable, stable string code representing the error type (e.g., `"VALIDATION_ERROR"`, `"NOT_FOUND"`, `"UNAUTHORIZED"`). Business logic on the frontend should branch on this field if needed.
- **`details`**: Optional field containing additional context, such as a map of field names to validation errors.

## Frontend Logic Guidelines
- **HTTP status code must match `body.code`**. Use the HTTP status code for transport-level handling (e.g., retries, generic error boundaries).
- Use `status` to definitively verify a successful response envelope.
- Use the `body.error` field to programmatically handle specific machine-readable app errors (e.g., redirecting to login on `"UNAUTHORIZED"`).
- Only use `message` for display purposes (toasts, alerts, inline errors), never for conditional logic.
