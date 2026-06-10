# Phase 0 Research: Modern CLI Interface

**Feature**: `005-cobra-cli-adoption` · **Date**: 2026-06-10

Sources: `golang-cli` skill (project standard, Constitution VIII); Cobra package docs
(`https://pkg.go.dev/github.com/spf13/cobra`, v1.10.2) and the Cobra shell-completion
guide; the existing Rune CLI (`rune/main.go`, `internal/cli/*`).

There were no `NEEDS CLARIFICATION` markers from the spec; clarification already fixed
the escape hatch (`--`), completion descriptions (yes), and `serve`/`mcp` presentation
(`serve` primary, `mcp` alias). This document records the *implementation* decisions
those answers imply, grounded in the Cobra API.

---

## Decision 1 — Command tree: real subcommands + root task dispatcher

**Decision**: Keep the root command as the task dispatcher (`RunE` runs tasks via
`cli.Run`) and register real subcommands on it:

- `serve` — `Aliases: []string{"mcp"}`; flags `--http`, `--addr`, `--token-file`,
  `--mcp`.
- `version` — prints the same string as `--version`.
- `completion` and `help` — **auto-added by Cobra** once ≥1 subcommand exists
  (`InitDefaultCompletionCmd`); not hand-written.

Root keeps `SilenceUsage: true`, `SilenceErrors: true`, `root.Version`,
`root.Flags().SetInterspersed(false)`, and all global task-runner flags
(`-f/--file`, `--list`, `--dry-run`, `--summary`, `--dump`, `--format`, `--set`,
`--watch`, `--choose`, `--yes`, `--quiet`, `--fmt`, `--clear-cache`). The reserved
word set is exactly the subcommand names + aliases: `serve`, `mcp`, `completion`,
`help`, `version`.

**Why it works (Cobra routing)**: Cobra resolves the target command from the first
*non-flag* token. If that token matches a subcommand/alias, Cobra executes the
subcommand; otherwise the target is root and root's `RunE` receives the args. So:

- `rune build …` → `build` is not a subcommand → root `RunE` → `cli.Run` (task). ✓
- `rune serve --http` → routes to `serve`, which parses its own flags. ✓ (built-in wins)
- `rune --list` → flag only, no subcommand → root `RunE`/flag handling (list). ✓

`SetInterspersed(false)` only affects flag *parsing after the first positional* on the
resolved command — it does not change which command is selected — so task arg/flag
pass-through (`rune build --watch` → `--watch` goes to the task) is preserved exactly
as today.

**Rationale**: This is the smallest change that turns hidden dispatch into discoverable,
idiomatic Cobra subcommands while keeping the dynamic-task model. It satisfies FR-001..
FR-005, FR-006/FR-007, FR-014, FR-023.

**Alternatives considered**:
- *Keep manual dispatch inside root `RunE`* (today). Rejected: not discoverable, no
  per-command help, no completion — the whole point of the feature.
- *`run` subcommand as the task entrypoint* (`rune run build`). Rejected: breaks the
  ergonomic `rune <task>` contract and backward compatibility (Governance).
- *`DisableFlagParsing` on root + fully manual parsing*. Rejected: loses Cobra's flag
  help/validation/completion — regresses the goal.

---

## Decision 2 — Collision escape hatch via `--` (clarification Q1)

**Decision**: A task whose name equals a reserved command is invoked as
`rune -- <task> [args…]`. No explicit code path is required for the *common* case:
Cobra's command resolution (`stripFlags`) stops at `--`, so tokens after `--` are not
considered for subcommand matching — the target stays root and `root.RunE` receives
`["<task>", …]`, which `cli.Run`/`splitArgs` already treats as a task invocation.
`cmd.ArgsLenAtDash()` is available if disambiguation is ever needed.

**Validation requirement**: Because this is a routing assumption, it MUST be covered by
an integration test asserting `rune -- serve` runs a *task* named `serve` (not the
server). This is the one empirical risk flagged in the plan's re-check.

**Rationale**: Minimal, standard Unix convention; keeps every task reachable (FR-008)
with zero new command surface.

**Alternatives considered**: a `run` subcommand or `--task` flag (more surface, less
conventional); "reserve only / not runnable" (rejected at clarification — drops
reachability).

---

## Decision 3 — Dynamic task-name completion with descriptions (clarification Q2)

**Decision**: Set `root.ValidArgsFunction` to a `cobra.CompletionFunc`:

```go
func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective)
```

It calls a new **cobra-free** helper `cli.TaskCandidates(opts)` returning
`[]cli.TaskCandidate{Name, Doc}` (non-private, OS-matching tasks — same filter as
`listTasks`), then maps each to `cobra.CompletionWithDesc(name, firstLineOfDoc)` and
returns `cobra.ShellCompDirectiveNoFileComp` (no file completion for task names).
`TaskCandidates` resolves the Runefile (`config.Resolve`), parses + composes
(`parser.Parse` + `config.Compose`), and extracts task name + first doc line. On **any**
error (no Runefile, parse failure) it returns `nil` → completion offers built-in
commands only and never errors into the shell (FR-012). It deliberately **skips the
analyzer** for speed and robustness (we want names even if a variable is undefined).

Cobra automatically merges subcommand names with `ValidArgsFunction` output for the
root, so `rune <TAB>` suggests `serve`/`version`/`completion`/`help` **and** the
Runefile's task names with descriptions (shown in zsh/fish). The completion function
must never write to stdout (Cobra owns that protocol); use the returned slice only.

