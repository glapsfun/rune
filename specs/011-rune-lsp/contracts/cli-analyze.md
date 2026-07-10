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

A structured array (stdout) where each diagnostic includes at least: `code`, `severity`, `message`, `range` (`{start:{line,character}, end:{line,character}}`, 0-based UTF-16), `file`, and `related` (array of `{file, range, message}`). Exact schema is stabilized during implementation and covered by a golden test.

## Exit codes (FR-025)

| Code | Meaning |
|------|---------|
| `0` | No error-severity diagnostics (warnings allowed) |
| `3` | At least one error-severity diagnostic |
| `1` | Internal failure (I/O, unexpected error) — distinct from "errors found" |

Warnings (e.g. `RUNE2010`) never by themselves cause exit `3`.

## Invariants

- Diagnostics are identical to what the LSP publishes for the same content (FR-002 / SC-002).
- stdout carries only analysis output; incidental logs (if any) go to stderr.
