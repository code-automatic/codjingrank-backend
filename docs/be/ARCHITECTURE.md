# RinnoGen Architecture

> **Cleanup 06/11/2026:** Repository/GitHub and Subject/Session/Lesson code,
> old migrations/seeds and the Vite test UI have been removed on
> `chore/remove-legacy-code`. Any inventory of those implementations below is
> historical. Retained Go persistence/DTOs still require UUID/schema integration;
> `/api/v1`, Forum and AI remain target-only. See [project state](../../PROJECT_STATE.md).


> **v2.0 target note (05/11/2026):** This document records the current Go
> services and the approved migration boundary. Run/Submit/Judge/Piston/Redis
> remain; Repository/GitHub/commit flow is legacy and will be removed. Forum and
> AI review are planned boundaries, not live services. AI starts only after an
> Accepted verdict and never changes ranking.

## Overview

RinnoGen is a LeetCode-style curriculum-based programming practice platform.
Go monorepo with microservice-style separation, one root `go.mod`.

**Codebase stats (GitNexus):** 132 files, 2,099 symbols, 6,525 relationships, 46 communities, 177 execution flows.

## Architecture Diagram

```mermaid
graph TB
    subgraph "Frontend"
        FE[Test UI :5173]
    end

    subgraph "API Gateway :8080"
        GW[api_gateway]
        MW[Middleware Stack]
        PR[Reverse Proxies]
        CR[Curriculum/Problem Handlers]
        AD[Admin CRUD]
    end

    subgraph "Auth :8081"
        AS[auth_service]
        ASH[Login / Register]
        ASM[/auth/me]
        JWT[JWT + bcrypt]
    end

    subgraph "Runner :8082"
        RS[run_service]
        RSJ[Sample Test Judge]
    end

    subgraph "Submissions :8083"
        SS[submission_service API]
        SW[submission_worker]
        RQ[Redis Queue]
    end

    subgraph "Target services"
        FS[forum_service - planned]
        AIS[ai_service - planned]
    end

    subgraph "Infrastructure"
        PG[(PostgreSQL 16)]
        RD[(Redis 7)]
        PT[Piston Engine]
    end

    subgraph "Shared"
        CF[config]
        DB[database/pgxpool]
        LG[logger/slog]
        RS2[response]
        MW2[middleware/Auth]
        MD[models]
        PI[piston client]
        JD[judge]
        CO[constants]
        LU[languageutil]
    end

    FE --> GW
    GW --> MW
    MW --> PR
    MW --> CR
    MW --> AD
    PR --> AS
    PR --> RS
    PR --> SS
    PR --> FS
    PR --> AIS

    AS --> PG
    AS --> JWT
    RS --> PT
    RS --> PG
    SS --> PG
    SS --> RQ
    SW --> RQ
    SW --> PT
    SW --> PG
    AIS --> PG
    FS --> PG

    GW --> CF
    AS --> CF
    RS --> CF
    SS --> CF
```

## Functional Areas (46 communities detected)

| Community | Description |
|-----------|-------------|
| **API Gateway** | Public entrypoint, JWT validation, reverse proxy, CORS, rate limiting |
| **Auth Service** | Login/Register with bcrypt and JWT generation; GitHub account routes are legacy |
| **Run Service** | Code execution via Piston, sample test case judging |
| **Submission Service** | Submission CRUD, Redis queue enqueue, async judging |
| **Submission Worker** | Dequeue, load test cases, execute via Piston, judge, update DB |
| **Forum Service (planned)** | Posts, comments, likes and moderation; separate from judge |
| **AI Service (planned)** | Code-quality review requested internally after Accepted |
| **Repo Service (legacy)** | GitHub App commit/webhook/OAuth code retained only until removal |
| **Admin** | CRUD for languages, subjects, sessions, lessons, problems, test cases, tags |
| **Curriculum** | Public read APIs for subjects, sessions, lessons, problems |
| **Shared** | Config, database, logger, response envelope, auth middleware, models |

## Key Execution Flows

### 1. Auth Login Flow
```
Client → POST /auth/login → api_gateway (strip headers) → auth_service
  → Validate email/password → bcrypt compare → GenerateToken (HS256, 24h)
  → Return { accessToken, user }
```

### 2. Submission → Judge → AI Review (target flow)
```
Client → POST /submit → api_gateway (validate JWT, inject X-User-ID)
  → submission_service → INSERT Pending → LPUSH Redis queue → return 201

Worker → BRPOP Redis → load submission → mark Running
  → load problem + language + ALL test cases
  → for each test case: Piston Execute → Judge Evaluate
  → aggregate: Compilation > Internal > TLE > RTE > WA > Accepted
  → update SubmissionResult

If Accepted:
  → persist verdict/progress
  → enqueue internal AI review with submission ID/source/language/problem context
  → ai_service stores ai_reviews (queued/running/completed/failed)
  → frontend reads review separately; verdict/rank never waits for AI
```

### 3. Legacy GitHub flows (removal scope)
```
GitHub → POST /webhooks/github → api_gateway (no auth, proxy)
  → repo_service → verify X-Hub-Signature-256 (HMAC-SHA256)

installation.created → upsert github_installations → backfill github_accounts
installation.deleted → mark status=deleted, uninstalled_at=now
installation.suspend → mark status=suspended
```

### 4. Legacy OAuth account linking (removal scope)
```
Client → GET /github/oauth/start-url → returns GitHub authorize URL
GitHub → GET /github/oauth/callback?code=&state=
  → validate state JWT → exchange code → GET /user
  → store github_user_id, username, avatar, noreply_email, commit_author_name
  → redirect to frontend ?oauth=connected
```

### 5. Admin Content CRUD
```
Client → POST /admin/languages (JWT + role=admin)
  → api_gateway → StripPrefix("/admin") → sub-mux
  → AdminHandler → AdminRepository → PostgreSQL
  → return 201 with created entity
```

## Service Ports

| Service | Port | Auth |
|---------|------|------|
| api_gateway | 8080 | JWT at edge |
| auth_service | 8081 | X-User-ID (trusted) |
| run_service | 8082 | X-User-ID (trusted) |
| submission_service | 8083 | X-User-ID (trusted) |
| submission_worker | — | Internal only |
| repo_service (legacy, removal) | 8084 | X-User-ID (trusted) |
| forum_service (planned) | TBD | Gateway JWT/service auth |
| ai_service (planned) | Internal only | Service token |

## Shared Packages

| Package | Purpose |
|---------|---------|
| `shared/config` | Environment variable loading |
| `shared/database` | pgxpool PostgreSQL connection |
| `shared/logger` | slog JSON logger |
| `shared/response` | JSON response envelope |
| `shared/middleware` | JWT Auth, X-User-ID extraction |
| `shared/models` | All database model structs |
| `shared/piston` | Piston API v2 HTTP client |
| `shared/judge` | Output comparison with TLE detection |
| `shared/constants` | Status string constants |
| `shared/languageutil` | File name resolution |

## Response Envelope

```json
// Success
{ "status": "success", "code": 200, "message": "OK", "data": {} }

// Error
{ "status": "error", "code": 404, "message": "Not found", "error": "NOT_FOUND", "details": null }
```

## Database

PostgreSQL 16 with pgcrypto. Schema source: `schema.sql`.
Current key tables include users, problems, submissions, repositories and GitHub
tables. The target migration adds Package/Level/progress/leaderboard/forum/
`ai_reviews` and removes repository/GitHub tables after a data-retention decision.

## Infrastructure

Docker Compose currently runs Postgres 16, Redis 7, Piston and the existing Go
services. Forum/AI are planned additions; no deployment claim is made for them yet.
