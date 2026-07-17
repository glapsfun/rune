# Contract: `rune lsp`

Starts the Rune language server over stdio (JSON-RPC 2.0, LSP 3.17). Executes nothing (FR-028).

## Synopsis

```
rune lsp [--log-file PATH] [--log-level LEVEL]
```

| Flag | Default | Meaning |
|------|---------|---------|
| `--log-file` | none | write logs to PATH instead of stderr |
| `--log-level` | `info` | `error` \| `warn` \| `info` \| `debug` |

## Transport & I/O discipline (FR-011, FR-012)

- Transport: stdin/stdout, `Content-Length`-framed JSON-RPC 2.0. stdio only — no TCP.
- **stdout carries protocol bytes ONLY.** All logs go to stderr or `--log-file`. Any stray stdout write is a defect (protocol corruption).

## Lifecycle

`initialize` → `initialized` → (requests/notifications) → `shutdown` → `exit`. The server exits 0 after a clean `shutdown`+`exit`, non-zero if `exit` arrives without a prior `shutdown`.

## `initialize` response capabilities (FR-014 — advertise only what is implemented)

```json
{
  "capabilities": {
    "textDocumentSync": { "openClose": true, "change": 2, "save": { "includeText": false } },
    "completionProvider": { "triggerCharacters": ["[", "(", ".", "{", ":"] },
    "definitionProvider": true,
    "hoverProvider": true,
    "documentSymbolProvider": true,
    "documentFormattingProvider": true
  },
  "serverInfo": { "name": "rune", "version": "<build version>" }
}
```

`serverInfo.version` is the real build version of the running binary (not a hard-coded `0.8.0`).

## Supported messages

**Notifications (client→server)**: `initialized`, `textDocument/didOpen`, `textDocument/didChange` (incremental), `textDocument/didSave`, `textDocument/didClose`, `workspace/didChangeWatchedFiles`, `$/cancelRequest`, `exit`.

**Requests (client→server)**: `initialize`, `textDocument/completion`, `textDocument/definition`, `textDocument/hover`, `textDocument/documentSymbol`, `textDocument/formatting`, `shutdown`.

**Notifications (server→client)**: `textDocument/publishDiagnostics`.

**Requests (server→client)**: `client/registerCapability` — sent after `initialized` to dynamically register `workspace/didChangeWatchedFiles` for `**/Runefile` and `**/*.rune`, so on-disk edits to imported files (not open in the editor) refresh dependents (FR-022). If the client does not support dynamic registration, the server falls back to client-driven watching.

## Behavior guarantees

| Capability | Contract |
|-----------|----------|
| Diagnostics | Published on open/change/save, on watched-file changes, and when imported files change (FR-013/022). Debounced ~100 ms; a superseded version's diagnostics are never published (FR-016). |
| Completion | Context-aware (dependency / variable+param / setting / attribute / executor / builtin); each item carries signature + documentation; private tasks only offered in their own file (FR-018/019a). |
| Definition | Resolves dependency, variable, parameter, module namespace (`mod`), flat-imported task (bare name), and module task (`name::task`) — cross-file; uses overlay content for open imports (FR-019). |
| Hover | Task / parameter / attribute / builtin info from the shared registry (FR-017). |
| Document symbols | Settings, variables, imports, tasks, modules (FR-017). |
| Formatting | Full-document `TextEdit` from `internal/formatter.Format`; never spawns a process; never writes files (FR-020). |

## Safety (FR-028)

No request or notification may run a task, shell, Python/Node, or agent; make a Runefile-driven network request; expand secrets into messages; or write project files. Enforced structurally (no `runtime`/`os/exec`/net imports in `analysis`/`language`/`lsp`) and behaviorally (SC-005 test).

## Malformed input

Malformed JSON-RPC messages MUST NOT crash the server (fuzz-tested); they are logged (stderr/file) and skipped or answered with the appropriate JSON-RPC error.
