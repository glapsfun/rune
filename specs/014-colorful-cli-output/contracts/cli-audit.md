# CLI Command & Flag Audit (FR-012)

Audit of every command, flag, and argument as of branch point `013-secret-masking`
(2026-07-22). Each finding is marked **fix (014)** — addressed by this feature —
or **defer** — recorded, out of scope. Re-verify this table during
implementation review (SC-006).

## Command inventory

| Command | Args | Short description present | Long/Example present | Help styled today |
|---|---|---|---|---|
| `rune` (root) | `[VAR=VALUE …] [TASK [ARGS…]]` | yes | custom grouped help | yes (root only) |
| `serve` (alias `mcp`) | — | yes | Long yes; Example: **missing** | no |
| `version` | — | yes | Long yes; Example: **missing** | no |
| `completion` | `[bash\|zsh\|fish\|powershell]` | yes | Long yes; Example: **missing** | no |
| `analyze` | `[path]` | yes | Long + Example yes | no |
| `lsp` | — | yes | Long yes; Example: **missing** | no |
| `help` (Cobra auto) | `[command]` | Cobra default | Cobra default | no |

## Global flags (root)

| Flag | Description present | Wording notes |
|---|---|---|
| `-f, --file` | yes | ok |
| `--list` | yes | ok |
| `--dry-run` | yes | ok |
| `--summary` | yes | ok |
| `--dump` | yes | ok |
| `--format` | yes | only meaningful with `--dump` — help text says so: ok |
| `--set` | yes | ok |
| `--watch` | yes | ok |
| `--choose` | yes | ok |
| `--yes` | yes | ok |
| `--quiet` | yes | ok |
| `--fmt` | yes | ok |
| `--clear-cache` | yes | ok |
| `--ignore-version` | yes | ok |
| `--color` | yes | ok (`auto|always|never` documented) |

Subcommand flags (`serve --http/--addr/--token-file/--mcp/--ignore-version`,
`version --check/--json`, `analyze --json`, `lsp --log-file/--log-level`): all
have descriptions; wording consistent (lower-case, imperative).

## Findings

| # | Finding | Disposition |
|---|---------|-------------|
| A1 | Failure banner (`rune: …`) unstyled while lesser messages are styled — the reported bug | **verified fixed** — `rune:` prefix in Error role, tolerant resolution at banner time (cmd/rune/main.go, internal/cli/banner.go); C1 tests green |
| A2 | `--ignore-version` warning plain; cache warning styled (inconsistent Warning role) | **verified fixed** — internal/cli/version_gate.go uses the Warning role; C2 test green |
| A3 | `analyze` text diagnostics use a third format (no caret/color), diverging from run/fmt/LSP-parity rendering | **verified fixed** — renders via `diag.RenderAllCoded` (keeps the `error[RUNE2001]` token), styled summary; C4 tests green |
| A4 | Subcommand help uses Cobra default layout, unstyled — 008's FR-019 wording promised subcommand coverage | **verified fixed** — one grouped renderer for root + subcommands (cmd/rune/help.go `subHelp`); C5 tests green |
| A5 | Bare-`rune` overview: plain version header above styled task list; plain no-tasks fallback | **verified fixed** — header label Heading, docs URL Muted; C3 tests green |
| A6 | Watch/serve/fmt/clear-cache/confirm status lines plain while structurally identical run-status lines are styled | **verified fixed** — role table per C3 (Success labels, Muted chrome, Warning prompt); tests green |
| A7 | `--list` colored branch pads by rune count vs plain branch's byte padding — feared wide/CJK misalignment | **verified unreachable** — the lexer restricts task names to ASCII (grammar frozen by Constitution III), so the two width models always agree; parity pinned by `TestListTasksColoredAlignmentMatchesPlain`; the planned go-runewidth promotion was dropped as dead code |
| A8 | TUI picker duplicates palette literals (170/245/241) from `internal/style` — single-source promise (008 SC-007) violated | **verified fixed** — exported `style.Color*` constants consumed by internal/tui/styles.go, guarded by a palette-literal test |
| A9 | Missing `Example` blocks on subcommands | **corrected & fixed** — investigation showed serve/version/analyze/lsp already had Examples; only `completion` lacked one (its examples hid inside `Long`) — moved to a proper `Example` block |
| A10 | `docsURL` constant points at `github.com/glapsfun/rune` while the module path is `github.com/rune-task-runner/rune` — likely stale org name in the no-tasks fallback | **defer** — factual link fix, not output styling; needs owner confirmation of canonical URL |
| A11 | Interp-executor tasks don't echo command lines (shell executor echoes, styled Muted) — echo parity gap | **defer** — execution-semantics change, separate feature; 008 scoped it out deliberately |
| A12 | `--summary` / `version` stdout products unstyled | **no action by design** — script-consumed stdout (D10); recorded so SC-001's checklist reads 100% |
| A13 | `--color` was a root-local flag: it did not parse at all with subcommands (`rune --color=always analyze` → "unknown flag"), so analyze/help/serve color could not be forced or suppressed by flag | **found & fixed during implementation** — `--color` is now a persistent flag inherited by every subcommand (cmd/rune/root.go), satisfying FR-008/C8; not a new control, the same flag properly scoped |

## Verdict

`NO_COLOR` precedence and per-stream TTY detection verified correct
(cmd/rune/color.go). One genuine flag-parsing defect was found during
implementation (A13: `--color` unusable with subcommands) and fixed. The rest
of the user-facing "bug" is the styling coverage gap (A1–A6, A8), all fixed;
A7 proved unreachable and is pinned by test instead.
