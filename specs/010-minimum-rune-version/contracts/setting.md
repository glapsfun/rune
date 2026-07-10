# Contract: `minimum_version` Setting

## Syntax

```rune
set minimum_version := "0.8.0"
```

- Consumed as an ordinary `set NAME := <string-literal>` setting; **no new grammar**.
- The value MUST be a static string literal (`*ast.StringLit`).
- The literal MUST be a single valid Semantic Version (`MAJOR.MINOR.PATCH[-pre][+build]`),
  no leading `v`, no operators, no ranges.
- Meaning: the Runefile requires the installed Rune binary to be **≥** this version.

## Independent from `rune_version`

`minimum_version` (binary release requirement) and `rune_version` (Runefile language
compatibility) are separate settings and do not interact. Both may appear:

```rune
set rune_version := "1"
set minimum_version := "0.8.0"
```

## Ownership

- Only the **root** Runefile's `minimum_version` is effective.
- A `minimum_version` declared in an imported/child file (or a `mod`) MUST NOT set, override,
  or relax the effective requirement. (Guaranteed by evaluating the setting before import
  splicing.)

## Accepted / Rejected

| Input | Result |
|-------|--------|
| `set minimum_version := "0.8.0"` | ✅ accepted |
| `set minimum_version := "1.0.0-rc.1"` | ✅ accepted (valid prerelease requirement) |
| (no setting) | ✅ no gate applied; existing behavior unchanged |
| `set minimum_version := required` (var/expr) | ❌ `minimum_version must be a static semantic version` |
| `set minimum_version := env("X")` | ❌ `minimum_version must be a static semantic version` |
| `set minimum_version := "0.8"` / `"latest"` / `"v0.8.0"` | ❌ not a valid semantic version |
| `set minimum_version := ">=0.8,<1.0"` | ❌ not a valid semantic version (ranges deferred, FR-009) |

## Comparison semantics (SemVer 2.0.0)

- Compare `MAJOR`, then `MINOR`, then `PATCH` numerically.
- A prerelease ranks **below** the equal release: `0.9.0-rc.1` does NOT satisfy `0.9.0`.
- Build metadata is ignored for precedence: `0.9.0+abc` has equal precedence to `0.9.0`.
- Installed satisfies requirement iff `installed ≥ required`.
