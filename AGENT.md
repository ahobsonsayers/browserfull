# Browserful — Agent Guidance

A Go HTTP server that proxies to browser CDP (Chrome DevTools Protocol) sessions via WebSocket.

## Developer Workflow (Taskfile first)

**Always use the Taskfile (`task <task>`) for common operations** like lint, test, format, build, generate, etc. Do not run ad-hoc `go`/`golangci-lint`/`gofumpt` commands directly — go through the defined task instead. The only exception is when no task exists for the specific thing being done; in that case, run the command directly and consider adding a task for it.

Available tasks:

- `task run` — Run the server locally
- `task test` / `task test:clean` — Run tests (none exist currently)
- `task lint` — Run `golangci-lint --fix ./...`
- `task format` — Run `gofumpt -l -w .`
- `task generate` (`task gen`) — Run `go generate ./...`
- `task build` — Build binary to `./bin/`
- `task build:docker` / `task run:docker` — Docker workflow

## Tooling

- **Go 1.24.3** — Uses `tool` directive in `go.mod`; tools (`golangci-lint`, `gofumpt`, `oapi-codegen`) are installed via `go get -tool`, not globally.
- **Linter**: `golangci-lint` v2 config (`.golangci.yml`). Enabled linters: `errcheck`, `gocritic`, `makezero`, `misspell`, `nolintlint`, `revive`, `testifylint`, `unparam`, `usestdlibvars`.
- **Formatter**: `gofumpt` with `extra-rules: true`.

## Architecture

```
main.go                  — Entry point: config → chi router → HTTP server
├── internal/config      — Env-config (port, sessions dir); validates port
├── internal/browser     — Launches browser sessions via `agent-browser` CLI
├── internal/proxy       — WebSocket proxy between client and browser CDP
├── api/handlers         — HTTP handlers (CDP endpoints)
├── api/oapi.gen.go      — Generated from `oapi.yaml` by oapi-codegen
└── oapi.yaml            — OpenAPI 3.0 spec
```

## Code Generation

`api/oapi.gen.go` is generated from `oapi.yaml` by `oapi-codegen/v2`. There are no `go:generate` directives in the source, but `task generate` runs `go generate ./...` if any are added. To regenerate manually:

```bash
go tool oapi-codegen -generate chi-server,strict,types -package api oapi.yaml > api/oapi.gen.go
```

## External Dependencies

- **`agent-browser` CLI** — The `internal/browser` package shells out to `agent-browser open` and `agent-browser get cdp-url`. The server does not run without this binary on `$PATH`.
- **Session names** must match `^[a-zA-Z0-9_-]+$` (enforced by OpenAPI spec, but validate in handlers too).

## Configuration

Loaded from environment variables via `go-envconfig`:

- `BROWSERFUL_PORT` — default `8080`
- `BROWSERFUL_SESSIONS_DIR` — default `data/sessions`

## Testing

No `*_test.go` files currently exist. If you add tests:
- Use `task test` to run.
- `task test:clean` clears the test cache first.
- Use [testify](https://github.com/stretchr/testify) for all tests.
- Always use `require` (e.g. `require.Equal`, `require.NoError`) instead of `assert` — failures should halt the test immediately.
- Prefer focused tests for fast feedback; the repo has no integration test harness or mocks yet.

## CI / Release

- **PRs**: lint + test via `.github/workflows/pull-request.yaml`
- **Main**: lint + test → then build and push Docker image (`arranhs/browserful`)
- **Releases**: manual workflow dispatch with semver bump; tags image as `latest` + version
