# Auth Flow for Frontend — `/api/v1`

## Login and registration

```http
POST /api/v1/auth/login
Content-Type: application/json

{"email":"learner@example.com","password":"Coding2026"}
```

Response data is `{ accessToken, user }` inside the standard success envelope. The access token is held in memory. The server sets `refreshToken` as an HttpOnly cookie; JavaScript cannot read it.

## Refresh and logout

```http
POST /api/v1/auth/refresh
Cookie: refreshToken=...
```

Call refresh with `credentials: "include"` after a 401, retry the original request once, and prevent concurrent refresh calls with a single-flight queue. Refresh rotates the cookie. Logout is `POST /api/v1/auth/logout`; it returns 204 and clears the shared cookie path.

## Registration payload

`POST /api/v1/auth/register` accepts `email`, `username`, `password`, and optional `fullName`. The target password policy is 8–72 characters with at least one letter and number. `role`, `isActive`, tokens, and password hashes are server-owned.

## Profile

`GET /api/v1/users/me` reads the current user. `PATCH /api/v1/users/me` changes only `fullName`, `avatarUrl`, or `bio`; those optional profile fields require the additive schema extension described in the OpenAPI contract. `PUT /api/v1/users/me/password` revokes refresh tokens and returns 204.

The frontend never sends `X-User-ID`, `X-User-Email`, or `X-User-Role`; Gateway derives them from the verified token. Do not store access or refresh tokens in localStorage.
