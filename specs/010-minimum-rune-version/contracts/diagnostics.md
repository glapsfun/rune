# Contract: Diagnostics

All diagnostics are produced before any execution, carry a `file:line:col` location with a
caret span pointing at the offending value, and result in a non-zero exit
(`ExitValidation`, 3) with zero side effects. Exact bytes are pinned by a golden fixture.

## Incompatible version

Given `set minimum_version := "0.8.0"` and installed `0.7.2`:

```text
error: this Runefile requires Rune >= 0.8.0

  Runefile:2:24
    |
  2 | set minimum_version := "0.8.0"
    |                        ^^^^^^^
    |
    = installed version: 0.7.2
    = required version:  0.8.0
    = upgrade: https://github.com/glapsfun/rune/releases

nothing was executed
```

- Caret spans the value literal (`"0.8.0"`).
- Includes installed version, required version, and the upgrade URL.
- Trailer: `nothing was executed`.

> Rendering note: the exact placement of the `= …` note lines / trailer is a renderer
> detail; the golden fixture is the source of truth. The required/installed/upgrade
> information and the caret-at-value behavior are the contract.

## Non-static value

Given a non-literal value:

```rune
required := env("RUNE_VERSION")
set minimum_version := required
```

```text
error: minimum_version must be a static semantic version
```

- Caret spans the offending value expression.

## Invalid semantic version

Given `set minimum_version := "0.8"` (or `"latest"`, `"v0.8.0"`, a range):

```text
error: minimum_version must be a static semantic version
```
(or a specific "not a valid semantic version" message — final wording pinned by golden).

- Caret spans the value literal.

## Override warning (not an error)

Given `--ignore-version` with an unmet requirement (installed `0.7.2`, required `0.8.0`):

```text
warning: ignoring Runefile minimum Rune version 0.8.0; running 0.7.2
```

- Written to stderr; execution proceeds; exit code reflects the task, not the gate.
