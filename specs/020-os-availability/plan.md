# Implementation Plan: OS Availability Enforcement

**Branch**: `020-os-availability` | **Date**: 2026-08-12 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/020-os-availability/spec.md`

## Summary

The OS attributes (`[linux]`, `[macos]`, `[windows]`, `[unix]`) already
filter `--list`, the picker, and completion via the unexported `osMatches`
helper in `internal/cli/run.go`, but the MCP server exposes OS-mismatched
tasks as tools and nothing stops them from running. Fix: promote the
predicate to the task model as `(*ast.Task).AvailableOn(goos)` — the single
authoritative rule (FR-001) — then apply it at the three missing surfaces:
the MCP adapter's `Tasks()`/`Call()` (FR-002), root resolution in
`resolveRoots` with a `ValidationError` (exit 3, nothing executed, FR-003),
and dependency/post-hook resolution in the scheduler via a new
`Engine.Available` method that silently skips mismatched targets (FR-004).
`--dump --format json` gains a computed `available` field (FR-005). No new
syntax, no new dependencies, no parser/lexer/formatter changes.

## Technical Context

**Language/Version**: Go 1.25 (`go.mod`); standard library only — the
feature is pure predicate plumbing over the existing AST

**Primary Dependencies**: existing internal packages only: `internal/ast`
(new method), `internal/runtime/scheduler` (interface extension),
`internal/cli` (run/mcp/dump call sites); public `mcpserver/` package
unchanged (its `Engine` interface already receives a filtered task list)

**Storage**: N/A

**Testing**: `go test` inside the Docker harness (`docker-compose run --rm
test go test ./...`); new unit tests in `internal/ast` (first tests for
this package) and `internal/runtime/scheduler`; CLI-level tests in
`internal/cli`; docs harness (`test/docs`) re-verifies the updated example

**Target Platform**: Linux, macOS, Windows (single static binary,
`CGO_ENABLED=0`); every OS path unit-testable from any host because the
predicate takes the target OS as a parameter (FR-007)

**Project Type**: CLI + embeddable MCP server

**Performance Goals**: availability checks are O(#attributes) slice scans
at resolution time — no measurable impact

**Constraints**: no new Runefile syntax or flags (NFR-001); existing
Runefiles parse and format byte-identically; `--list`/picker/completion
output unchanged for identical inputs (SC-005); error message follows the
existing `ValidationError` diagnostic style

**Scale/Scope**: ~1 method moved + ~5 call sites + 1 interface method;
~6 files touched in `internal/`, 3 docs files, plus tests. No changes to
parser, lexer, formatter, LSP, or VS Code grammar (attribute set is
unchanged)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| # | Principle | Verdict | Notes |
|---|-----------|---------|-------|
| I | Command Runner, Not a Build System | PASS | No caching/skip-by-freshness semantics. The dependency skip is a declared platform constraint, not an implicit "work is unnecessary" decision; `--dry-run`/`--summary` plans reflect it visibly. |
| II | Errors Are a Feature | PASS | Direct invocation of an unavailable task fails as a `ValidationError` (exit 3) **before any execution**, naming the task, its required OS(es), and the host OS. |
| III | Minimal, Total DSL | PASS | Zero grammar/DSL surface change; the four OS attributes already exist. |
| IV | Hand-Written Front End, Idiomatic Go | PASS | Predicate joins `IsPrivate()` on `ast.Task` (the layout's natural home); scheduler `Engine` interface grows one method; `mcpserver/` public API untouched. |
| V | Boringly Portable | PASS | Pure Go string comparison on GOOS names; `unix` = everything except Windows, tested for all three CI platforms. |
| VI | Test-First, Multi-Layer Verification | PASS | Red-green order in tasks.md: predicate table tests, scheduler skip tests, CLI/MCP/dump tests, docs harness re-run. `[linux]`-style fixtures get their first-ever coverage. |
| VII | AI-Native, Secure by Default | PASS (strengthened) | This feature exists to make the agent surface truthful: agents cannot see or invoke platform-incompatible tasks. |
| VIII | Go Engineering Discipline | PASS | Errors wrapped per house style; no goroutines added; no globals; golangci-lint clean required. |

**Engineering Constraints check**: Docker-only testing — respected. Locked
package layout — respected (no new packages). **Backward compatibility —
DEVIATION**, justified in Complexity Tracking below. Surface changes carry
docs — no DSL surface change; behavior docs (`docs/runefile.md`,
`docs/mcp.md`, os-filtering example) updated in the same PR.

**Post-Phase-1 re-check**: design artifacts introduce no new packages, no
DSL changes, no additional deviations. Gate still passes.

## Project Structure

### Documentation (this feature)

```text
specs/020-os-availability/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/
│   ├── availability.md  # Predicate + CLI + scheduler behavior contract
│   ├── mcp-surface.md   # MCP tool-list / tool-call contract
│   └── dump-schema.md   # --dump --format json field addition
├── checklists/
│   └── requirements.md  # Spec quality checklist (done)
└── tasks.md             # Phase 2 output (/speckit-tasks — NOT created here)
```

### Source Code (repository root)

```text
internal/
├── ast/
│   ├── ast.go                    # + (*Task).AvailableOn(goos), (*Task).OSFilters()
│   └── availability_test.go      # NEW — first tests in this package (SC-004)
├── runtime/scheduler/
│   ├── scheduler.go              # Engine interface + Available(*ast.Task) bool;
│   │                             #   runDep/runDepsParallel skip unavailable targets
│   └── scheduler_test.go         # + skip semantics tests (deps, post-hooks, memo)
└── cli/
    ├── run.go                    # osMatches deleted; call sites → AvailableOn;
    │                             #   resolveRoots availability gate (ValidationError);
    │                             #   engine gets goos field (injectable, default runtime.GOOS)
    ├── mcp.go                    # Tasks() filter + Call() defense-in-depth check
    ├── dump.go                   # taskDTO.Available; toDTO takes goos
    ├── complete.go               # call site → AvailableOn
    └── *_test.go                 # CLI-level: direct-invoke error, dispatch pattern,
                                  #   --list regression, dump field, MCP adapter filter

