# OS filtering

> **Use case:** keep platform-specific setup tasks in one Runefile — each only shows and runs
> on the OS it targets, and one cross-platform task dispatches to the right one.

**Demonstrates:** OS filtering + per-OS dispatch  ·  **Guide:** [Attributes](../../runefile.md#attributes)

**Prerequisites:** none

## Run it

```sh
rune setup
```

## Expected output

```text
apt-get install build-essential
toolchain ready
```

(The first line reflects *your* platform — `brew install coreutils` on macOS,
`choco install make` on Windows.) Run `rune --list` and you'll only see the
setup task for your current OS; the others are hidden. Try `rune info` for the
`os()`/`arch()` built-ins.

## How it works

`[linux]`, `[macos]`, and `[windows]` restrict a task to one OS ([`unix`]
matches everything except Windows). On a non-matching host the task is hidden
from listings and the MCP tool list, refuses direct invocation (exit 3), and
is **silently skipped as a dependency** — which is what lets the
cross-platform `setup` task depend on all three variants and run exactly the
matching one before its own body.
