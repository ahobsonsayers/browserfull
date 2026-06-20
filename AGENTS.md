# Browserfull — Agent Guidance

A Go HTTP server that proxies to browser CDP (Chrome DevTools Protocol) sessions via WebSocket.

## Documentation Maintenance

**For any and all code/config changes, you MUST review both `README.md` and `AGENTS.md` to determine whether they need updates, and make those updates if so.** This includes new features, refactors, behavior changes, new dependencies, endpoint/CLI/config changes, architecture shifts, workflow additions, or anything else that affects how the project is built, run, tested, configured, or understood. Treat this as a mandatory final step of every change.

## Developer Workflow (Taskfile first)

**Always use the Taskfile (`task <task>`) for common operations** like lint, test, format, build, generate. Do not run ad-hoc `go`/`golangci-lint`/`gofumpt`/`oapi-codegen` commands directly — go through the defined task. If no task exists for the specific thing being done, run the command directly and consider adding a task for it.

Available tasks:

- `task run` — Run the server locally (`go run .`)
- `task test` — Run `go test ./...`
- `task test:clean` — Clear test cache then run tests
- `task lint` — Run `golangci-lint run --fix ./...`
- `task format` — Run `gofumpt -l -w .`
- `task generate` (`task gen`) — Regenerate `api/api.gen.go` from `openapi.yaml` via `oapi-codegen`
- `task build` — Build binary to `./bin/`
- `task build:docker` / `task run:docker` — Docker workflow

## Tooling

- **Go 1.24.3** — Uses `tool` directive in `go.mod`; tools (`golangci-lint`, `gofumpt`, `oapi-codegen`) are installed via `go get -tool`, not globally. Always invoke them via `go tool <tool>` (the Taskfile does this).
- **Linter**: `golangci-lint` v2 config (`.golangci.yml`). Enabled linters: `errcheck`, `gocritic` (`enable-all` with few disabled checks), `makezero`, `misspell`, `nolintlint`, `revive` (`enable-all-rules` with specific disabled rules), `testifylint`, `unparam`, `usestdlibvars`.
- **Formatter**: `gofumpt` with `extra-rules: true` (runs as golangci-lint formatter and as `task format`).
- **Line length limit**: 120 chars (revive `line-length-limit`).
- **Max function statements**: 50 (revive `function-length`, statements only).
- **Exported naming**: revive `exported` is enabled only for types/consts/vars/methods — function names must not stutter (e.g. `proxy.CDP`, not `proxy.ProxyCDP`).
- **Error handling**: Never use inline `if err := ...; err != nil` — always assign the error to a variable first, then check it on the next line.

## Architecture

```
main.go               — Entrypoint: loads config, builds chi router with OpenAPI validation + logger middleware, serves on 0.0.0.0:<port>
internal/config        — Env-config (port, data dir, allowed origins); validates fields
internal/agentbrowser  — Launches browser sessions via `agent-browser` CLI, reads session metadata from <DataDir>/
internal/proxy         — WebSocket reverse proxy to a CDP URL (gorilla/websocket + koding/websocketproxy); origin checking against AllowedOrigins
api/                   — oapi-codegen generated HTTP server (chi) + hand-written handlers
api/api.gen.go         — GENERATED. Do not edit. Regenerate with `task generate`.
api/server.go          — Strict server + ServerOverrides wrapper holding agentBrowser + allowedOrigins
api/connect.go         — Connect (launch + CDP proxy) and close session handlers; default session uses haikunator for random names
api/health.go          — Health check handler
api/middleware/        — OpenAPI request validation + structured request logging (httplog, includes recoverer)
```

### OpenAPI / oapi-codegen

- `openapi.yaml` is the source of truth for the HTTP API.
- `oapi.config.yaml` configures oapi-codegen: `chi-server`, `strict-server`, `embedded-spec`, `models`.
- Generated output: `api/api.gen.go`. Regenerate after editing `openapi.yaml`: `task generate`.
- `api/api.gen.go` embeds the OpenAPI spec and exposes `api.GetSpec()` (returns `*openapi3.T`). **Use `GetSpec()`, not the deprecated `GetSwagger()`** (the latter is retained only for backwards compatibility).
- Wiring pattern (see `main.go`): `api.NewServer(ab, cfg)` returns a `ServerInterface`; pass it to `api.HandlerFromMux(server, router)`.
- **Endpoint structure**:
  - `GET /connect` — connect to a randomly-named session's CDP WebSocket (launches if not running); name generated via haikunator (e.g. `wispy-dust-1337`)
  - `GET /connect/{sessionName}` — connect to named session's CDP WebSocket (launches if not running)
  - `DELETE /sessions/{sessionName}` — close a session
  - `GET /health` — health check
- **WebSocket endpoints must be GET.** Per RFC 6455, WebSocket upgrades are always HTTP GET requests. The `/connect` endpoints must be declared as `get` (not `post`) in `openapi.yaml`, otherwise the chi router returns `405 Method Not Allowed` and the WebSocket upgrade never reaches the handler.

### `internal/agentbrowser`

