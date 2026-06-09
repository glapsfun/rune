# OS filtering

> **Use case:** keep platform-specific setup tasks in one Runefile — each only shows and runs
> on the OS it targets.

**Demonstrates:** OS filtering  ·  **Guide:** [Attributes](../../runefile.md#attributes)

**Prerequisites:** none

## Run it

```sh
rune info
```

## Expected output

```text
running on linux/amd64
```

(The platform reflects *your* machine — e.g. `darwin/arm64` on an Apple Silicon Mac.) Run
`rune --list` and you'll only see the setup task for your current OS; the others are hidden.

## How it works

`[linux]`, `[macos]`, and `[windows]` restrict a task to one OS. The `os()` and `arch()`
built-ins report the current platform — handy for cross-platform call-outs.
