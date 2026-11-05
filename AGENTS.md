# AGENTS.md — innogen-backend

## Project overview

RinnoGen backend — a LeetCode-like platform for curriculum-based programming practice. Go monorepo with microservice-style separation, one root `go.mod`.

## Architecture

- **One root `go.mod`** — module `innogen-backend`, Go 1.26.2
- **`shared/`** — packages imported by all services: `config`, `logger`, `database`, `response`, `errors`, `models`, `middleware`
- **Services** (each has `cmd/main.go` as entrypoint):
  - `api_gateway/` — API Gateway (port 8080)
  - `auth_service/` — Authentication (port 8081)
  - `run_service/` — Code execution via Piston (port 8082)
  - `submission_service/` — Submission + judging (API port 8083, worker has no HTTP)
  - `repo_service/` — GitHub repository management (port 8084)
- **`schema.sql`** — Full PostgreSQL schema (PostgreSQL, pgcrypto). This is the source of truth for the data model.

## Key commands

```bash
# Run individual services
make run-gateway       # API Gateway on :8080
make run-auth          # Auth on :8081
make run-runner        # Run service on :8082
make run-submission    # Submission API on :8083
make run-submission-worker  # Submission worker (no HTTP)
make run-repo          # Repo service on :8084

# Development
make tidy              # go mod tidy
make fmt               # gofmt -s -w .
make test              # go test ./...
```

## Dependencies

- **Database**: `github.com/jackc/pgx/v5/pgxpool` (PostgreSQL connection pool)
- **No web framework** — use `net/http` with Go 1.22+ `ServeMux` pattern routing (`"GET /path"`)
- **Logging**: stdlib `log/slog` with JSON handler
- **JSON**: `encoding/json` only (no third-party JSON libraries)

## Conventions

- All log messages and error messages **must be in English**
- Go variables use **camelCase**
- Run `gofmt -s -w .` before committing
- Config is loaded from environment variables via `shared/config.Load()`
- Default config values are in `shared/config/config.go`
- Infrastructure dependencies (Postgres, Redis) are in `docker-compose.yml`

## Database

- PostgreSQL 16
- Database: `innogen`, User: `innogen`, Password: `innogen`
- Schema file: `schema.sql` — includes tables, triggers, indexes, seed data
- Key triggers to be aware of:
  - **anti-spam**: 10-second cooldown between submissions per user/problem (DB-level trigger)
  - **acceptance rate**: auto-updated on submission insert/status change

## External services

