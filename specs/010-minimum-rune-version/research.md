# Phase 0 Research: Minimum Rune Version

All Technical Context unknowns are resolved below. Each item records the decision, the
rationale, and the alternatives considered. Findings are grounded in the existing codebase
map (see references to `file:line`).

## R1. SemVer comparison strategy

**Decision**: Implement a small, dependency-free `internal/semver` package that parses
`MAJOR.MINOR.PATCH[-prerelease][+build]` and compares two versions by SemVer 2.0.0
precedence. Expose at minimum: `Parse(string) (Version, error)` and
`Version.Compare(other) int` (or a `Satisfies(min)` helper for the `>=` case).

**Rationale**:
- Principle V (Boringly Portable) values zero/minimal dependencies; `go.mod` currently has
  no SemVer library.
- Rune has a *Rune-specific* dev-build format (`0.9.0-dev+commit`) and an explicit rule
  that prerelease `0.9.0-rc.1` does **not** satisfy `0.9.0`. Owning the comparator lets us
  encode these rules precisely and test them, rather than fighting a general library's
  edge behavior.
- The needed surface is tiny (parse + precedence), well specified by SemVer 2.0.0, and
  fuzz/table testable.

**Alternatives considered**:
- `github.com/Masterminds/semver/v3` — full constraint language we explicitly do not want
  yet (FR-009); adds a dependency for a fraction of its surface.
- `golang.org/x/mod/semver` — requires a leading `v`, treats invalid input as lowest, and
  its API is string-based; would need adapting and still pulls a dependency. Reserve for
  reconsideration only if the internal comparator proves insufficient.

**Comparator rules to implement (SemVer 2.0.0)**:
- Compare major, then minor, then patch numerically.
- A version with a prerelease has **lower** precedence than the same version without one
  (so `0.9.0-rc.1 < 0.9.0`), satisfying FR-010.
- Prerelease identifiers compared per spec (numeric < alphanumeric, field-by-field).
- **Build metadata is ignored** for precedence (`+13dbf54` does not affect comparison).
- `Satisfies(min)`: installed satisfies requirement iff `installed.Compare(min) >= 0`.

## R2. Installed-version source and test injection (FR-011)

**Decision**: The gate helper `config.CheckMinimumVersion(file, installed string)` takes the
installed version as an explicit argument — no ambient global read inside the helper. The
CLI passes `opts.Version` (already threaded from `main.version` via
`cmd/rune/root.go:59` → `internal/cli/dispatch.go:44`). For integration tests that run the
built binary, add a **test-only environment hook** read once at the `cmd/rune` layer (e.g.
`RUNE_TEST_VERSION`) that, when set, overrides the reported installed version before it is
threaded into `Options`. This env hook is documented as test-only and is never consulted
inside `internal/`.

**Rationale**:
- Keeps the comparison logic pure and unit-testable with explicit inputs (Principle VI/VIII).
- The integration harness builds with default ldflags, so the binary reports `version =
  "dev"` (`cmd/rune/main.go:18`); without an injection hook there is no way to simulate an
  older/newer installed release end-to-end. Threading via env at the outermost layer keeps
  `internal/` free of test-only branches.

**Alternatives considered**:
- ldflags per test build — slow, complicates the single-build harness (`test/integration/
  harness_test.go:21-48`).
- A hidden `--installed-version` flag — leaks a test concern into the public CLI surface.

**Dev-build semantics**: A dev binary reports e.g. `0.9.0-dev+13dbf54`. Under R1 rules the
`+13dbf54` build metadata is ignored and `-dev` is a prerelease, so a dev build's base
release is `0.9.0` but it is treated as a prerelease of `0.9.0` (lower than the release).
The spec's development-version rule is honored: releases satisfy their own version;
prereleases (incl. `-dev`) do not satisfy the equal release. Whether a `-dev` build should
be treated leniently for *local* development is handled by `--ignore-version`, not by
special-casing the comparator.

