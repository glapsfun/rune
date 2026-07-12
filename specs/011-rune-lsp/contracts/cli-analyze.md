# Contract: `rune analyze`

Standalone static analysis of a Runefile and its transitive imports. Executes nothing (FR-023, FR-028).

## Synopsis

```
rune analyze [path] [--json]
```

| Arg / Flag | Default | Meaning |
|------------|---------|---------|
| `path` | `Runefile` (in cwd) | entry Runefile to analyze |
| `--json` | off | emit machine-readable diagnostics instead of human text |

## Behavior

- Runs the shared `analysis.Service` on `path` **with its transitive imports/mods**; reports diagnostics from all files, each attributed to its own `file:line:col` (FR-009a).
- Never runs a task, shell, agent, or network request; never writes project files.

## Human output (default)

One line per diagnostic, then a summary:

```
Runefile:12:9: error[RUNE2001]: unknown dependency "buid"
Runefile:24:1: warning[RUNE2010]: public task "release" has no documentation
1 error, 1 warning
```

- Line format: `FILE:LINE:COL: SEVERITY[CODE]: MESSAGE`.
- Summary counts errors and warnings (pluralized naturally). Zero-diagnostic runs print a success/empty summary.

## JSON output (`--json`)

A structured object on stdout: `{ "diagnostics": [...], "errors": N, "warnings": M }`. Each diagnostic has `file`, `code` (omitted when uncoded), `severity`, `message`, `range` (`{start:{line,column,offset}, end:{...}}` — 1-based line/column in **bytes**, plus a 0-based byte `offset`), and `related` (array of `{file, message, range}`). Byte-based positions match the human `file:line:col` output and keep the CLI free of the UTF-16 conversion, which is an LSP-only concern (the language server emits UTF-16 ranges in `publishDiagnostics`).

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | No error-severity diagnostics (warnings allowed) |
| `2` | Usage/discovery failure: no Runefile found, or it cannot be read |
| `3` | At least one error-severity diagnostic |

`rune analyze` uses Rune's global exit-code scheme (0 success · 2 usage · 3 validation), so it behaves consistently with `rune` itself. This supersedes the feature spec's original "1 = internal failure" (FR-025), which predated checking Rune's established codes. Warnings (e.g. `RUNE2010`) never by themselves cause a non-zero exit.

## Invariants

- Diagnostics are identical to what the LSP publishes for the same content (FR-002 / SC-002).
- stdout carries only analysis output; incidental logs (if any) go to stderr.
