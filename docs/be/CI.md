# CI Pipeline

## Jobs

| Job | Purpose |
|-----|---------|
| **go-test** | gofmt check + go test ./... |
| **docker-build** | Verify Docker images build |
| **test-ui-build** | Build React test UI |
| **openapi-validate** | Validate OpenAPI spec |
| **secret-scan** | Prevent accidental secret commits |

## Local

```bash
make ci-go        # Go tests
make ci-docker    # Docker build
make ci-test-ui   # UI build
make ci-openapi   # OpenAPI lint
make ci           # All CI checks
```
