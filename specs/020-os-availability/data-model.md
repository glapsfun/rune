# Data Model: OS Availability Enforcement

**Feature**: 020-os-availability | **Date**: 2026-08-12

No persistent data. The "model" is the availability semantics attached to
existing AST entities and the two machine-facing projections of it.

## Entities

### Task (existing — `internal/ast.Task`)

| Aspect | Detail |
|--------|--------|
| Relevant fields | `Attributes []*Attribute`, `Deps []*DepCall`, `PostHooks []*DepCall` |
| New derived property | **Availability**: `AvailableOn(goos string) bool` |
| New helper | `OSFilters() []string` — the task's declared OS attribute kinds, in source order, empty when unrestricted |

**Availability rule** (semantics preserved from today's `osMatches`):

1. Collect all attributes of kind `linux`, `macos`, `windows`, `unix`
   (across all attribute lines; `Attr()` is unsuitable — it returns only
   the first).
2. Empty set ⇒ available on every OS.
3. Otherwise available iff ANY filter matches (OR):
   - `linux` ⇔ goos == "linux"
   - `macos` ⇔ goos == "darwin"
   - `windows` ⇔ goos == "windows"
   - `unix` ⇔ goos != "windows"

**Derived truth table** (validation fixtures):

| Filters | linux | darwin | windows |
|---------------------|-------|--------|---------|
| (none) | ✓ | ✓ | ✓ |
| linux | ✓ | ✗ | ✗ |
| macos | ✗ | ✓ | ✗ |
| windows | ✗ | ✗ | ✓ |
| unix | ✓ | ✓ | ✗ |
| linux, windows | ✓ | ✗ | ✓ |
| unix, windows | ✓ | ✓ | ✓ |

**Interaction with privacy**: independent, composed with AND at every
listing surface — shown/exposed iff `!IsPrivate() && AvailableOn(host)`.
Execution differs: private tasks are callable as deps *and* directly;
unavailable tasks are skippable as deps but NOT directly callable.

### Scheduler Engine (existing interface — `internal/runtime/scheduler.Engine`)

| Aspect | Detail |
|--------|--------|
| New method | `Available(task *ast.Task) bool` |
| Semantics | Consulted in `runDep` after `ResolveDep`; `false` ⇒ dependency/post-hook is skipped (nil result), target never enters `run`, no memo entry, chain untouched |
| Sole implementation | `internal/cli.engine`, backed by `task.AvailableOn(e.goos)`; `goos` field defaults to `runtime.GOOS`, injectable in tests |

**State transitions** (dependency resolution, per dep/post-hook):

```text
ResolveDep ok? ──no──► error (unknown dep, arity, eval) — unchanged
     │yes
Available?    ──no──► SKIP (return nil; nothing recorded)     ← NEW
     │yes
run(target)   ─────► memoized execute (deps → body → hooks) — unchanged
```

### MCP TaskInfo projection (existing — `mcpserver.TaskInfo`)

Struct unchanged. Population rule in `mcpAdapter.Tasks()` changes from
`!IsPrivate()` to `!IsPrivate() && AvailableOn(host)`. `mcpAdapter.Call()`
re-validates availability before scheduling (defense-in-depth).

### Dump projection (existing — `internal/cli.taskDTO`)

| Field | Type | JSON | Rule |
|-------|------|------|------|
| `Available` (new) | `bool` | `available` (always emitted) | `task.AvailableOn(goos)` at dump time; all tasks remain listed |

Existing `Attributes []string` keeps carrying the raw OS attribute names,
so consumers see both rule and verdict.

## Validation rules (from requirements)

- FR-001: every surface computes availability ONLY via `AvailableOn` — no
  second implementation may exist after `osMatches` is deleted.
- FR-003: root-gate error must name task, required OS(es) (`OSFilters`
  joined with " or "), and host OS; `ValidationError` ⇒ exit 3.
- FR-004: skip is silent, error-free, and order-preserving for remaining
  deps and hooks.
- FR-007: `AvailableOn`/`OSFilters` take no implicit host dependency; all
  host defaults live at call sites.
