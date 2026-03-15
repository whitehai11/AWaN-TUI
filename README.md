# AWaN-TUI

AWaN-TUI is a lightweight terminal interface for AWaN Core.

It also checks for updates asynchronously on startup using a shared `internal/updater` module.

It connects to the local runtime by default at:

```text
http://localhost:7452
```

## Features

- terminal chat with agent
- memory viewer
- agent selection

## Layout

- left panel: agent list
- center panel: chat or memory view
- bottom panel: command input

## Keybindings

- `up` / `down`: change selected agent
- `tab`: toggle chat and memory view
- `enter`: send prompt to `/agent/run`
- `ctrl+r`: refresh memory from `/memory`
- `esc`: clear input
- `q` or `ctrl+c`: quit

## Run

```bash
go run ./cmd
```

If your AWaN Core runtime uses a different endpoint:

```bash
set AWAN_CORE_URL=http://localhost:7452
go run ./cmd
```

## Auto Updates

AWaN-TUI defines a local version constant and checks the latest GitHub release for `whitehai11/AWaN-TUI` in the background at startup.

To disable auto updates for AWaN apps:

```text
~/.awan/config/runtime.awan
```

```text
auto_update = false
```
