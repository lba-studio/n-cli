# Hook subcommands

The `n-cli hook` subcommand is internal plumbing invoked by LLM agent tools (Cursor, Codex, Claude Code) when hook events fire. Users should not call it directly; use `n-cli setup <tool>` to register hooks instead.

Each subcommand reads JSON from stdin, formats a notification message, and calls `notifier.Notify`.
