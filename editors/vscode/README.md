# Rune for VS Code

Language support for Runefiles, powered by the embedded `rune lsp` server:
real-time diagnostics, context-aware completion, go-to-definition (including
across imports), hover documentation, a document outline, and canonical
formatting. Nothing is executed while you edit.

## Requirements

- The `rune` binary on your `PATH` (or set `rune.path`). Verify with `rune version`.

## Install (from source)

```sh
cd editors/vscode
npm install
npm run package     # produces rune-<version>.vsix
code --install-extension rune-*.vsix
```

Or press **F5** in this folder to launch an Extension Development Host.

## Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `rune.path` | `rune` | Path to the `rune` executable. |
| `rune.trace.server` | `off` | Trace JSON-RPC traffic (`off` / `messages` / `verbose`). |

## What you get

- Diagnostics with stable `RUNE####` codes as you type.
- Completion for dependencies, variables/parameters, settings, attributes,
  executors, and built-in functions.
- Go-to-definition and hover for tasks, variables, parameters, attributes, and
  built-ins.
- Outline (document symbols) grouped by settings / variables / imports / tasks.
- Format Document runs Rune's canonical formatter.
