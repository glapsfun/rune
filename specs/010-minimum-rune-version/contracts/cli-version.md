# Contract: CLI — version command & `--ignore-version`

## `rune version`

Prints the installed Rune version and the Runefile language version.

```text
$ rune version
rune version 0.8.3 (commit none)
runefile language 1
```

- Line 1: byte-identical to `rune --version` (cobra's default template:
  `rune version <installed-version> (commit <commit>)`).
- Line 2: `runefile language <CurrentVersion>` (from `internal/config`).
- Exit code: 0.

## `rune version --check`

Reports compatibility of the installed binary against the applicable Runefile's
`minimum_version`. Resolves the Runefile from the working directory (same resolution as a
normal run).

```text
$ rune version --check
Runefile requires >= 0.8.0
Installed Rune: 0.8.3
Status: compatible
```

- Exit 0 when compatible.
- Exit non-zero (`ExitValidation`, 3) when incompatible; **no task is executed**.
- When no Runefile or no `minimum_version` is present: report that no requirement is
  declared and exit 0 (NOT treated as incompatible).

## `rune version --check --json`

Machine-readable form for CI/scripts.

```json
{
  "installed": "0.8.3",
  "required": "0.8.0",
  "compatible": true,
  "development": false,
  "runefile": "/project/Runefile"
}
```

- Keys: `installed` (string), `required` (string; empty when none), `compatible` (bool),
  `development` (bool; true when the installed version is not a recognized semantic version —
  a local dev build — and the requirement was waved through), `runefile` (string; resolved
  absolute path, empty when none found).
- Exit code mirrors the non-JSON `--check` (non-zero when incompatible).
- Output is stable, indented JSON (modeled on the existing `--dump --format json` DTO style).

## Global flag: `--ignore-version`

```sh
rune --ignore-version build
```

- Bypasses the `minimum_version` gate for this invocation only.
- When the requirement would otherwise fail, prints a warning to **stderr** and proceeds:

  ```text
  warning: ignoring Runefile minimum Rune version 0.8.0; running 0.7.2
  ```

- Must precede the task name (global flags are non-interspersed, consistent with existing
  globals like `--dry-run`).
- CANNOT be enabled from within a Runefile — there is no setting that maps to it.
- Not applicable on the MCP/agent path (that path uses the operator option below).

## MCP / agent path override (operator-only)

- On the MCP/agent static-load path the gate is enforced by default; an incompatible
  requirement refuses with the standard incompatibility error.
- Ignoring is available only when the operator explicitly enables
  `mcpserver.Options.AllowIgnoreVersion` (default `false`), fed from a `rune serve`
  operator flag/env — never from a Runefile.
