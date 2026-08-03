# Contract: Styled Output Surfaces

Binding output contract for feature 014. Each clause is testable at the binary
level. "Plain" = `--color=never`, `NO_COLOR` set, or non-TTY stream under
`auto`. "Colored" = `--color=always`, or TTY under `auto`. ANSI = any `\x1b[`
escape sequence.

## C1 — Failure banner

- Colored stderr: `rune:` prefix rendered in the Error role (bold red); the
  message text carries no escape codes of its own.
- Plain: exactly `rune: <message>\n` — byte-identical to pre-014 output.
- Applies to: top-level banner (any error class that prints it), cycle error,
  watch per-iteration error (`rune: …`), watch error-channel banner
  (`rune: watch error: …`).
- Silent failures (`[no-exit-message]`, declined confirm) still print nothing.
- Exit codes unchanged in all modes.
- With a masked secret in the message: `***` appears, the secret value never
  appears, in both colored and plain modes.

## C2 — Warnings

- Every stderr warning uses the shape `warning: <text>` with the `warning`
  label in the Warning role when colored; label plain otherwise.
- Covered sites: cache-write warning (already compliant), min-version override
  warning (newly compliant). No other warning-shaped line may bypass the role.

## C3 — Status notices

| Line (plain form, frozen) | Colored treatment |
|---|---|
| `formatted: <path>` | label Success, path plain |
| `cleared: <path>` | label Success, path plain |
| `watching <dir> (Ctrl-C to stop)` | whole line Muted |
| `change detected, re-running…` | whole line Muted |
| `rune MCP server on http://<addr> (token required)` / `rune MCP server on stdio` | whole line Muted |
| `<prompt> [y/N] ` | prompt text Warning, ` [y/N] ` plain |

Plain forms are byte-frozen; only escape codes may differ between modes.

## C4 — Analyze text output (reformatted surface)

- `rune analyze` (no `--json`), stdout: diagnostics rendered by the shared
  diagnostic renderer — severity word colored by severity, `file:line:col`
  locator faint, caret span in Caret role — identical rendering (modulo
  stream) to the same Runefile failing under `rune <task>`.
- Summary line `<N> error(s), <M> warning(s)` (existing wording): severity
  words colored by their roles when the respective count > 0.
- New plain layout becomes the golden baseline in this feature and is frozen
  thereafter. `analyze --json` output is byte-identical to pre-014 in all
  modes. Exit codes 0/3/2 unchanged.

## C5 — Help (reformatted surface: subcommands)

- Every command's `--help` uses the grouped layout (description, Usage,
  Examples where defined, Flags; root additionally Tasks + Commands).
- Headings in Heading role when stdout is colored; body always plain.
- Piped/`NO_COLOR`/`--color=never` help contains zero ANSI.
- Root help plain bytes unchanged from 008 baseline; subcommand plain help
  changes once to the grouped layout, then frozen.
- `--help` exits 0; unknown-flag errors unchanged.

## C6 — `--list` alignment

- For any set of task names, the `#` doc column starts at the same visible
  column in colored and plain modes.
- Colored mode pads by display width (wide/CJK names align correctly).
- Plain mode output is byte-identical to pre-014 for all-ASCII names (the
  entire existing corpus).

## C7 — Machine surfaces (never ANSI, any mode incl. `--color=always`)

`--dump` (both formats), `analyze --json`, `version --json`, `completion
<shell>` scripts, LSP JSON-RPC stream, MCP tool-result content, agent
write-back capture. Zero `\x1b` bytes, all color modes.

## C8 — Color controls

- `--color auto|always|never` and `NO_COLOR` govern every surface above; no
  new flag, env var, or setting is introduced.
- Decision remains per-stream: e.g. stdout piped + stderr TTY under `auto`
  colors stderr banners while `--list` stays plain.
- Invalid `--color` value: unchanged usage error (exit 2) on normal paths;
  help/banner tolerate and fall back to `auto` (pre-existing help behavior,
  extended to the banner).

## C9 — Ordering, streams, exit codes

Message ordering, stdout/stderr assignment, and every exit code are identical
to pre-014 behavior across all surfaces and modes.