- **Piston** — code execution engine, default URL `http://localhost:2000`
- **GitHub App** — for repository management (integration not yet implemented)

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **innogen-backend-microservice** (2099 symbols, 6525 relationships, 177 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

> If any GitNexus tool warns the index is stale, run `npx gitnexus analyze` in terminal first.

## Always Do

- **MUST run impact analysis before editing any symbol.** Before modifying a function, class, or method, run `gitnexus_impact({target: "symbolName", direction: "upstream"})` and report the blast radius (direct callers, affected processes, risk level) to the user.
- **MUST run `gitnexus_detect_changes()` before committing** to verify your changes only affect expected symbols and execution flows.
- **MUST warn the user** if impact analysis returns HIGH or CRITICAL risk before proceeding with edits.
- When exploring unfamiliar code, use `gitnexus_query({query: "concept"})` to find execution flows instead of grepping. It returns process-grouped results ranked by relevance.
- When you need full context on a specific symbol — callers, callees, which execution flows it participates in — use `gitnexus_context({name: "symbolName"})`.

## Never Do

- NEVER edit a function, class, or method without first running `gitnexus_impact` on it.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis.
- NEVER rename symbols with find-and-replace — use `gitnexus_rename` which understands the call graph.
- NEVER commit changes without running `gitnexus_detect_changes()` to check affected scope.

## Resources

| Resource | Use for |
|----------|---------|
| `gitnexus://repo/innogen-backend-microservice/context` | Codebase overview, check index freshness |
| `gitnexus://repo/innogen-backend-microservice/clusters` | All functional areas |
| `gitnexus://repo/innogen-backend-microservice/processes` | All execution flows |
| `gitnexus://repo/innogen-backend-microservice/process/{name}` | Step-by-step execution trace |

## GitNexus Usage for OpenCode

Before editing unfamiliar code:

- Use GitNexus MCP to query related symbols, callers, callees, and execution flows.
- Run impact analysis before changing exported functions, handlers, service clients, middleware, DB access functions, or queue logic.
- For cross-service bugs, trace the full request lifecycle first.
- After edits, check affected symbols and flows with GitNexus before committing.

Do not blindly grep and patch if GitNexus can provide call graph or impact context.

## Code Search

Use `semble search` to find code by describing what it does or naming a symbol/identifier, instead of grep:

​```bash
semble search "authentication flow" ./my-project
semble search "save_pretrained" ./my-project
semble search "save model to disk" ./my-project --top-k 10
​```

If you anticipate doing more than one search, use `semble index` to create an index.

​```bash
semble index ./my-project -o my_index
​```

You can then reuse this index later on:

​```bash
semble search "save_pretrained" --index my_index
​```

An index is not automatically updated, so if the code changes significantly, reindex. If you notice stale results while resolving searches to files, reindex.

Use `--content docs` to search documentation and prose, `--content config` for config files (yaml, toml, etc.), or `--content all` to search code, docs, and config:

​```bash
semble search "deployment guide" ./my-project --content docs
semble search "database host port" ./my-project --content config
semble search "authentication" ./my-project --content all
​```

Use `semble find-related` to discover code similar to a known location (pass `file_path` and `line` from a prior search result):

​```bash
semble find-related src/auth.py 42 ./my-project
​```

Like search, `find-related` also accepts an `--index` argument.

`path` defaults to the current directory when omitted; git URLs are accepted.

If `semble` is not on `$PATH`, use `uvx --from "semble[mcp]" semble` in its place.

### Workflow

1. Index the repo using `semble index -o cached_index`.
2. Start with `semble search` to find relevant chunks. Pass the index to achieve results faster.
3. Use `--content docs` for documentation, `--content config` for config files, or `--content all` for everything.
4. Inspect full files only when the returned chunk does not give enough context.
5. Optionally use `semble find-related` with a promising result's `file_path` and `line` to discover related implementations.
6. Use grep only when you need exhaustive literal matches or quick confirmation of an exact string.

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **codjingrank** (1603 symbols, 3860 relationships, 132 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

> If any GitNexus tool warns the index is stale, run `npx gitnexus analyze` in terminal first.

## Always Do

- **MUST run impact analysis before editing any symbol.** Before modifying a function, class, or method, run `gitnexus_impact({target: "symbolName", direction: "upstream"})` and report the blast radius (direct callers, affected processes, risk level) to the user.
- **MUST run `gitnexus_detect_changes()` before committing** to verify your changes only affect expected symbols and execution flows.
- **MUST warn the user** if impact analysis returns HIGH or CRITICAL risk before proceeding with edits.
- When exploring unfamiliar code, use `gitnexus_query({query: "concept"})` to find execution flows instead of grepping. It returns process-grouped results ranked by relevance.
- When you need full context on a specific symbol — callers, callees, which execution flows it participates in — use `gitnexus_context({name: "symbolName"})`.

## When Debugging

1. `gitnexus_query({query: "<error or symptom>"})` — find execution flows related to the issue
2. `gitnexus_context({name: "<suspect function>"})` — see all callers, callees, and process participation
3. `READ gitnexus://repo/codjingrank/process/{processName}` — trace the full execution flow step by step
4. For regressions: `gitnexus_detect_changes({scope: "compare", base_ref: "main"})` — see what your branch changed

## When Refactoring

- **Renaming**: MUST use `gitnexus_rename({symbol_name: "old", new_name: "new", dry_run: true})` first. Review the preview — graph edits are safe, text_search edits need manual review. Then run with `dry_run: false`.
- **Extracting/Splitting**: MUST run `gitnexus_context({name: "target"})` to see all incoming/outgoing refs, then `gitnexus_impact({target: "target", direction: "upstream"})` to find all external callers before moving code.
- After any refactor: run `gitnexus_detect_changes({scope: "all"})` to verify only expected files changed.

## Never Do

- NEVER edit a function, class, or method without first running `gitnexus_impact` on it.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis.
- NEVER rename symbols with find-and-replace — use `gitnexus_rename` which understands the call graph.
- NEVER commit changes without running `gitnexus_detect_changes()` to check affected scope.

## Tools Quick Reference

| Tool | When to use | Command |
|------|-------------|---------|
| `query` | Find code by concept | `gitnexus_query({query: "auth validation"})` |
| `context` | 360-degree view of one symbol | `gitnexus_context({name: "validateUser"})` |
| `impact` | Blast radius before editing | `gitnexus_impact({target: "X", direction: "upstream"})` |
| `detect_changes` | Pre-commit scope check | `gitnexus_detect_changes({scope: "staged"})` |
| `rename` | Safe multi-file rename | `gitnexus_rename({symbol_name: "old", new_name: "new", dry_run: true})` |
| `cypher` | Custom graph queries | `gitnexus_cypher({query: "MATCH ..."})` |

## Impact Risk Levels

| Depth | Meaning | Action |
|-------|---------|--------|
| d=1 | WILL BREAK — direct callers/importers | MUST update these |
| d=2 | LIKELY AFFECTED — indirect deps | Should test |
| d=3 | MAY NEED TESTING — transitive | Test if critical path |

## Resources

| Resource | Use for |
|----------|---------|
| `gitnexus://repo/codjingrank/context` | Codebase overview, check index freshness |
| `gitnexus://repo/codjingrank/clusters` | All functional areas |
| `gitnexus://repo/codjingrank/processes` | All execution flows |
| `gitnexus://repo/codjingrank/process/{name}` | Step-by-step execution trace |

## Self-Check Before Finishing

Before completing any code modification task, verify:
1. `gitnexus_impact` was run for all modified symbols
2. No HIGH/CRITICAL risk warnings were ignored
3. `gitnexus_detect_changes()` confirms changes match expected scope
4. All d=1 (WILL BREAK) dependents were updated

## Keeping the Index Fresh

After committing code changes, the GitNexus index becomes stale. Re-run analyze to update it:

```bash
npx gitnexus analyze
```

If the index previously included embeddings, preserve them by adding `--embeddings`:

```bash
npx gitnexus analyze --embeddings
```

To check whether embeddings exist, inspect `.gitnexus/meta.json` — the `stats.embeddings` field shows the count (0 means no embeddings). **Running analyze without `--embeddings` will delete any previously generated embeddings.**

> Claude Code users: A PostToolUse hook handles this automatically after `git commit` and `git merge`.

## CLI

| Task | Read this skill file |
|------|---------------------|
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus/gitnexus-cli/SKILL.md` |

<!-- gitnexus:end -->
