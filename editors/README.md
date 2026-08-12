# Rune editor integration

Rune ships a language server in the `rune` binary itself — `rune lsp` speaks
JSON-RPC/LSP 3.17 over stdio. Any LSP-capable editor can use it; there is no
separate server to install. All you configure per editor is:

1. a **filetype** for Runefiles (`Runefile`, `.runefile`, and `*.rune`), and
2. a **language server command**: `rune lsp`.

> Prerequisite: `rune` on your `PATH`. Check with `rune version`. See the
> [installation guide](../docs/installation.md) if you don't have it yet.

| Editor | How it integrates | Setup |
|--------|-------------------|-------|
| [VS Code](#vs-code) | Marketplace / Open VSX extension | Install **Rune Task Runner** |
| [Neovim](#neovim-011-built-in-lsp) | Built-in LSP client (0.11+) | Lua snippet |
| [Helix](#helix-confighelixlanguagestoml) | Built-in LSP client | `languages.toml` snippet |
| [Zed](#zed) | Small Zed extension | Extension scaffold |
| [Emacs](#emacs-29-eglot) | Built-in Eglot (29+) | Elisp snippet |
| [Sublime Text](#sublime-text-lsp-package) | [LSP](https://packagecontrol.io/packages/LSP) package | Syntax stub + client config |

## What you get

Every editor below gets the same capabilities, served by the same parser and
analyzer that power `rune` itself:

- **Diagnostics** in real time, with stable `RUNE####` codes — the same
  analysis that runs before every `rune` execution.
- **Completion** for dependencies, variables/parameters, settings, attributes,
  executors, and built-in functions.
- **Go-to-definition** for tasks (including across imports), variables, and
  parameters.
- **Hover** — signature, doc comment, executor, group, and location.
- **Document symbols** — an outline of settings / variables / imports / tasks.
- **Formatting** — Rune's canonical formatter (`rune fmt`).

Nothing is executed while editing: the server never runs tasks or shell
commands.

## VS Code

Install **Rune Task Runner** from the [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=rune-task-runner.runefile)
(extension ID `rune-task-runner.runefile`) — search "Rune" in the Extensions
view. VSCodium and other VS Code-compatible editors get the same extension
from [Open VSX](https://open-vsx.org/extension/rune-task-runner/runefile).

Offline/air-gapped: every stable [GitHub release](https://github.com/glapsfun/rune/releases)
attaches the `runefile-<version>.vsix`; install it with
`code --install-extension runefile-<version>.vsix`.

The extension's own README covers settings (`rune.path`, `rune.trace.server`)
and troubleshooting — see [`vscode/`](./vscode/). Contributors can build from
source there (or press F5 to debug).

## Neovim (0.11+, built-in LSP)

```lua
-- Recognize Runefiles as the 'runefile' filetype.
vim.filetype.add({
  filename = { ["Runefile"] = "runefile", [".runefile"] = "runefile" },
  extension = { rune = "runefile" },
})

-- Start `rune lsp` for Runefiles.
vim.api.nvim_create_autocmd("FileType", {
  pattern = "runefile",
  callback = function(args)
    vim.lsp.start({
      name = "rune",
      cmd = { "rune", "lsp" },
      root_dir = vim.fs.root(args.buf, { "Runefile", ".git" }),
    })
  end,
})
```

With `nvim-lspconfig` you can instead define a custom config whose `cmd` is
`{ "rune", "lsp" }` and `filetypes = { "runefile" }`.

## Helix (`~/.config/helix/languages.toml`)

```toml
[[language]]
name = "runefile"
scope = "source.rune"
file-types = [{ glob = "Runefile" }, { glob = ".runefile" }, "rune"]
comment-tokens = ["#"]
indent = { tab-width = 4, unit = "    " }
language-servers = ["rune"]

[language-server.rune]
command = "rune"
args = ["lsp"]
```

## Zed

Zed language servers are provided by a small Zed extension. In an extension's
`extension.toml`, register the language and a language server whose binary is
`rune` with arguments `["lsp"]`:

```toml
[language_servers.rune]
name = "Rune"
languages = ["Runefile"]

# In the extension's Rust `language_server_command`, return:
#   command = "rune", args = ["lsp"]
```

Add a `languages/runefile/config.toml` with `name = "Runefile"`, a
`grammar`/`path-suffixes = ["rune"]`, and `line_comments = ["# "]`. See the Zed
extension docs for the scaffold; the only Rune-specific parts are the filetype
globs and the `rune lsp` command.

## Emacs (29+, Eglot)

Emacs 29 ships [Eglot](https://www.gnu.org/software/emacs/manual/html_mono/eglot.html)
built in. Define a minimal major mode for Runefiles and point Eglot at
`rune lsp` (in your `init.el`):

```elisp
;; A minimal major mode for Runefiles.
(define-derived-mode runefile-mode prog-mode "Runefile"
  "Major mode for Rune task files."
  (setq-local comment-start "# ")
  (setq-local comment-start-skip "#+\\s-*"))

;; Recognize Runefiles.
(add-to-list 'auto-mode-alist '("/\\.?[Rr]unefile\\'" . runefile-mode))
(add-to-list 'auto-mode-alist '("\\.rune\\'" . runefile-mode))

;; Serve them with `rune lsp`.
(with-eval-after-load 'eglot
  (add-to-list 'eglot-server-programs '(runefile-mode . ("rune" "lsp"))))
```

Then `M-x eglot` in a Runefile buffer (or add `runefile-mode-hook` →
`eglot-ensure` to start it automatically). `lsp-mode` users: register a client
whose `new-connection` is `(lsp-stdio-connection '("rune" "lsp"))` with
`:major-modes '(runefile-mode)`.

## Sublime Text (LSP package)

Install the [LSP](https://packagecontrol.io/packages/LSP) package first.
Sublime matches language servers by scope, so Runefiles need a syntax that
assigns `source.rune`. Save this as
`Packages/User/Runefile.sublime-syntax`:

```yaml
%YAML 1.2
---
name: Runefile
scope: source.rune
file_extensions: [rune, runefile]
contexts:
  main:
    - match: '#.*$'
      scope: comment.line.number-sign.rune
```

(For the extensionless `Runefile` itself, open one and pick
**View → Syntax → Open all with current extension as… → Runefile**.)

Then register the server in **Preferences → Package Settings → LSP →
Settings** (`LSP.sublime-settings`):

```json
{
  "clients": {
    "rune": {
      "enabled": true,
      "command": ["rune", "lsp"],
      "selector": "source.rune"
    }
  }
}
```

## Manual validation checklist (quickstart §9)

These steps require a running editor GUI and are performed by a human. For each
editor you configure, open a Runefile and confirm:

1. **Diagnostics** — introduce `deploy: missing`; an error underlines `missing`
   with code `RUNE2001`. Fix it; the error clears.
2. **Completion** — type `deploy: bu` and trigger completion → `build` is
   suggested with its signature.
3. **Go-to-definition** — invoke on a dependency → jumps to the task; on a
   `mod` task (`ns::task`) → opens the module file.
4. **Hover** — hover a task → signature, doc, executor, group, and location.
5. **Document symbols** — the outline lists settings/variables/imports/tasks.
6. **Formatting** — Format Document produces Rune's canonical output.
7. **Safety** — no task or shell command runs during any of the above.

Recommended coverage: VS Code plus at least one of Neovim/Helix.
