# Data Model: Colorful CLI Output Everywhere

No persistent data. The feature's "model" is the mapping between output
surfaces, semantic theme roles, and the per-stream color decision. This file is
the authoritative inventory that SC-001's checklist verifies against.

## Entity: Semantic Theme (`internal/style.Theme`)

Existing roles (unchanged set), plus exported palette constants (D9).

| Role | Palette | Used for (after this feature) |
|---|---|---|
| `Error` | red (1), bold | diagnostic severity, caret, **failure banners (`rune:` prefix), cycle/watch errors, analyze summary error count** |
| `Warning` | yellow (3), bold | cache-write warning, **min-version override warning, `[confirm]` prompt text, analyze summary warning count** |
| `Success` | green (2) | `running:` label, **`formatted:` / `cleared:` labels** |
| `TaskName` | accent (170), bold | `--list` task names |
| `Heading` | accent (170), bold | `--list` group headers, help section titles (**all commands**), **overview `rune version:` label** |
| `Muted` | grey (245) | docs, echo, cache/dry-run notices, **watch chrome, MCP server banners, docs URL** |
| `Locator` | faint | `file:line:col` in diagnostics (**now also via analyze**) |
| `Caret` | red (1), bold | diagnostic caret span (**now also via analyze**) |

New exported constants (no new roles): `ColorError`, `ColorWarning`,
`ColorSuccess`, `ColorAccent`, `ColorMuted`, `ColorMutedDark` (241; consumed by
`internal/tui` for its Help style).

**Validation rules**: zero `Theme` renders all roles as identity (plain);
`New(false, w)` MUST return zero Theme; palette literals MUST NOT appear
outside `internal/style` (enforced by audit/grep test).

## Entity: Output Surface

Classification is exhaustive; every emitter is either **themed**, **plain by
design**, or **machine (never themed)**. State transition per surface:
`unstyled → themed` happens exactly once, gated by the stream's color decision.

### Themed (human-facing) — target state

| # | Surface | Site | Stream | Roles | Plain bytes |
|---|---------|------|--------|-------|-------------|
| S1 | Failure banner | cmd/rune/main.go | stderr | Error (prefix) | frozen |
| S2 | Cycle-error banner | internal/cli/run.go | stderr | Error (prefix) | frozen |
| S3 | Watch error banners | internal/cli/watch.go | stderr | Error (prefix) | frozen |
| S4 | Watch chrome (watching…/change detected…) | internal/cli/watch.go | stderr | Muted | frozen |
| S5 | Min-version warning | internal/cli/version_gate.go | stderr | Warning (label) | frozen |
| S6 | Analyze text diagnostics + summary | internal/cli/analyze.go | stdout | via diag renderer; Error/Warning counts | **reformatted once** (FR-006 exception) |
| S7 | Subcommand help (serve/version/analyze/lsp/completion/help) | cmd/rune/help.go | stdout | Heading | **reformatted once** (FR-006 exception) |
| S8 | Overview header + no-tasks fallback | internal/cli/run.go | stdout | Heading label / Muted URL | frozen |
| S9 | `[confirm]` prompt | internal/cli/run.go | stderr | Warning (text) | frozen |
| S10 | `formatted:` notice | internal/cli/fmt.go | stderr | Success (label) | frozen |
| S11 | `cleared:` notice | internal/cli/run.go | stderr | Success (label) | frozen |
| S12 | MCP server banners | internal/cli/serve.go | stderr | Muted | frozen |
| S13 | `--list` colored padding fix | internal/cli/run.go | stdout | (existing roles; width model fix) | frozen |
| — | Already themed (008): `--list`, run status, echo, dry-run, diagnostics, root help, picker | — | — | — | frozen |

### Plain by design (human-facing, script-consumed stdout products)

`--summary`, `rune version` / `version --check` text, `[confirm]`'s literal
`[y/N] ` suffix, all values (paths, addresses, version numbers) inside themed
lines.

### Machine (never themed, any color mode)

`--dump` (text + `--format json`), `analyze --json`, `version --json`,
completion scripts, LSP JSON-RPC stream, MCP tool-result buffers, agent
write-back capture.

## Entity: Color Decision (existing, unchanged)

`resolveColor(mode, isTTY)` per stream: `never` → off; `always` → on;
`NO_COLOR` non-empty → off; else TTY. Two resolution contexts:

| Context | Where | Streams |
|---|---|---|
| Normal run | `PersistentPreRunE` → `Options.ColorStdout/ColorStderr` | both |
| Cobra-early paths (help, failure banner) | tolerant flag re-read (help.go pattern, extended to banner per D1) | the one stream being written |

**Invariant**: no third resolution mechanism may be introduced; every new
surface consumes one of these two (FR-008).

## Relationships

```text
colorMode + TTY ──resolveColor──▶ ColorStdout/ColorStderr (bool, per stream)
                                          │
                     style.New(enabled, maskedWriter)
                                          │
                                    Theme (roles)
                                          │
        S1–S13 render via roles ──▶ mask.Writer (if secrets) ──▶ os.Stdout/err
                                          
style constants ──▶ internal/tui Styles (picker)   [D9: single palette source]
```

Masking sits *below* styling (theme renders onto the masked writer; `maskErr`
pre-masks banner messages), so FR-011 holds by construction and is regression-
tested, not re-implemented.
