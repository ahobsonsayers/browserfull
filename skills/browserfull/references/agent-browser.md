# agent-browser Reference

Read SKILL.md first — server check and the CDP URL to substitute below. Command knowledge always comes from the loaded agent-browser skill; this file only covers how agent-browser attaches to a browserfull session.

## Connect — persistent mode

Build `<cdp-url>` from the server URL per SKILL.md: `ws://<server>/connect/<session>` (`wss://` behind https; locally, e.g. `ws://localhost:8080/connect/my-session`).

```bash
agent-browser connect "<cdp-url>"
agent-browser snapshot   # same link persists across invocations
```

`connect` keeps the link to the browserfull session across commands. Avoid per-command `--cdp "<cdp-url>"`: that mode can drop the tab binding between commands and needs `--pin-tab` plus a `--session` name to stay put. Use `--cdp` only for one-shot commands.

## Rules

- Never `agent-browser install` — browserfull launches the browser server-side
- Never `agent-browser close` to end a browserfull session — that only closes the local link/browser, the session stays listed. Always `DELETE /sessions/<name>` per SKILL.md
- Connection or page in a bad state: retry on the same URL first; still wedged → `DELETE /sessions/<name>`, then connect fresh

## Closing

`DELETE /sessions/<name>` via the API (SKILL.md), then re-check `/sessions` until the name is gone. `agent-browser close` before that is optional cleanup of the local link, never a substitute.