# Contract: `--color` flag & color decision

## Flag

```
--color string   when to colorize output: auto|always|never (default "auto")
```

- Global persistent flag on the root command; available to every invocation.
- Default `auto` reproduces the pre-feature behavior exactly (FR-007).
- Invalid value (e.g. `--color=sometimes`) → clear error to stderr, **non-zero
  exit, no task execution** (FR-009).

## Per-stream resolution (FR-004, FR-005, FR-008)

`resolve(mode, stream)` returns whether to emit ANSI on that stream. Evaluated
independently for stdout and stderr. Precedence, highest first:

| # | Condition | Result |
|---|-----------|--------|
| 1 | `--color=never` | OFF |
| 2 | `--color=always` | ON |
| 3 | `NO_COLOR` set (non-empty) | OFF |
| 4 | global color disabled (`fatih/color.NoColor`) | OFF |
| 5 | otherwise | `isatty(stream)` |

Not consulted: `FORCE_COLOR`, `CLICOLOR`, `CLICOLOR_FORCE`.

## Behavior matrix (acceptance-aligned)

| Invocation | stdout (`--list`/`--help`) | stderr (status/diag) |
|------------|----------------------------|----------------------|
| TTY, no flags, `NO_COLOR` unset | color | color |
| piped stdout, stderr=TTY, `auto` | plain | color |
| `NO_COLOR=1` (any stream) | plain | plain |
| `--color=never` on a TTY | plain | plain |
| `--color=always` through a pipe | color | color |
| `--color=sometimes` | — error, exit ≠ 0, nothing runs — | |

## Invariance guarantee (FR-010, SC-001)

For every command **except `--help`/usage**, when both streams resolve OFF, the
emitted stdout and stderr bytes are identical to the pre-feature release — no
ANSI, no whitespace/column/order changes. `--help` is exempt: its redesigned
plain form is the new reviewed baseline.
