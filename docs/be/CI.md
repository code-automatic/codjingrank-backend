# Local verification

- `make ci-go`: format check and Go tests.
- `make ci-docker`: build retained service images.
- `make ci-openapi`: existing optional OpenAPI lint (currently masks failures).
- `make ci`: the above commands.

The legacy Vite UI build target was removed with `tools/test_ui`.
No GitHub Actions workflow is currently checked in; local targets do not prove
an automated pipeline exists. Go tests do not prove new-schema integration.