docs/
├── runefile.md                   # attributes table: document enforcement + skip
├── mcp.md                        # "non-private AND host-OS-available" exposure rule
└── examples/os-filtering/
    ├── Runefile                  # + dispatch-pattern task (deps on per-OS tasks)
    └── README.md                 # "shows and runs" claim becomes true; describe skip
```

**Structure Decision**: No new packages. The predicate moves to
`internal/ast` (joining `IsPrivate`, the existing task-classification
helper), the skip lives in `internal/runtime/scheduler` where dependency
traversal already happens, and `internal/cli` keeps only call sites. The
public `mcpserver/` package needs no change because its `Engine.Tasks()`
contract already means "the tasks to expose" — the CLI adapter simply stops
offering unavailable ones.

## Complexity Tracking

> Fill ONLY if Constitution Check has violations that must be justified

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Backward compatibility constraint ("existing Runefiles keep working" / breaking changes opt-in per file): an OS-mismatched dependency that previously **ran** is now skipped, and direct invocation of a mismatched task now errors, with no per-file opt-in. | The old behavior is a defect, not a contract: `docs/examples/os-filtering/README.md` already promised the task "only shows **and runs** on that OS", and leaving execution open defeats the feature's purpose (agents or scripts invoking by name still run incompatible commands). Spec records this as an intentional defect fix (Assumptions; Edge Cases). | A per-file opt-in (e.g. a `set enforce-os` setting) would leave every existing Runefile agent-unsafe by default, contradict the already-published docs, and add a permanent setting for behavior nobody should opt out of. A `--ignore-os` escape hatch was explicitly deferred by the spec (Assumptions). Changelog carries a pre-1.0 behavior-change note instead. |
