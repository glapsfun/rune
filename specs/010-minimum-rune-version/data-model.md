# Phase 1 Data Model: Minimum Rune Version

This feature introduces no persisted storage. The "data model" is the set of in-memory
values and the validation rules that govern them, derived from the spec's Key Entities and
Functional Requirements.

## Entities

### Version (`internal/semver.Version`)

A parsed Semantic Version used for both the installed binary version and the declared
requirement.

| Field | Type | Notes |
|-------|------|-------|
| Major | int | ≥ 0 |
| Minor | int | ≥ 0 |
| Patch | int | ≥ 0 |
| Prerelease | []string | dot-separated identifiers; empty for a release |
| Build | []string | build metadata; **ignored** for precedence |

**Behavior**:
- `Parse(s string) (Version, error)` — accepts `MAJOR.MINOR.PATCH[-pre][+build]`; rejects
  malformed input with an error (no leading `v`).
- `Compare(other Version) int` — SemVer 2.0.0 precedence: numeric core first; a prerelease
  ranks below the equal release; prerelease identifiers compared field-by-field (numeric <
  alphanumeric); build metadata ignored.
- `Satisfies(min Version) bool` — `v.Compare(min) >= 0`.

**Validation rules**:
- `0.9.0-rc.1 < 0.9.0` (FR-010).
- `0.9.0 == 0.9.0` regardless of build metadata (`0.9.0+abc` equal precedence to `0.9.0`).
- `0.9.0-dev+13dbf54` is a prerelease of `0.9.0` (build metadata ignored) — does not satisfy
  a `0.9.0` requirement (dev builds handled via `--ignore-version`, not comparator special-casing).

### MinimumRequirement (derived, not a stored struct)

The requirement declared by the **root** Runefile.

| Attribute | Source | Notes |
|-----------|--------|-------|
| RawValue | `(*ast.StringLit).Value` of the `minimum_version` setting | must be a static literal |
| Parsed | `semver.Version` | from `RawValue` |
| Span | `(*ast.StringLit).Sp` | caret target for diagnostics |
| Origin | root file only | evaluated pre-`Compose` so imports cannot contribute |

**Validation rules** (order matters):
1. If no `minimum_version` setting on the root file → no requirement; gate is a no-op (FR-023, FR-022).
2. If present but `s.Value` is not `*ast.StringLit` → error `minimum_version must be a static
   semantic version`, caret at `s.Value.Span()` (FR-007).
3. If a string literal that fails `semver.Parse` → error "not a valid semantic version",
   caret at `Sp` (FR-008).
4. Range/compound syntax (contains operators/commas/`^`/`~`) → rejected as not a valid single
   semantic version (FR-009).

### CompatibilityResult (for `rune version --check`)

| Field | Type | JSON key |
|-------|------|----------|
| Installed | string | `installed` |
| Required | string (or empty/absent when none) | `required` |
| Compatible | bool | `compatible` |
| RunefilePath | string | `runefile` |

**Rules**:
- `Compatible = installed.Satisfies(required)` when a requirement exists.
- When no requirement/no Runefile: report "no requirement declared"; `--json` sets
  `required` empty and `compatible` true (nothing to violate) — see contract for exact shape.
- Incompatible ⇒ non-zero exit (FR-020), no task executed.

## Configuration inputs (not persisted)

| Input | Where | Default | Purpose |
|-------|-------|---------|---------|
| `Options.IgnoreVersion` | `internal/cli` (set by global `--ignore-version`) | false | CLI break-glass override (FR-014) |
| `Options.Version` (installed) | threaded from `main.version` | build ldflag | version compared against requirement |
| `RUNE_TEST_VERSION` | `cmd/rune` layer, test-only | unset | integration test injection of installed version (FR-011) |
| `mcpserver.Options.AllowIgnoreVersion` | `mcpserver` | false | operator-only override on the agent/MCP path (FR-017) |

## State transitions

There are no long-lived state machines. The per-invocation gate flow:

```text
Parse root Runefile
      │
      ▼
MinimumVersion(root settings) ── none ──▶ continue (no gate)
      │ present
      ▼
static literal? ── no ──▶ error: must be a static semantic version (exit 3)
      │ yes
      ▼
valid semver?  ── no ──▶ error: not a valid semantic version (exit 3)
      │ yes
      ▼
installed.Satisfies(required)? ── yes ──▶ continue (Compose → Analyze → run)
      │ no
      ├── --ignore-version (CLI) / AllowIgnoreVersion (MCP) ──▶ warn + continue
      └── otherwise ──▶ incompatibility error w/ caret (exit 3), nothing executed
```
