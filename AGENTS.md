# Agents

If a `.cursor/rules` directory or file exists in this repository, read and follow the rules defined there.

**Maintenance:** When adding new commands or subcommands, update both `AGENTS.md` and `README.md` to reflect the change.

## Project Overview

`n-cli` is a notification CLI tool that lets you send notifications to yourself (desktop, Discord, Slack, custom webhooks) from the command line. A primary use case is getting notified when long-running commands or LLM coding agents finish.

For the full list of commands and configuration options, see `README.md`.

## Folder Structure

```
cmd/            # Cobra command definitions — one file per top-level command
  setup/        # Subcommands for agent integrations (Cursor, Codex, Claude Code)
  where/        # Subcommands for the `where` command
internal/
  config/       # Config loading and initialisation (viper-based, ~/.n-cli/config.yaml)
pkg/
  notifier/     # Core notification logic — each channel (Discord, Slack, custom, system) has its own file
    marker/     # Tracks command run metrics (duration, exit code) for `n-cli run`
    webhook/    # Shared HTTP webhook utilities
  formatter/    # Message formatting helpers
  monitor/      # Process monitoring utilities
```

## Examples

**Adding a new notification channel** (e.g. Telegram): add a `telegram.go` in `pkg/notifier/`, implement the `Notifier` interface, and wire it up in `pkg/notifier/notify.go` alongside the existing channels. Add its config struct to `internal/config`.

**Adding a new top-level command**: create `cmd/cmd_<name>.go`, define a `New<Name>Cmd() *cobra.Command` function, and register it in `cmd/root.go`'s `init()`.

**Adding a new `setup` subcommand** (e.g. for a new agent integration): add a file in `cmd/setup/`, implement the setup logic there, and register it in `cmd/cmd_setup.go`'s `NewSetupCmd()`.

## Releasing a New Version

Use `release.sh` to bump the version, commit, tag, and push in one step:

```sh
./release.sh          # patch bump (default) — for bug fixes, refactors, etc.
./release.sh minor    # minor bump — required when adding new commands or subcommands
./release.sh patch    # patch bump — explicit form of the default
```

Semver rules:
- **minor** — any new command or subcommand added
- **patch** — everything else (fixes, refactors, dependency updates, docs)

The script updates `pkg/version/get_version.go`, commits as `release vX.Y.Z`, creates a tag `vX.Y.Z`, and pushes both the commit and the tag to `origin`.