## R3. Gate insertion points (FR-004, FR-012)

**Decision**: Call the gate in both static-load pipelines immediately **after**
`parser.Parse` and **before** `config.Compose`:
- `internal/cli/run.go` between line 57 (`parser.Parse`) and line 61 (`config.Compose`).
- `internal/cli/serve.go` `loadModule` between line 40 (`parser.Parse`) and line 42
  (`config.Compose`).

**Rationale**:
- At this point `file.Settings` contains **only the root file's** `set` statements; imports
  have not been spliced. `config.Compose` merges imported settings only into gaps
  (`internal/config/compose.go:77-87`), so evaluating `minimum_version` pre-`Compose`
  structurally guarantees a child cannot inject or relax the requirement — this *is* the
  enforcement mechanism for FR-012 ("root owns the requirement"). No extra bookkeeping.
- It is before analysis (`analyzer.Analyze`) and before any engine/scheduler construction,
  satisfying FR-004 (before imports, analysis, and execution) with zero side effects.
- The same insertion covers MCP/agent runs because `serve.go loadModule` is the shared
  static-load path for `rune serve`/MCP and the agent callback server.

**Alternatives considered**:
- In `analyzer.Analyze` — runs *after* `Compose`, so a child's value could already be
  present, breaking FR-012; also later than required. Rejected.
- In `config.ResolveSettings` — execution-prep time, far too late. Rejected.

## R4. Static-value guard (FR-007, FR-008)

**Decision**: `config.MinimumVersion(f)` scans root `f.Settings` for `minimum_version` and
requires `s.Value` to be a `*ast.StringLit` (mirroring `RuneVersion` at
`internal/config/version.go:13-22`). If present but not a string literal →
`minimum_version must be a static semantic version`, caret at `s.Value.Span()`. If a string
literal that fails `semver.Parse` → a "not a valid semantic version" diagnostic, caret at
the `*ast.StringLit.Sp`. Both are `diag` errors surfaced through the existing
`renderDiags` path with a non-zero (`ExitValidation`, 3) exit.

**Rationale**: The AST already carries per-value spans (`internal/ast/ast.go:200-206`), so
caret placement on the value literal (FR-006) is free. Reusing the `rune_version` literal
check keeps behavior consistent between the two settings.

## R5. Diagnostic rendering (FR-006)

**Decision**: Build the incompatibility error with `diag.New(stringLit.Sp, message)` and
attach the required/installed/upgrade lines. Use the existing renderer
(`internal/diag/render.go`) via `renderDiags` (`internal/cli/run.go:754`). The message and
note lines follow the spec's example (`this Runefile requires Rune >= X`, `installed
version`, `required version`, `upgrade: <URL>`).

**Open detail for design**: The `diag.Diagnostic` model is `{Severity, Span, Message}`
(`internal/diag/diagnostic.go:24-28`) with no structured "notes" field; the renderer prints
`file:line:col: error: message` + snippet. The spec's example shows extra `= ` note lines.
**Decision**: encode the required/installed/upgrade lines into the message (multi-line) or
add a minimal notes affordance — pick the lowest-churn option during implementation;
golden fixture will pin the exact bytes. This is a rendering-detail choice, not a
requirement risk.

**Upgrade URL**: Use `https://github.com/glapsfun/rune/releases` (per spec example and the
repo-badge target `glapsfun/rune` recorded in the docs feature). Centralize as a constant.

## R6. `--ignore-version` override (FR-014–FR-017)

**Decision**:
- Add a persistent/global bool flag `--ignore-version` on the root command
  (`cmd/rune/root.go:81-95`), backed by a new `Options.IgnoreVersion` field
  (`internal/cli/dispatch.go`). Because `SetInterspersed(false)` is set
  (`root.go:79`), it must precede the task name — consistent with all other globals.
- When set and the requirement would fail, print a **warning** (`warning: ignoring Runefile
  minimum Rune version X; running Y`) to stderr and proceed. Use `diag.Warn`/styled stderr.
