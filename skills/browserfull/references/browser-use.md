# browser-use Reference

Read SKILL.md first — server check and the CDP URL to substitute below. Helpers, syntax and workflow come from the loaded browser-use skill (`browser-use skill show` prints the upstream reference for the installed CLI); this file only covers attaching browser-use to a browserfull session.

## Connect

Build `<cdp-url>` from the server URL per SKILL.md: `ws://<server>/connect/<session>` (`wss://` behind https; locally, e.g. `ws://localhost:8080/connect/my-session`).

```bash
export BU_CDP_WS="<cdp-url>"
browser-use <<'PY'
new_tab("https://example.com")
wait_for_load()
print(page_info())
PY
```

- Must be `BU_CDP_WS` (WebSocket). `BU_CDP_URL` only speaks http/https and will NOT connect to browserfull
- The daemon caches the CDP URL at start-up: to switch browserfull session, restart the daemon — kill the pidfile process and delete the `.pid` + `.sock` under `~/.config/browser-harness/runtime/`, re-export `BU_CDP_WS`, run any command
- Page titles carry a `🐴` marker appended by the harness; set `BH_TAB_MARKER=0` before daemon start for clean titles

## Close

- browser-use has no remote close — end the browserfull session with `DELETE /sessions/<name>` (SKILL.md), then stop the daemon as above
- Closing destroys all browser state in the session; only close when the work is done

## Troubleshooting

| Error | Action |
|-------|--------|
| Element not found | Re-query the accessibility tree after page changes (per the browser-use skill) |
| Connected to the wrong / old page | Daemon cached a previous `BU_CDP_WS` — restart the daemon, re-export, reconnect |