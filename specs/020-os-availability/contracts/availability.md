# Contract: Availability Predicate, CLI, and Scheduler Behavior

**Feature**: 020-os-availability

## Predicate (Go API, `internal/ast`)

```go
// AvailableOn reports whether the task may run on the given GOOS.
// No OS attribute ⇒ true. Multiple OS attributes OR together.
// "unix" matches every goos except "windows"; "macos" matches "darwin".
func (t *Task) AvailableOn(goos string) bool

// OSFilters returns the task's declared OS attribute kinds in source
// order (e.g. ["linux", "windows"]), or an empty slice when unrestricted.
func (t *Task) OSFilters() []string
```

Guarantees:

- Pure functions of the task's attributes and the `goos` argument; no
  reads of `runtime.GOOS` or other process state (FR-007).
- Semantics identical to the deleted `internal/cli.osMatches` (truth table
  in [data-model.md](../data-model.md)).

## CLI: direct invocation (root gate)

| Aspect | Contract |
|--------|----------|
| Scope | Every task named on the command line, in all plan modes (`run`, `--dry-run`, `--summary`) |
| Check point | Root resolution, before the scheduler starts and before any confirmation prompt, cache check, or body execution |
| On mismatch | Entire invocation fails: nothing executes, even for other (available) tasks named in the same invocation |
| Error class | `ValidationError` → exit code **3** (`ExitValidation`) |
| Message shape | `task "NAME" is not available on HOST (requires F1 or F2)` where `HOST` is the host GOOS name and `F1..Fn` are the task's OS filters in source order |
| Stream | stderr, styled like existing validation diagnostics |

## Scheduler: dependency & post-hook skip

| Aspect | Contract |
|--------|----------|
| Scope | Every dependency and post-hook edge, serial or `[parallel]`, at any depth |
| On mismatch | Edge is skipped: target's deps/body/hooks never run, no error, no output in default (non-verbose) mode |
| Ordering | Remaining deps and post-hooks keep their declared order and semantics |
| Memoization | A skipped target gets no memo entry; if the same task is later reached on a matching edge (impossible for OS within one process, but contract-relevant for future predicates) it would evaluate independently |
| All-deps-skipped | The depending task's own body still runs (Story 3 scenario 2) |
| Direct-vs-dep | The SAME task errors when invoked directly but skips when reached as a dep — both behaviors must hold in one binary (spec Edge Cases) |

## Unchanged surfaces (regression contract)

- `--list`, bare `rune` overview, interactive picker, and shell completion
  produce byte-identical output for identical inputs.
- Parsing, formatting (`rune fmt`), `docs/GRAMMAR.md`, LSP completion/hover,
  and the VS Code grammar are untouched.
- Private-task semantics (hidden but dep-callable and directly runnable)
  are unchanged.
