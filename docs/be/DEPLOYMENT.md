# Deployment status

Updated 06/11/2026 after legacy cleanup.

The retained Compose stack has Gateway, Auth, Run, Submission API and worker,
PostgreSQL, Redis and Piston. Repository/GitHub credentials, mounts, callbacks
and service dependencies have been removed.

## Local checks

```bash
docker compose config --quiet
make ci-go
```

`make compose-up` starts infrastructure; `make compose-full-up` builds and starts
the retained services. Neither command bootstraps or integrates the new database.
Go persistence and DTO integration with UUIDs/new columns is still pending.
Do not deploy this state as a working new-product release.

## Database and target services

`schema.sql` is a destructive bootstrap for a fresh standalone database.
No old migration or seed step remains. Use a dedicated database/container and
never reuse another project's infrastructure for validation.

Only Gateway should be public in the target deployment. AI and Forum services
have not been implemented. Consult [frontend contract](../FRONTEND_API_CONTRACT.md)
for target `/api/v1`; the current Gateway still registers root paths.
