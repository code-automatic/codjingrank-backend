# Deployment Guide

> **v2.0 target (05/11/2026):** Run/Submit/Judge/Redis/Piston remain. The
> repository/GitHub deployment below is legacy and must not be added to a new
> environment. Forum and AI services are planned until their implementation and
> health checks land.

## Docker Architecture

| Service | Port | Depends On |
|---------|------|-----------|
| api_gateway | 8080 | auth, run, submission |
| auth_service | 8081 | postgres |
| run_service | 8082 | postgres, piston |
| submission_service | 8083 | postgres, redis |
| submission_worker | — | postgres, redis, piston, ai_service (target) |
| forum_service (target) | TBD | postgres |
| ai_service (target) | internal | provider + postgres |

## Quick Start (Docker)

```bash
# Build and start all services
make compose-full-up

# Apply schema + seed
make seed-dev

# Check health
curl http://localhost:8080/health

# View logs
make compose-full-logs

# Stop everything
make compose-full-down

# Reset volumes
make docker-reset
```

## Service URLs (Docker)

Internal: `http://<service_name>:<port>`
Host: `http://localhost:<port>`

## Legacy GitHub configuration (do not use for v2.0)

Mount private key:
```yaml
volumes:
  - ./secrets:/app/secrets:ro
```
Set: `GITHUB_PRIVATE_KEY_PATH=/app/secrets/github-key.pem`
