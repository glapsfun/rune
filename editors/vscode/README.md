# Rune for VS Code

Language support for [Runefiles](https://github.com/glapsfun/rune) — the
task files of the Rune command runner — powered by the language server
embedded in the `rune` binary (`rune lsp`). Nothing is executed while you
edit.

> Using VSCodium or another VS Code-compatible editor? The same extension is
> on [Open VSX](https://open-vsx.org/extension/rune-task-runner/rune).

## Requirements

**This extension needs the `rune` binary** — it is not bundled. Install
Rune first ([installation guide](https://github.com/glapsfun/rune/blob/main/docs/installation.md)),
make sure `rune` is on your `PATH` (verify with `rune version`), or point
the `rune.path` setting at the executable. If the binary is missing or too
old to serve the language server, the extension shows a notification with
install/upgrade guidance instead of starting.

## Features

Open any `Runefile`, `.runefile`, or `*.rune` file and you get:

- **Diagnostics** with stable `RUNE####` codes as you type — the same
  analyzer that runs before every `rune` execution.
- **Completion** for dependencies, variables/parameters, settings,
  attributes, executors, and built-in functions.
- **Go-to-definition** and **hover** for tasks (including across imports),
  variables, parameters, attributes, and built-ins.
- **Outline** (document symbols) grouped by settings / variables / imports
  / tasks.
- **Formatting** — Format Document runs Rune's canonical formatter.
- **Syntax highlighting** via a TextMate grammar.

## Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `rune.path` | `rune` | Path to the `rune` executable. |
| `rune.trace.server` | `off` | Trace JSON-RPC traffic (`off` / `messages` / `verbose`). |

## Versioning

The extension is versioned in lockstep with Rune: extension `X.Y.Z` is
published by Rune release `vX.Y.Z`. Any reasonably recent `rune` binary
works — the protocol surface is stable LSP 3.17.

## Contributing / building from source

The extension lives in [`editors/vscode/`](https://github.com/glapsfun/rune/tree/main/editors/vscode)
of the Rune repository. To build and side-load it:

```sh
cd editors/vscode
npm ci
npx vsce package 0.0.1 --no-update-package-json   # produces rune-0.0.1.vsix
code --install-extension rune-0.0.1.vsix
```

(`--no-update-package-json` keeps the repo's `0.0.0` placeholder version
untouched — the working tree stays clean.)

Or open the folder in VS Code and press **F5** for an Extension Development
Host. Issues and PRs: [github.com/glapsfun/rune](https://github.com/glapsfun/rune/issues).