- Shells out to the `agent-browser` CLI (`exec.LookPath("agent-browser")`).
- `runCmd` is the central exec helper — ensures config file exists, sets `AGENT_BROWSER_CONFIG` + `AGENT_BROWSER_SOCKET_DIR` env vars, returns `cmd.CombinedOutput()`.
- `ensureConfigFile` creates `<DataDir>/config.json` with default content if missing (checks `os.IsNotExist`, returns other stat errors).
- Session metadata (PID, engine, stream port, version) is read from `<DataDir>/<session>.{pid,engine,stream,version}` files — not from CLI JSON output.
- `getCDPURL` parses `data.cdpUrl` from CLI JSON output using `gjson`.
- `ListSessions` parses `data.sessions` array from CLI JSON output using `gjson`.
- Uses `gjson` (not `encoding/json`) for all JSON parsing — check `gjson.Result.Type` / `IsArray()` rather than unmarshaling into structs.
- Errors from `runCmd` are non-zero exit codes; `success` field in JSON output is not checked (redundant with exit code).
- `StartDashboard`/`StopDashboard` (`dashboard.go`) wrap `agent-browser dashboard start [--port <n>]` / `dashboard stop`. Default port is 4848 when `port` arg is 0.

### `internal/proxy`

- `proxy.CDP(w, r, cdpURL, allowedOrigins)` upgrades the inbound request to a WebSocket and proxies it to the CDP URL.
- Origin checker: if `allowedOrigins` contains `"*"`, all origins are accepted; otherwise the Origin header host must match the request host or be in `allowedOrigins`. Missing Origin header is allowed.

## External Dependencies

- **`agent-browser` CLI** — The `internal/agentbrowser` package shells out to `agent-browser open`, `agent-browser get cdp-url`, `agent-browser session list`, `agent-browser close`. The server (`main.go`) fails to start without this binary on `$PATH` (`exec.LookPath`). Integration tests also require it.
- **Docs**: https://agent-browser.dev/ (Commands: https://agent-browser.dev/commands, Configuration: https://agent-browser.dev/configuration, Sessions: https://agent-browser.dev/sessions)
- **Docker image** — The Dockerfile bundles `agent-browser` (v0.28.0, installed via npm builder stage) and cloakbrowser (stealth-patched Chromium from `cloakhq/cloakbrowser:latest`). `BROWSERFULL_BROWSER_EXECUTABLE_PATH=/opt/cloakbrowser/chrome` wires agent-browser to use the bundled cloakbrowser as its default browser. No host `agent-browser` or Chrome install is needed when running via Docker. Override `BROWSERFULL_BROWSER_EXECUTABLE_PATH` to use a different browser.

## Configuration

Loaded from environment variables via `go-envconfig` (`internal/config/config.go`):

- `BROWSERFULL_PORT` — default `8080`
- `BROWSERFULL_DATA_DIR` — default `$HOME/.browserfull`; sets `AGENT_BROWSER_SOCKET_DIR` (session metadata files) and `AGENT_BROWSER_CONFIG` (`<DataDir>/config.json`). See https://agent-browser.dev/configuration.
- `BROWSERFULL_ALLOWED_ORIGINS` — comma-separated list of allowed WebSocket origin hostnames; `*` disables origin checking.
- `BROWSERFULL_BROWSER_EXECUTABLE_PATH` — optional; sets `AGENT_BROWSER_EXECUTABLE_PATH` to point agent-browser at a custom browser binary (e.g. the bundled cloakbrowser in Docker).
- `BROWSERFULL_DASHBOARD_PORT` — optional; no default (0 lets agent-browser use its own default, 4848).
- `go-envconfig` runs default values through `os.Expand`, so `$HOME` in the `default=` tag works.

## Testing

- Use `task test` to run. `task test:clean` clears cache first.
- Use [testify](https://github.com/stretchr/testify) for all tests.
- **Always use `require` (not `assert`) for error checks** — `require.NoError`, `require.Error`, etc. Failures should halt the test immediately. `assert` is fine for non-critical value checks.
- `internal/agentbrowser` has an **integration test** (`TestAgentBrowserLaunch`) that launches a real `agent-browser` session. It has no skip guard — it runs on every `task test` and fails if `agent-browser` is not on `$PATH`.
- **Unix socket path limit**: macOS enforces 103 chars max. The integration test uses `os.MkdirTemp("", "")` (short path) and `time.Now().UnixNano()` as session name to stay under the limit. Do not use `t.TempDir()` or UUID-based session names in that test — the paths exceed 103 chars and `agent-browser` exits 1.
- CI runs `task test` on `ubuntu-latest` (see `.github/workflows/lint-test.yaml`), where `agent-browser` is installed via `npm install -g agent-browser` (Node LTS setup) before running tests.

## CI / Release

- **PRs** (`.github/workflows/pull-request.yaml`): lint (golangci-lint-action) + test (`task test`) via reusable `lint-test.yaml`.
- **Main** (`.github/workflows/main.yaml`): lint-test → build and push Docker image (`arranhs/browserfull`) tagged `develop` + commit SHA.
- **Releases** (`.github/workflows/release.yaml`): manual workflow dispatch with semver bump; creates annotated git tag, GitHub release, then builds/pushes Docker image tagged `latest` + version.
