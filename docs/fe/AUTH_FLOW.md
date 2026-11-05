# Auth Flow for Frontend

## Tokens

| Token | Storage | Accessible by JS | Purpose |
|-------|---------|-----------------|---------|
| `accessToken` | JSON response → FE stores in memory/localStorage | Yes | API authorization (Bearer header) |
| `refreshToken` | HttpOnly cookie (Set-Cookie) | No | Get new accessToken silently |

## Login

```http
POST /auth/login
Content-Type: application/json

{"email":"admin@example.com","password":"password"}

Response:
{
  "status":"success","code":200,
  "data": {
    "accessToken": "eyJ...",
    "user": {...}
  }
}
Set-Cookie: refreshToken=abc123; HttpOnly; Path=/auth; Max-Age=604800
```

## Refresh

When accessToken expires (401), call refresh:

```http
POST /auth/refresh
Cookie: refreshToken=abc123

Response:
{
  "status":"success","code":200,
  "data": {
    "accessToken": "eyJ...new...",
    "user": {...}
  }
}
Set-Cookie: refreshToken=xyz789; HttpOnly; Path=/auth; Max-Age=604800
```

## Logout

```http
POST /auth/logout
Cookie: refreshToken=abc123

Response:
{"status":"success","code":200,"message":"Logged out","data":null}
Set-Cookie: refreshToken=; Max-Age=-1
```

## Frontend Requirements

- Use `credentials: "include"` in fetch/axios to send/receive cookies
- Store accessToken in memory (not localStorage for refresh flow)
- On 401, call `/auth/refresh` before retrying the original request
- FE cannot read refreshToken from cookie (HttpOnly)
