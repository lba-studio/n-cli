# Codex Hook Payload Schema

Source: https://developers.openai.com/codex/hooks

## Common fields (all events)

| Field | Type | Description |
|---|---|---|
| `hook_event_name` | string | Event type (e.g. `Stop`, `PermissionRequest`) |
| `session_id` | string | UUID identifying the session |
| `transcript_path` | string \| null | Path to session transcript file, if any |
| `cwd` | string | Current working directory |
| `model` | string | Model used |
| `permission_mode` | string | `default`, `acceptEdits`, `plan`, `dontAsk`, or `bypassPermissions` |

## Stop event (additional fields)

| Field | Type | Description |
|---|---|---|
| `turn_id` | string | |
| `stop_hook_active` | boolean | |
| `last_assistant_message` | string \| null | |

## PermissionRequest event (additional fields)

| Field | Type | Description |
|---|---|---|
| `turn_id` | string | |
| `tool_name` | string | Tool requiring approval |
| `tool_input` | JSON value | Input to the tool |
| `tool_input.description` | string \| null | Human-readable description of the request |

## Session naming

Session names (set via `/session rename`) are stored in `~/.codex/session_index.jsonl`.
Each line is a JSON object:

```json
{"id":"<session-uuid>","thread_name":"<name>","updated_at":"<iso8601>"}
```

The file is append-only. A rename appends a new line with the same `id` and the new name.
To resolve the current name for a session, find the **last** line matching that `id`.

Transcript paths are UUID/timestamp-based (`rollout-YYYY-MM-DDTHH-MM-SS-<uuid>.jsonl`)
and do not reflect the session name.
