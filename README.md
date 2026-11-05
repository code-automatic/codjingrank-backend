# RinnoGen Backend

A LeetCode-like curriculum-based programming practice platform.

## Architecture

See [docs/be/ARCHITECTURE.md](docs/be/ARCHITECTURE.md) for full architecture documentation.

## CI/CD

GitHub Actions validates on every push/PR: Go tests, Docker build, UI build, OpenAPI lint, secret scan.

```bash
make ci    # Run all CI checks locally
```

## Docker

```bash
make compose-full-up    # Build and start all services
make compose-full-down  # Stop all services
make compose-full-logs  # View all logs
```

## Production Deployment

See [docs/be/VPS_NGINX_DEPLOYMENT.md](docs/be/VPS_NGINX_DEPLOYMENT.md) for full VPS setup with Nginx + HTTPS.

## API Documentation
- Frontend: <http://localhost:8080/docs/fe>
- Full: <http://localhost:8080/docs/be>

## Quick Start

```bash
# 1. Start infrastructure
make compose-up

# 2. Apply schema and seed data
make seed-dev

# 3. Start all services (each in a separate terminal)
make run-gateway &
make run-auth &
make run-runner &
make run-submission &
make run-submission-worker &
make run-repo &

# 4. Run E2E validation
make e2e
```

## Development

```bash
make fmt          # Format all Go files
make tidy         # Clean up dependencies
make test         # Run all tests
```

## Local Test UI

A lightweight React UI for manual backend testing:

```bash
make test-ui-install   # Install dependencies
make test-ui           # Start dev server (http://localhost:5173)
make test-ui-build     # Production build
```

Features: Dashboard, Auth, Curriculum browser, Runner, Submissions, Admin CRUD, GitHub webhook notes.

## API Endpoints (via API Gateway :8080)

| Method | Path | Auth |
|--------|------|------|
| POST | /auth/login | No |
| GET | /auth/me | Bearer |
| GET | /subjects | No |
| POST | /run | Bearer |
| POST | /submit | Bearer |
| GET | /submissions/{id} | Bearer |
| GET | /me/submissions | Bearer |

## Environment

Copy `.env.example` to `.env` and adjust as needed:

```bash
cp .env.example .env
```
