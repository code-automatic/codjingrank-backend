### Refresh Token Flow
- **Decision**: accessToken returned in JSON body (short-lived, 15 min). refreshToken stored in HttpOnly cookie (7 days).
- **Storage**: Only SHA-256 hash stored in `refresh_tokens` table. Raw token never persisted.
- **Rotation**: Each `/auth/refresh` revokes the old token and issues a new one.
- **Logout**: Revokes token in DB and clears cookie.
- **Security**: HttpOnly prevents XSS access. Secure=true in production. SameSite=Lax default.
