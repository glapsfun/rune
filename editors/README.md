# Rune editor integration

Rune ships a language server in the `rune` binary itself — `rune lsp` speaks
JSON-RPC/LSP 3.17 over stdio. Any LSP-capable editor can use it; there is no
separate server to install. All you configure per editor is:

1. a **filetype** for Runefiles (`Runefile`, `.runefile`, and `*.rune`), and
2. a **language server command**: `rune lsp`.

Capabilities: real-time diagnostics, completion, go-to-definition (incl. across
imports), hover, document symbols, and formatting. Nothing is executed while
editing.

> Prerequisite: `rune` on your `PATH`. Check with `rune version`.

## VS Code

See [`vscode/`](./vscode/) — a ready client extension. `npm install && npm run
package`, then install the `.vsix` (or press F5 to debug).

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
