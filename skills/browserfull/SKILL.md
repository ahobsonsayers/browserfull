---
name: browserfull
description: Use when automating a browser through a Browserfull server — a browser-as-a-service HTTP server that launches browsers on demand and proxies their CDP WebSocket. Triggers on BROWSERFULL_API_URL, a running browserfull instance, /connect WebSocket session URLs, listing or closing browserfull sessions, or the user naming browserfull.
allowed-tools: Bash(curl:*), Bash(printenv:*)
---

# browserfull automation

Browserfull is a BaaS server: one HTTP endpoint that launches browsers (via agent-browser under the hood) and proxies their CDP WebSocket at `/connect/{sessionName}`. This skill covers the server — environment, connection flow, session rules, dashboard. Tool-specific connect/close mechanics live in the reference files; all tool command knowledge lives in each tool's own skill.

## Step 1: check the server

```bash
printenv BROWSERFULL_API_URL && echo OK || echo MISSING
curl -s "$BROWSERFULL_API_URL/health" -o /dev/null -w "HTTP %{http_code}\n"
```

**If `BROWSERFULL_API_URL` is MISSING: use the server URL the user gave in their request if there is one. If they gave none, stop and ask** (locally it defaults to `http://localhost:8080`).

If the server is not reachable: report that and stop — the task is blocked until the user fixes it. Never start the server, never inspect docker/k8s to hunt for one, never fall back to a direct local browser or another automation tool.

## Step 2: select tool and load its skill

**NEVER write agent-browser or browser-use commands from memory. Always load the tool's skill and follow it exactly.** Reference files below only cover how the tool attaches to a browserfull session — they carry no command syntax.

| Situation | Action |
|-----------|--------|
| Only agent-browser installed | Load the agent-browser skill, then read [references/agent-browser.md](references/agent-browser.md) |
| Only browser-use installed | Load the browser-use skill, then read [references/browser-use.md](references/browser-use.md) |
| Both installed, tool unspecified | agent-browser — it is the default preference; no need to ask |
| Neither skill installed | Give install options below, ask, and **stop and wait** |

Install options (offer and ask first — never run without an actual user "yes"):

- agent-browser skill: `npx skills add https://github.com/vercel-labs/agent-browser --skill agent-browser`
- browser-use: `uv tool install browser-use && browser-use skill install`

The server host usually has the **CLI** — CLI presence does NOT mean the skill is installed. The skill is what carries correct usage. If nothing grants permission, report the task as blocked; never install silently, never proceed from memory.

## Sessions

```bash
curl -s "$BROWSERFULL_API_URL/sessions"        # JSON array of active session names
curl -s -X DELETE "$BROWSERFULL_API_URL/sessions/my-task"   # close one (204)
```

- CDP URL: `ws://host:port/connect/{sessionName}` (`wss://` if the server sits behind https); `sessionName` must match `^[a-zA-Z0-9_-]+$`
- New name → a browser launches on demand; same name → reconnects to the existing session
- Reuse an existing session with a suitable name before opening a new one
- **Close only via `DELETE /sessions/{name}` — never a tool-level close.** A CLI close does not remove the session server-side: the name still shows in `/sessions`. DELETE is the only call that fully stops the session and removes it from the list
- Already ran a tool-level close? Harmless but insufficient — run DELETE anyway and verify `/sessions`
- Closing destroys all browser state in that session. Only close when done; if unsure, ask the user
- Before claiming a session is closed, re-check `/sessions` until the name is gone

## Dashboard / live view

The agent-browser dashboard is served at the server root, `http://host:port/`: session status and live view.

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| Connection refused / empty reply | Server down or wrong port. Confirm with the user. |
| 404 on `/connect/...` | Wrong path — it is exactly `/connect/{sessionName}` |
| 400 Bad Request on close or connect | `sessionName` must match `^[a-zA-Z0-9_-]+$` |
| Session name still in `/sessions` after tool close | Expected — close with `DELETE /sessions/{name}` instead |
| Origin blocked at WS connect | Server config: user must set `BROWSERFULL_ALLOWED_ORIGINS` (or `*`) on the server |

## Red flags — stop

- "They probably approved" / "the task implies approval" — approval is an actual user reply saying yes; nothing else counts
- "I'll just start the server / check docker for one" — the server belongs to the user; report and wait
- Writing agent-browser or browser-use commands without the tool's skill loaded
- "`agent-browser close` is close enough" — only `DELETE /sessions/{name}` plus an empty `/sessions` check counts