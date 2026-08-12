# Rune for VS Code

Language support for [Runefiles](https://github.com/glapsfun/rune) — the
task files of the Rune command runner — powered by the language server
embedded in the `rune` binary (`rune lsp`). Nothing is executed while you
edit.

> Using VSCodium or another VS Code-compatible editor? The same extension is
> on [Open VSX](https://open-vsx.org/extension/rune-task-runner/runefile).

## Quick start

1. **Install Rune** — see the
   [installation guide](https://github.com/glapsfun/rune/blob/main/docs/installation.md)
   (Homebrew, install script, Scoop, or a release binary). Verify with
   `rune version`.
2. **Install the extension** — from the
   [Marketplace](https://marketplace.visualstudio.com/items?itemName=rune-task-runner.runefile),
   or via the Quick Open palette (`Ctrl+P` / `Cmd+P`):
   `ext install rune-task-runner.runefile`.
3. **Open a `Runefile`** — diagnostics, completion, and the rest activate
   automatically.

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

## Troubleshooting

- **"Rune binary not found" notification** — the extension could not run the
  `rune` executable. Install Rune (see Requirements) or set `rune.path` to
  the absolute path of the binary — changing `rune.path` takes effect
  immediately, no window reload needed. The notification's **Retry** button
  re-probes after you've installed.
- **"…cannot run `rune lsp`" notification** — your `rune` binary predates the
  language server. Upgrade Rune (any release with `rune lsp` works).
- **Inspect what the server is doing** — set `rune.trace.server` to
  `messages` or `verbose` and read the **Rune Language Server** output
  channel (View → Output).

## Other editors

The same language server works in Neovim, Helix, Zed, Emacs, and Sublime Text
— setup snippets live in the repository's
[editor integration guide](https://github.com/glapsfun/rune/tree/main/editors).

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
npx vsce package 0.0.1 --no-update-package-json   # produces runefile-0.0.1.vsix
code --install-extension runefile-0.0.1.vsix
```

(`--no-update-package-json` keeps the repo's `0.0.0` placeholder version
untouched — the working tree stays clean.)

Or open the folder in VS Code and press **F5** for an Extension Development
Host. Issues and PRs: [github.com/glapsfun/rune](https://github.com/glapsfun/rune/issues).
