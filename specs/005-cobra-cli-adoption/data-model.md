# Phase 1 Data Model: Modern CLI Interface

**Feature**: `005-cobra-cli-adoption` · **Date**: 2026-06-10

This feature has no persistent data. The "entities" are the in-memory CLI structures
and the small new types introduced in `internal/cli`. Cobra types stay in `package
main`; `internal/cli` stays framework-free.

## Entities

### Built-in Command (Cobra `*cobra.Command`, in `package main`)

A reserved, discoverable capability of the CLI.

| Field | Type | Notes |
|-------|------|-------|
| Use | string | Name + usage line (e.g. `serve`, `version`). |
| Aliases | []string | `serve` carries `["mcp"]`. |
| Short / Long | string | One-line + full description (FR-002). |
| Example | string | ≥1 concrete example per command (FR-002, SC-002). |
| Flags | flagset | Local flags (e.g. serve's `--http`/`--addr`/`--token-file`/`--mcp`). |
| RunE | func | Returns error (never `os.Exit`); maps to `cli.*` error types. |
| ValidArgsFunction | CompletionFunc | **Root only** — dynamic task-name completion (D3). |
| Hidden | bool | False for all (every built-in is discoverable, SC-001). |

**Reserved names** (precedence over tasks): `serve`, `mcp`, `completion`, `help`,
`version`. Built-ins win; `rune -- <task>` escapes (FR-008).

**Relationships**: root *has many* subcommands; `serve` *aliases* `mcp`. Root also
holds the global task-runner flags and the task `RunE`/`ValidArgsFunction`.

### Task Invocation (existing — `internal/cli`)

A dynamic, positional request to run a Runefile task. Unchanged: resolved at runtime by
`config.Resolve` → `parser.Parse` → `analyzer.Analyze` → `splitArgs` →
`scheduler.Run`. Not part of the fixed command set; supplied by the user as positional
args after global flags. Arguments after the task name are bound by arity
(`splitArgs`/`paramCapacity`) and passed untouched (FR-007).

### TaskCandidate (NEW — `internal/cli/complete.go`)

Completion input, framework-free.

| Field | Type | Notes |
|-------|------|-------|
| Name | string | Task name. |
| Doc | string | First line of the task's doc comment (`firstLine(t.Doc)`). |

**Producer**: `func TaskCandidates(opts Options) []TaskCandidate` — resolves + parses +
composes the Runefile, includes only tasks where `!IsPrivate()` and `osMatches(t,
runtime.GOOS)` (same predicate as `listTasks`). **Returns `nil` on any error** (graceful
degradation, FR-012). Does **not** run the analyzer.

**Consumer**: `package main` maps `[]TaskCandidate` → `[]cobra.Completion` via
`cobra.CompletionWithDesc(c.Name, c.Doc)`, returning `ShellCompDirectiveNoFileComp`.

### Flag (Cobra/pflag, in `package main`)

| Attribute | Notes |
|-----------|-------|
| Name / Shorthand | e.g. `--file`/`-f`; serve `--http`, `--addr`, `--token-file`, `--mcp`. |
| Type / Default | bool/string; defaults preserve current behavior. |
| Scope | Global flags live on root; serve flags are local to `serve`. |
| Validation | serve: `--addr`/`--token-file` require `--http` → `UsageError` (FR-015). |
| Completion | `--token-file` → file completion via `RegisterFlagCompletionFunc`. |

### Options (existing — `internal/cli/dispatch.go`, ONE new field)

Carries resolved global flags + I/O streams. **Add** `Commands []string` — the reserved
command names, injected by `main`, used by `splitArgs` to include commands in
did-you-mean suggestions (D6). All other fields unchanged.

### Suggestion (NEW — `internal/cli/suggest.go`)

Pure helper, no state: `nearest(token string, candidates []string) (string, bool)` —
Levenshtein distance; returns the closest candidate when distance ≤ 2 (or ≤ ⌊len/3⌋).
Candidates = task names ∪ `opts.Commands`. Used to enrich the `unknown task` error.

## Validation Rules (cross-references)

- Every built-in command MUST have `Short`, `Long`, and `Example` (FR-002, SC-002).
- Completion MUST never error/panic into the shell; `TaskCandidates` returns `nil` on
  failure (FR-012).
- serve flag validation MUST produce exit 2 on conditional-requires violations (FR-015).
- Exit-code mapping is owned by `internal/cli/exit.go` and is unchanged (FR-018).

## State Transitions

None. Each invocation is stateless: parse args → route to a command → run → map exit
code. Signal (SIGINT/SIGTERM) cancels the shared context → exit 130 (FR-022), unchanged.
