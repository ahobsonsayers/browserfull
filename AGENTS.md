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
main.go               — Entrypoint: loads config, starts dashboard, builds chi router with API routes (OpenAPI-validated) + dashboard proxy (only on its known paths), serves on 0.0.0.0:<port>
internal/config        — Env-config (port, data dir, allowed origins); validates fields
internal/agentbrowser  — Launches browser sessions via `agent-browser` CLI, reads session metadata from <DataDir>/
internal/proxy         — WebSocket reverse proxy to a CDP URL (gorilla/websocket + koding/websocketproxy); HTTP reverse proxy to the dashboard (httputil)
api/                   — oapi-codegen generated HTTP server (chi) + hand-written handlers
api/api.gen.go         — GENERATED. Do not edit. Regenerate with `task generate`.
api/server.go          — Strict server holding all state (agentBrowser, allowedOrigins) + ServerOverrides wrapper; overrides exist only for the WebSocket connect handlers, which need the raw http.ResponseWriter
api/connect.go         — Connect (launch + CDP proxy) and close session handlers; default session uses haikunator for random names
api/health.go          — Health check handler
api/middleware/        — OpenAPI request validation + structured request logging (httplog, includes recoverer)
```

### Dashboard proxy

The agent-browser dashboard binds to `127.0.0.1` only, so it cannot be exposed directly via Docker port mapping. Browserfull starts the dashboard (via `agent-browser dashboard start`) and reverse-proxies it on the same port as the HTTP API, but only on the dashboard's known paths: `/`, `/_next/*` (Next.js assets), `/api/*` (dashboard API incl. the `wss /api/session/{port}/stream` WebSocket), and `/favicon.ico`. All other paths return chi's default 404.

### OpenAPI / oapi-codegen

- `openapi.yaml` is the source of truth for the HTTP API.
- `oapi.config.yaml` configures oapi-codegen: `chi-server`, `strict-server`, `embedded-spec`, `models`.
- Generated output: `api/api.gen.go`. Regenerate after editing `openapi.yaml`: `task generate`.
- `api/api.gen.go` embeds the OpenAPI spec and exposes `api.GetSpec()` (returns `*openapi3.T`). **Use `GetSpec()`, not the deprecated `GetSwagger()`** (the latter is retained only for backwards compatibility).
- Wiring pattern (see `main.go`): `api.NewServer(ab, cfg)` returns a `ServerInterface`; pass it to `api.HandlerFromMux(server, router)`.
- **Endpoint structure**:
  - `GET /connect` — connect to a randomly-named session's CDP WebSocket (launches if not running); name generated via haikunator (e.g. `wispy-dust-1337`)
  - `GET /connect/{sessionName}` — connect to named session's CDP WebSocket (launches if not running)
  - Both take `?headed=true|false`, overriding `BROWSERFULL_DEFAULT_HEADED` for the launch. The mode is a launch option: a running session keeps the one it started with.
  - Both take `?extensions=name1,name2`, added to `BROWSERFULL_DEFAULT_EXTENSIONS`. Names resolve only against `<DataDir>/extensions/` (see Configuration), so a client cannot point the browser at an arbitrary path. Extensions are also launch-time only.
  - `GET /sessions` — list active session names
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
- `Launch(sessionName string, opts LaunchOptions)` appends `--headed` and one `--extension <path>` per entry in `opts.Extensions` to the `open` command. Extensions must be already-installed extension folders with a `manifest.json`.
- `StartDashboard`/`StopDashboard` (`dashboard.go`) wrap `agent-browser dashboard start` / `dashboard stop`. The dashboard always uses agent-browser's default port (4848) on loopback; it is not configurable because it is reverse-proxied through browserfull at `/`.

### `internal/proxy`

- `proxy.CDP(w, r, cdpURL, allowedOrigins)` upgrades the inbound request to a WebSocket and proxies it to the CDP URL.
- `proxy.Dashboard()` returns an `httputil.ReverseProxy` handler to the dashboard (hardcoded `http://127.0.0.1:4848`). Mounted only on the dashboard's known paths (`/`, `/_next/*`, `/api/*`, `/favicon.ico`) on the main chi router in `main.go`; unknown paths 404.
- Origin checker: if `allowedOrigins` contains `"*"`, all origins are accepted; otherwise the Origin header host must match the request host or be in `allowedOrigins`. Missing Origin header is allowed.

## External Dependencies

- **`agent-browser` CLI** — The `internal/agentbrowser` package shells out to `agent-browser open`, `agent-browser get cdp-url`, `agent-browser session list`, `agent-browser close`. The server (`main.go`) fails to start without this binary on `$PATH` (`exec.LookPath`). Integration tests also require it.
- **Docs**: https://agent-browser.dev/ (Commands: https://agent-browser.dev/commands, Configuration: https://agent-browser.dev/configuration, Sessions: https://agent-browser.dev/sessions)
- **Docker image** — The Dockerfile bundles `agent-browser` (installed via npm builder stage) and clearcote (open-source stealth Chromium from `clearcotelabs/clearcote-browser`). Both are **pinned** by top-of-file ARGs — `ARG AGENT_BROWSER_VERSION=`, `ARG CLEARCOTE_VERSION=` (a `teamflatearth/clearcote` image tag) — which the `update.yaml` workflow bumps daily; never use `latest` or a floating tag for either. `BROWSERFULL_BROWSER_EXECUTABLE_PATH=/opt/clearcote/chrome` wires agent-browser to use the bundled clearcote as its default browser. No host `agent-browser` or Chrome install is needed when running via Docker. Override `BROWSERFULL_BROWSER_EXECUTABLE_PATH` to use a different browser.

  The image runs headed on Xvfb (`docker-entrypoint.sh`, `BROWSERFULL_DEFAULT_HEADED=true`, `DISPLAY=:99`) because headless Chrome is itself a detection signal, and sets `AGENT_BROWSER_ARGS=--test-type` to drop the infobar that `--no-sandbox` raises in a headed window.

  Clearcote was chosen over a closed stealth Chromium because the open build is BSD-3 with every patch public, needs no licence key or licence server, has no cap on concurrent browsers (so one container serves many sessions), and keeps shipping current Chromium releases. The browser is copied out of the pinned Clearcote Docker image, so each release is one version to bump.

## Configuration

Loaded from environment variables via `go-envconfig` (`internal/config/config.go`):

- `BROWSERFULL_PORT` — default `8080`
- `BROWSERFULL_DATA_DIR` — default `$HOME/.browserfull`; sets `AGENT_BROWSER_SOCKET_DIR` (session metadata files) and `AGENT_BROWSER_CONFIG` (`<DataDir>/config.json`). See https://agent-browser.dev/configuration.
- `BROWSERFULL_ALLOWED_ORIGINS` — comma-separated list of allowed WebSocket origin hostnames; `*` disables origin checking.
- `BROWSERFULL_BROWSER_EXECUTABLE_PATH` — optional; sets `AGENT_BROWSER_EXECUTABLE_PATH` to point agent-browser at a custom browser binary (e.g. the bundled clearcote in Docker).
- `BROWSERFULL_DEFAULT_HEADED` — default `false`; launches sessions with a visible window (`--headed`). Needs `DISPLAY`; without one the launch falls back to headless with a log line. The Docker image sets it `true` and runs Xvfb on `:99`.
- `BROWSERFULL_DEFAULT_EXTENSIONS` — comma-separated extension folder names under `<DataDir>/extensions/`, loaded into every session. `Validate()` resolves each name at startup, so a typo stops the server instead of breaking connects.
- `go-envconfig` runs default values through `os.Expand`, so `$HOME` in the `default=` tag works.

## Skills

- `skills/browserfull/SKILL.md` — agent skill for using Browserfull (connect URL, sessions, close). Generic server knowledge sits in the root file; tool-specific attach/close mechanics live in `references/agent-browser.md` and `references/browser-use.md`. It intentionally contains no agent-browser or browser-use command knowledge: it always instructs loading the tool's skill. Tool selection follows what is installed; both or neither defaults to agent-browser. When skills are missing it offers `npx skills add https://github.com/vercel-labs/agent-browser --skill agent-browser` or `uv tool install browser-use && browser-use skill install` after user approval. If you change API endpoints or behavior, update the skill and its references to match.

## Testing

- Use `task test` to run. `task test:clean` clears cache first.
- Use [testify](https://github.com/stretchr/testify) for all tests.
- **Always use `require` (not `assert`) for error checks** — `require.NoError`, `require.Error`, etc. Failures should halt the test immediately. `assert` is fine for non-critical value checks.
- `internal/agentbrowser` has an **integration test** (`TestAgentBrowserLaunch`) that launches a real `agent-browser` session. It has no skip guard — it runs on every `task test` and fails if `agent-browser` is not on `$PATH`.
- **Unix socket path limit**: macOS enforces 103 chars max. The integration test uses `os.MkdirTemp("", "")` (short path) and `time.Now().UnixNano()` as session name to stay under the limit. Do not use `t.TempDir()` or UUID-based session names in that test — the paths exceed 103 chars and `agent-browser` exits 1.
- CI runs `task test` on `ubuntu-latest` (see `.github/workflows/lint-test.yaml`), where `agent-browser` is installed via `npm install -g agent-browser` (Node LTS setup) before running tests.

## CI / Release

Workflows for the Docker image live in this repo; the build and dependency-bump jobs call the reusable workflows in [`ahobsonsayers/github-workflows`](https://github.com/ahobsonsayers/github-workflows) (`docker-build-push.yaml`, `docker-update-app-version.yaml`) at `@main`. They need the repo **variable** `DOCKER_REGISTRY_USERNAME` and **secret** `DOCKER_REGISTRY_PASSWORD` (Docker Hub); callers pass `secrets: inherit`.

- **PRs** (`.github/workflows/pull-request.yaml`): lint (golangci-lint-action) + test (`task test`) via reusable `lint-test.yaml`.
- **Main** (`.github/workflows/main.yaml`): on push to main, and nightly at 03:00 UTC → lint-test → build and push `arranhs/browserfull` tagged `latest`, `develop`, commit SHA, and the combined version tag.
- **Dependency update** (`.github/workflows/update.yaml`): nightly at 02:00 UTC → bumps `ARG AGENT_BROWSER_VERSION` (npm) and `ARG CLEARCOTE_VERSION` (latest GitHub release tag, which Clearcote also publishes as a Docker Hub tag) in the Dockerfile and commits to main. A `GITHUB_TOKEN` push does not trigger `main.yaml`, which is why the build runs an hour later.
- **Releases** (`.github/workflows/release.yaml`): manual workflow dispatch with semver bump; creates annotated git tag, GitHub release, then builds/pushes tagged `latest` + the new version tag.
- **Image tags** (computed in `build-push.yaml`): `latest`, caller-supplied tags, and always `<app-version>-ab<agent-browser>-cc<clearcote>` — e.g. `v0.3.0-ab0.43.0-cc0.1.0-pre.25`. App version is the `app-version` input (release), else the tag on the built commit, else `dev-<short sha>`. Untagged commits never reuse an earlier tag, so a combination tag always names the code it contains.