- The override is **never** readable from a Runefile (no setting maps to it) — FR-016 holds
  by construction.
- **MCP/agent path**: the gate in `serve.go loadModule` does **not** consult a CLI flag;
  instead ignoring is gated by an explicit `mcpserver.Options.AllowIgnoreVersion` (default
  `false`), fed from a `rune serve` operator flag/env — mirroring how `AllowDestructive`
  flows from `--yes` (`internal/cli/serve.go:104`). Default behavior: incompatibility on the
  agent path refuses with the standard error (FR-017).

**Rationale**: Keeps the safety valve loud and CLI-only, and keeps agent execution safe by
default while still giving operators an explicit escape hatch.

## R7. `rune version` extensions (FR-018–FR-022)

**Decision**: Extend `cmd/rune/version.go`:
- Bare `rune version`: print installed Rune version and a language line, e.g.
  ```
  rune <version>
  runefile language 1
  ```
  Language version from `internal/config` `CurrentVersion` (`internal/config/version.go:10`).
- `rune version --check`: resolve the applicable Runefile via `config.Resolve` (as
  `dispatch.go` does), read `minimum_version`, print required/installed/status; exit
  non-zero when incompatible and run no task (FR-020). When no Runefile/requirement, report
  "no requirement declared" rather than incompatible (FR-022).
- `rune version --check --json`: emit `{ "installed", "required", "compatible",
  "runefile" }`, modeled on the DTO/`json.MarshalIndent` pattern in `internal/cli/dump.go`.

**Convention note**: Existing JSON output uses `--format json` (dump), but the spec
explicitly requests `--json` on `version`. **Decision**: add a local `--json` bool flag on
the `version` command (not global). Divergence is intentional and spec-driven; documented.

## R8. Exit codes

**Decision**: Reuse `ExitValidation` (3) for an incompatibility abort during a run
(static error, nothing executed — matches its meaning in `internal/cli/exit.go:8-14`) and
for `rune version --check` incompatibility. Static-value/semver-parse errors also map to
`ExitValidation`.

**Rationale**: Consistent with how the analyzer's static errors already exit; avoids
inventing a new code for the same "refused before execution" category.

## R9. Docs surface (Principle: surface changes carry their docs)

**Decision**: Update `docs/GRAMMAR.md` (settings list) and the settings/runefile docs
(`docs/runefile.md`) to document `minimum_version`, the `--ignore-version` override, and
`rune version --check`. Because the repo has a `docs-verify` gate and `test/docs` harness,
any code blocks added must be self-contained/valid; add the new setting to the docs tests as
needed. Ship docs in the same change as the code and fixtures.

## Summary of resolved decisions

| # | Topic | Decision |
|---|-------|----------|
| R1 | SemVer compare | New dependency-free `internal/semver`; SemVer 2.0.0 precedence, build metadata ignored, prerelease < release |
| R2 | Version injection | Explicit `installed` arg; test-only `RUNE_TEST_VERSION` env hook at `cmd/rune` layer |
| R3 | Gate location | Pre-`Compose` in `run.go` and `serve.go loadModule` (root-only settings ⇒ FR-012) |
| R4 | Static guard | `*ast.StringLit` required; else "must be a static semantic version"; invalid semver rejected |
| R5 | Diagnostics | `diag.New(stringLit.Sp, …)` via existing renderer; upgrade URL `github.com/glapsfun/rune/releases` |
| R6 | Override | CLI-only global `--ignore-version` (warn+proceed); MCP via `Options.AllowIgnoreVersion` default false |
| R7 | version cmd | `--check` and `--json`; language line from `config.CurrentVersion` |
| R8 | Exit code | `ExitValidation` (3) for incompatibility and static-value errors |
| R9 | Docs | GRAMMAR.md + runefile/settings docs + docs-verify fixtures in same change |

No NEEDS CLARIFICATION markers remain.
