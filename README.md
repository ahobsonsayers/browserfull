# Browserfull

A BAAS (browser-as-a-service) server built on [agent-browser](https://agent-browser.dev/) that spins up browsers on demand and exposes their CDP (Chrome DevTools Protocol) WebSocket endpoint.

Connect tools like [browser-use](https://github.com/browser-use/browser-use), [agent-browser](https://agent-browser.dev/), Puppeteer, Playwright, or any CDP client to a browser started via this API — yes, even agent-browser itself can connect back to a session Browserfull launched.

## Quick start

### Run with Docker

The Docker image is self-contained — it bundles `agent-browser`, [cloakbrowser](https://github.com/CloakHQ/CloakBrowser) (a stealth-patched Chromium), and all system libraries. No host dependencies beyond Docker.

```bash
docker build -t browserfull .
docker run --rm -p 8080:8080 browserfull
```

agent-browser uses the bundled cloakbrowser chromium as its default browser. To use a different browser, override `BROWSERFULL_BROWSER_EXECUTABLE_PATH` at `docker run` time.

### Run locally

Prerequisites: Go 1.24.3+ and [`agent-browser`](https://agent-browser.dev/) on your `$PATH`.

```bash
go run .
```

Server starts on `0.0.0.0:8080`. The agent-browser dashboard is started automatically and served at the url.

## Connect to a session

Point any CDP-compatible client at a Browserfull `/connect` URL. The connection upgrades to a WebSocket and is transparently proxied to the underlying browser's CDP endpoint — no separate launch or CDP-discovery step needed.

```python
# Example with browser-use
from browser_use import BrowserUse

bu = BrowserUse(ws_endpoint="ws://localhost:8080/connect/my-session")
# Browserfull launches the browser and hands you its CDP stream
```

```bash
# Example with agent-browser CLI
agent-browser connect "ws://localhost:8080/connect/my-session"
agent-browser snapshot
```

## API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/connect` | Connect to a randomly-named session's CDP WebSocket (launches if not running) |
| `GET` | `/connect/{sessionName}` | Connect to a named session's CDP WebSocket (launches if not running) |
| `DELETE` | `/sessions/{sessionName}` | Close a session |
| `GET` | `/health` | Health check |

`sessionName` must match `^[a-zA-Z0-9_-]+$`.

## Configuration

Configured via environment variables:

| Variable                              | Default              | Description                                                                      |
| ------------------------------------- | -------------------- | -------------------------------------------------------------------------------- |
| `BROWSERFULL_PORT`                    | `8080`               | HTTP listen port                                                                 |
| `BROWSERFULL_DATA_DIR`                | `$HOME/.browserfull` | Session metadata + agent-browser config dir                                      |
| `BROWSERFULL_ALLOWED_ORIGINS`         | _none_               | Comma-separated allowed WebSocket origin hostnames; `*` allows all               |
| `BROWSERFULL_BROWSER_EXECUTABLE_PATH` | _none_               | Browser executable path passed to agent-browser; overrides Chrome auto-discovery |

## How it works

```
Your CDP client ──WS──▶ Browserfull ──WS──▶ agent-browser ──▶ Chrome/Chromium
                     (HTTP server)         (manages browser)   (CDP target)
```

Browserfull is a thin HTTP + WebSocket layer over `agent-browser`. When you hit `/connect/{name}`, it asks `agent-browser` to launch a browser, then proxies your WebSocket connection straight through to that browser's CDP endpoint. Closing a session calls `agent-browser close` under the hood.

## Development

Built with Go + [chi](https://github.com/go-chi/chi) + [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen). The HTTP API is defined in `openapi.yaml` and generated into `api/api.gen.go`.

Common tasks (see `Taskfile.yaml`):

```bash
task run          # run the server
task test         # run tests
task lint         # golangci-lint --fix
task format       # gofumpt
task generate     # regenerate api/api.gen.go from openapi.yaml
task build        # build binary to ./bin/
task build:docker # build docker image
```

## Acknowledgements

This project is a thin server on top of [agent-browser](https://agent-browser.dev/) — give it a star if you find Browserfull useful.

## License

See [LICENSE](LICENSE).
