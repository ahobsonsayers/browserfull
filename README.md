# Browserful

A BAAS (browser-as-a-service) server built on [agent-browser](https://agent-browser.dev/) that spins up browsers on demand and exposes their CDP (Chrome DevTools Protocol) WebSocket endpoint.

Connect tools like [browser-use](https://github.com/browser-use/browser-use), [agent-browser](https://agent-browser.dev/), Puppeteer, Playwright, or any CDP client to a browser started via this API

## Features

- **On-demand browser sessions** — Launch a browser with a single HTTP call, get back a live CDP WebSocket.
- **Named sessions** — Run multiple isolated browser sessions side-by-side, each addressable by name.
- **CDP proxying** — Browserful upgrades your request to a WebSocket and proxies it straight to the browser's CDP endpoint. Standard CDP clients just work.

## Quick start

### Prerequisites

- Go 1.24.3+
- [`agent-browser`](https://agent-browser.dev/) on your `$PATH`

### Run locally

```bash
go run .
```

Server starts on `0.0.0.0:8080`.

### Run with Docker

```bash
docker build -t browserful .
docker run --rm -p 8080:8080 browserful
```

## Configuration

Configured via environment variables:

| Variable                     | Default             | Description                                                              |
| ---------------------------- | ------------------- | ------------------------------------------------------------------------ |
| `BROWSERFUL_PORT`            | `8080`              | HTTP listen port                                                         |
| `BROWSERFUL_DATA_DIR`        | `$HOME/.browserful` | Session metadata + agent-browser config dir                              |
| `BROWSERFUL_ALLOWED_ORIGINS` | _none_              | Comma-separated allowed WebSocket origin hostnames; `0.0.0.0` allows all |

## API

### `POST /sessions`

Launch the default browser session. The response is a WebSocket upgrade (HTTP `101`) proxied to the session's CDP endpoint.

### `POST /sessions/{sessionName}`

Launch a named browser session. `sessionName` must match `^[a-zA-Z0-9_-]+$`.

### `DELETE /sessions/{sessionName}`

Close a named browser session. Returns `204` on success.

### `GET /health`

Health check. Returns `200` if the server is up.

## Using the CDP endpoint

Once you `POST` to `/sessions` (or `/sessions/{name}`), the connection upgrades to a WebSocket and is transparently proxied to the underlying browser's CDP URL. Point your CDP client at the Browserful URL you called — no separate CDP discovery step needed.

```python
# Example with browser-use
from browser_use import BrowserUse

bu = BrowserUse(ws_endpoint="ws://localhost:8080/sessions/my-session")
# Browserful launches the browser and hands you its CDP stream
```

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

## How it works

```
Your CDP client ──WS──▶ Browserful ──WS──▶ agent-browser ──▶ Chrome/Chromium
                     (HTTP server)         (manages browser)   (CDP target)
```

Browserful is a thin HTTP + WebSocket layer over `agent-browser`. When you hit `/sessions`, it asks `agent-browser` to launch a browser, then proxies your WebSocket connection straight through to that browser's CDP endpoint. Closing a session calls `agent-browser close` under the hood.

## Acknowledgements

This project is a thin server on top of [agent-browser](https://agent-browser.dev/) — give it a star if you find Browserful useful.

## License

See [LICENSE](LICENSE).