**Rationale**: Reuses existing filtering logic (`IsPrivate`, `osMatches`, `firstLine`),
keeps `internal/cli` framework-free, and satisfies FR-009..FR-012 + SC-003.

**Alternatives considered**: names without descriptions (rejected at clarification);
running the full pipeline for completion (rejected — slow, and analyzer errors would
suppress otherwise-valid names).

---

## Decision 4 — Shell completion generation via Cobra's built-in command

**Decision**: Delete the custom `genCompletion` / `cmd/rune/completion.go`. Rely on Cobra's
auto-added `completion` command (`rune completion {bash|zsh|fish|powershell}`), which
ships per-shell install instructions in its help and uses the hidden `__complete`
protocol to drive dynamic completion. Keep descriptions enabled (default); do not set
`DisableDescriptions`. Leave `CompletionOptions` at defaults (the command is visible and
discoverable, satisfying FR-009/FR-001).

**Rationale**: Idiomatic, less code, supports all four shells, and the help text already
documents installation — directly satisfies FR-009/FR-010 and SC-003.

**Alternatives considered**: keep the custom wrapper (rejected: re-implements Cobra,
doesn't do dynamic args, no install help).

---

## Decision 5 — `serve` flags, validation, and the `mcp` alias

**Decision**: `serve` defines real flags bound to local variables:
`--http` (bool), `--addr` (string), `--token-file` (string), `--mcp` (bool, accepted for
clarity / no-op as today). `RunE` validates and calls the unchanged
`cli.ServeMCP(opts, useHTTP, addr, tokenFile)`. Validation: `--addr`/`--token-file`
without `--http` is a usage error (returned as a `cli.UsageError` → exit 2; FR-015).
Register flag completion: `--token-file` → default file completion; (no enum flags
today). `Aliases: ["mcp"]` makes `rune mcp` route to `serve` with stdio defaults
(no `--http`), preserving today's `mcp` = stdio shorthand and satisfying FR-023.

**Rationale**: Replaces the hand-rolled arg loop in `runServe` with validated, helpable,
completable flags while preserving `ServeMCP` semantics (Constitution VII).

**Alternatives considered**: separate `serve` and `mcp` commands (rejected at
clarification — `serve` primary, `mcp` alias); `MarkFlagsMutuallyExclusive` (doesn't fit
— the relation is conditional-requires, not mutual exclusion, so validate in `RunE`).

---

## Decision 6 — "Did you mean" for mistyped commands (FR-016)

**Decision**: Cobra's built-in suggestion only fires for unknown *subcommands*; because
unknown tokens are valid task names here, Cobra will not produce "unknown command". So
implement the suggestion in Rune's own `unknown task` path: `internal/cli/suggest.go`
provides a small Levenshtein `nearest(token, candidates)`; `splitArgs` builds candidates
from the Runefile's task names **plus** the reserved command names (injected via a new
`Options.Commands []string`, set by `main`) and appends `… (did you mean "serve"?)` to
the existing `unknown task: <tok>` usage error. Threshold: suggest when edit distance
≤ 2 (or ≤ ⌊len/3⌋), pick the single closest.

**Rationale**: Keeps the suggestion useful for both task typos and command typos without
importing Cobra into `internal/cli` (its `ld` helper is unexported anyway). Concise error
+ `SilenceUsage` already prevents full-usage dumps (FR-017). Exit code stays 2.

**Alternatives considered**: vendoring Cobra's unexported distance fn (not exported);
turning task typos into Cobra unknown-command errors (impossible — tasks are dynamic).

---

## Decision 7 — Error/usage discipline, context, and the error banner

**Decision**: Set `SilenceUsage`/`SilenceErrors` on root (Cobra propagates to children
during execute). Use `root.ExecuteContext(ctx)` so the signal-cancellable context flows
to `RunE` and `ValidArgsFunction` via `cmd.Context()` (still also set `opts.Ctx`). Keep
`main.run()`'s terminal error mapping unchanged: print `rune: <err>` to stderr unless the
error is a `*cli.ValidationError` (diagnostics already rendered) or a silent
`*cli.TaskFailure` ([no-exit-message]); exit via `cli.CodeFor(err)`. Subcommand errors
(e.g. `serve`) flow through the same `CodeFor` mapping (default → exit 2 / usage).

**Rationale**: Preserves FR-017..FR-022 and the exact exit-code contract in
`internal/cli/exit.go` (`0/1/2/3/130`). No `os.Exit` inside any `RunE` (golang-cli skill;
Cobra cleanup must run).

**Alternatives considered**: per-command `SilenceUsage` (redundant); `os.Exit` in `RunE`
(rejected — skips deferred cleanup, violates the skill's Common Mistakes table).

---

## Resolved unknowns summary

| Topic | Resolution |
|-------|-----------|
| Reserved-word precedence vs. dynamic tasks | Cobra first-non-flag-token routing; built-ins win; `--` escapes (D1, D2) |
| Keep task arg pass-through? | Yes — `SetInterspersed(false)` retained; doesn't affect routing (D1) |
| Completion data without framework coupling | `cli.TaskCandidates` (cobra-free) → mapped to `cobra.Completion` in `main` (D3) |
| Generate completion scripts | Cobra auto `completion` command; delete custom wrapper (D4) |
| serve flag validation | conditional-requires checked in `RunE` → `UsageError` (D5) |
| did-you-mean for dynamic tasks | own Levenshtein over tasks + reserved names (D6) |
| context + exit codes | `ExecuteContext`; unchanged `CodeFor` mapping (D7) |

No unresolved `NEEDS CLARIFICATION` remain. Ready for Phase 1.
