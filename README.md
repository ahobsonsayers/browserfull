# Browserful

A BAAS (browser-as-a-service) server built on [agent-browser](https://agent-browser.dev/) that spins up browsers on demand and exposes their CDP (Chrome DevTools Protocol) WebSocket endpoint.

Connect tools like [browser-use](https://github.com/browser-use/browser-use), [agent-browser](https://agent-browser.dev/), Puppeteer, Playwright, or any CDP client to a browser started via this API — yes, even agent-browser itself can connect back to a session Browserful launched.

## Quick start

### Run with Docker

```bash
docker build -t browserful .
docker run --rm -p 8080:8080 browserful
```

### Run locally

Prerequisites: Go 1.24.3+ and [`agent-browser`](https://agent-browser.dev/) on your `$PATH`.

```bash
go run .
```

Server starts on `0.0.0.0:8080`.

## Connect to a session

Point any CDP-compatible client at a Browserful `/connect` URL. The connection upgrades to a WebSocket and is transparently proxied to the underlying browser's CDP endpoint — no separate launch or CDP-discovery step needed.

```python
# Example with browser-use
from browser_use import BrowserUse

bu = BrowserUse(ws_endpoint="ws://localhost:8080/connect/my-session")
# Browserful launches the browser and hands you its CDP stream
```

```bash
# Example with agent-browser CLI
agent-browser connect "ws://localhost:8080/connect/my-session"
agent-browser snapshot
```

## API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/connect` | Connect to the default session's CDP WebSocket (launches if not running) |
| `GET` | `/connect/{sessionName}` | Connect to a named session's CDP WebSocket (launches if not running) |
| `DELETE` | `/sessions/{sessionName}` | Close a session |
| `GET` | `/health` | Health check |

`sessionName` must match `^[a-zA-Z0-9_-]+$`.

## Configuration

Configured via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `BROWSERFUL_PORT` | `8080` | HTTP listen port |
| `BROWSERFUL_DATA_DIR` | `$HOME/.browserful` | Session metadata + agent-browser config dir |
| `BROWSERFUL_ALLOWED_ORIGINS` | _none_ | Comma-separated allowed WebSocket origin hostnames; `0.0.0.0` allows all |

## How it works

```
Your CDP client ──WS──▶ Browserful ──WS──▶ agent-browser ──▶ Chrome/Chromium
                     (HTTP server)         (manages browser)   (CDP target)
```

Browserful is a thin HTTP + WebSocket layer over `agent-browser`. When you hit `/connect/{name}`, it asks `agent-browser` to launch a browser, then proxies your WebSocket connection straight through to that browser's CDP endpoint. Closing a session calls `agent-browser close` under the hood.

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

This project is a thin server on top of [agent-browser](https://agent-browser.dev/) — give it a star if you find Browserful useful.

## License

See [LICENSE](LICENSE).