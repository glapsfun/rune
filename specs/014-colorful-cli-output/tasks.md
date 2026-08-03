# Tasks: Colorful CLI Output Everywhere

**Input**: Design documents from `/specs/014-colorful-cli-output/`

**Prerequisites**: plan.md, spec.md, research.md (D1–D10), data-model.md (S1–S13), contracts/styled-output.md (C1–C9), contracts/cli-audit.md (A1–A12)

**Tests**: Included — the constitution (Principle VI) mandates test-first. Every test task is written first and MUST fail (or, for invariance tests, pass against pre-014 bytes) before its implementation task lands.

**Organization**: Grouped by user story; each story is independently implementable and testable. All tests run inside Docker: `docker-compose run --rm test go test ./...`.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: US1 (errors/warnings), US2 (status messages + analyze), US3 (help + audit)

## Phase 1: Setup

**Purpose**: Dependency housekeeping — no behavior change.

- [x] T001 Promote `mattn/go-runewidth` from indirect to direct dependency in go.mod (add explicit require + `docker-compose run --rm test go mod tidy`); no version change (v0.0.19 already in module graph)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared palette surface both US2 (list padding context) and Polish (TUI unification) build on; zero behavior change, so it cannot break any story.

- [x] T002 Export palette constants `ColorError`, `ColorWarning`, `ColorSuccess`, `ColorAccent`, `ColorMuted` and add `ColorMutedDark` (= "241", the picker's help grey) in internal/style/style.go, replacing the unexported `color*` consts; add unit test in internal/style/style_test.go asserting the constants' values and that `New(false, …)` still returns the zero Theme (D9)

**Checkpoint**: `go test ./internal/style` green; no output change anywhere.

---

## Phase 3: User Story 1 - Failures and warnings are visually consistent (Priority: P1) 🎯 MVP

**Goal**: Every error banner (`rune: …`) and every warning uses the shared Error/Warning roles; plain output byte-frozen; masking intact. Contracts C1, C2.

**Independent Test**: quickstart.md Scenarios 1, 2, 7 — styled banner under `--color=always`, byte-identical plain banner under `--color=never`/`NO_COLOR`, exit codes unchanged, `***` masking preserved in styled banners.

### Tests for User Story 1 (write first, watch them fail)

- [x] T003 [P] [US1] Integration test in test/integration/error_banner_test.go: failure banner `rune:` prefix carries ANSI under `--color=always` (piped), message text escape-free; byte-identical `rune: <msg>` under `--color=never` and `NO_COLOR=1`; exit code preserved; `[no-exit-message]` still prints nothing; cycle error (`rune: cycle …`) gets the same treatment; invalid `--color` value on the banner path falls back to auto without a second error (C1, C8)
- [x] T004 [P] [US1] Integration test in test/integration/watch_styling_test.go: watch per-iteration failure banner and watch-error banner styled under `--color=always`, plain bytes frozen otherwise; watch keeps running after a styled failure (C1)
- [x] T005 [P] [US1] Extend test/integration/minimum_version_test.go: `--ignore-version` warning label `warning` styled identically to the cache-write warning under `--color=always`; plain line byte-identical to pre-014 (C2)
- [x] T006 [P] [US1] Extend test/integration/secret_masking_cli_test.go: failing task whose error embeds a secret, run with `--color=always` — banner shows ANSI + `***`, never the secret value, in both colored and plain modes (C1, FR-011)

### Implementation for User Story 1

- [x] T007 [US1] Add `printErrorBanner(opts Options, msg string)` in new file internal/cli/banner.go rendering `rune:` via `opts.themeStderr().Error` + one space + plain message; switch the cycle-error site in internal/cli/run.go (`fmt.Fprintln(e.opts.Stderr, "rune: "+cyc.Error())`) and both watch error sites in internal/cli/watch.go (`runOnce` failure, watcher error channel) to it (D2)
- [x] T008 [US1] Style the top-level banner in cmd/rune/main.go: tolerant resolution at banner time — read `--color` off the root command falling back to auto (help.go pattern), `resolveColor(mode, streamIsTTY(os.Stderr))`, build `style.New` on os.Stderr, render `rune:` prefix with Error role (D1); keep ValidationError/Silent suppression and `cli.CodeFor` exactly as-is
- [x] T009 [P] [US1] Render the min-version override warning label via `opts.themeStderr().Warning.Render("warning")` in internal/cli/version_gate.go, mirroring the cache-write warning's shape (D3)
- [x] T010 [US1] Run the full golden + integration suite in Docker and confirm zero golden diffs (US1 surfaces are all byte-frozen); fix any invariance break before proceeding

**Checkpoint**: US1 fully functional — the reported bug is fixed; MVP shippable.

---

## Phase 4: User Story 2 - Every remaining status message joins the theme (Priority: P2)

**Goal**: Overview header, confirm prompt, `formatted:`/`cleared:` labels, watch/serve chrome styled; `analyze` text output reaches run-path diagnostic parity (the one authorized plain-layout change besides help); `--list` colored padding uses display width; machine surfaces proven ANSI-free. Contracts C3, C4, C6, C7.

**Independent Test**: quickstart.md Scenarios 3, 4, 6, 8.

### Tests for User Story 2 (write first)

- [x] T011 [P] [US2] Integration test in test/integration/analyze_styling_test.go: `rune analyze` text output on a broken Runefile matches the run-path diagnostic rendering (severity word, faint locator, caret span) modulo stream; summary severity words styled only when count > 0 under `--color=always`; `analyze --json` byte-identical to pre-014 under all color modes; exit codes 0/3/2 unchanged (C4)
- [x] T012 [P] [US2] Integration test in test/integration/status_styling_test.go covering the C3 table: `formatted:`/`cleared:` labels Success, watch chrome and MCP server banners Muted, confirm prompt text Warning with ` [y/N] ` plain, bare-`rune` overview header Heading + docs URL Muted — each asserted colored under `--color=always` and byte-frozen plain under `--color=never`/`NO_COLOR` (C3, S8–S12)
- [x] T013 [P] [US2] Integration test in test/integration/machine_plain_test.go: zero `\x1b` bytes under `--color=always` for `--dump`, `--dump --format json`, `analyze --json`, `version --json`, `completion bash|zsh|fish|powershell` (C7)
- [x] T014 [P] [US2] Extend internal/cli/mcp_test.go: MCP tool-result content contains no ANSI even when the engine Options are constructed alongside forced-color CLI settings (adapter keeps color booleans false) (C7)
- [x] T015 [P] [US2] Unit test in internal/cli/run_test.go: `listTasks` colored branch aligns the `#` doc column for a corpus including a double-width (CJK) task name — visible column identical after ANSI strip to the plain branch's column for ASCII names; plain branch bytes unchanged (C6, D8)

### Implementation for User Story 2

- [x] T016 [US2] Rework `printAnalyzeText` in internal/cli/analyze.go: render diagnostics via `diag.RenderAll(diags, src, themeStdout())` with a source provider over the analyzed files; keep the summary line wording, styling severity words via Error/Warning roles when counts > 0; leave `printAnalyzeJSON` untouched; regenerate analyze text goldens deliberately in the same change with a commit message flagging the authorized reformat (D4, C4)
- [x] T017 [US2] In internal/cli/run.go: (a) overview — `rune version:` label Heading, no-tasks docs URL Muted (D6); (b) confirm prompt text Warning with ` [y/N] ` plain (D7); (c) `cleared:` label Success (D7); (d) colored `--list` branch pads via `runewidth.StringWidth` for both the max-width scan and per-row pad, plain branch untouched (D8)
- [x] T018 [P] [US2] Style `formatted:` label with Success role via `opts.themeStderr()` in internal/cli/fmt.go (D7)
- [x] T019 [P] [US2] Style both `rune MCP server on …` banners Muted via `opts.themeStderr()` in internal/cli/serve.go (D7)
- [x] T020 [P] [US2] Style `watching … (Ctrl-C to stop)` and `change detected, re-running…` Muted in internal/cli/watch.go (D7; error banners already themed by T007)

**Checkpoint**: US1 + US2 independently green; only analyze goldens changed, deliberately.

---

## Phase 5: User Story 3 - Help is uniform across every command (Priority: P3)

**Goal**: One grouped, theme-aware help renderer for root and all subcommands; command metadata completed; audit dispositions confirmed. Contracts C5; audit A4, A9; FR-012/SC-006.

**Independent Test**: quickstart.md Scenario 5 — every command's `--help` shows the grouped layout, zero ANSI piped, styled headings on a TTY.

### Tests for User Story 3 (write first)

- [x] T021 [P] [US3] Extend test/integration/cli_help_test.go: for root and each subcommand (serve, version, analyze, lsp, completion, help) assert grouped sections (description, Usage:, Examples: where defined, Flags:), zero ANSI when piped, Heading-role ANSI on section titles under `--color=always`, exit 0; root help plain bytes unchanged from the 008 baseline (C5)

### Implementation for User Story 3

- [x] T022 [US3] Generalize `rootHelp` in cmd/rune/help.go into a grouped renderer driven by cobra metadata (Long/Short, `UseLine()`, `Example`, `Flags().FlagUsages()`; root additionally keeps its Tasks and Commands sections and worked examples); remove the `cmd.HasParent()` early-out so `SetHelpFunc` covers subcommands; keep the tolerant `--color` resolution (D5)
- [x] T023 [P] [US3] Add missing `Example` blocks to `newServeCmd` in cmd/rune/serve.go, `newVersionCmd` in cmd/rune/version.go, `newCompletionCmd` in cmd/rune/completion.go, and `newLSPCmd` in cmd/rune/lsp.go, wording-consistent with analyze's existing example (audit A9)
- [x] T024 [US3] Regenerate subcommand help goldens as the new frozen baseline (authorized reformat #2); verify root help plain bytes are byte-identical to pre-014
- [x] T025 [US3] Re-verify every row of specs/014-colorful-cli-output/contracts/cli-audit.md against the implemented tree; flip A1–A9 dispositions to "verified fixed", keep A10/A11 deferred with rationale, and confirm A12's by-design exclusions still hold (SC-006)

**Checkpoint**: All three stories independently functional.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T026 [P] Replace the palette literals in internal/tui/styles.go ("170", "245", "241") with `style.ColorAccent`, `style.ColorMuted`, `style.ColorMutedDark`; add a guard test in internal/style/style_test.go that greps the repo's non-test Go sources for `lipgloss.Color("` outside internal/style and fails on any hit (FR-010, D9, SC-007 of 008)
- [x] T027 [P] Update docs/cli.md's color/output section: styled failure banners and warnings, analyze diagnostic parity, uniform subcommand help, unchanged `--color`/`NO_COLOR` controls and machine-format guarantees (docs are a tested fixture — keep examples runnable)
- [x] T028 Execute all 8 quickstart.md scenarios manually and tick the sign-off checklist in specs/014-colorful-cli-output/quickstart.md
- [x] T029 Final gates in Docker: `docker-compose run --rm test go test ./...`, `docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...`, `golangci-lint run` zero issues, `gofumpt`/`goimports` clean; review the full golden diff and confirm it is limited to analyze text + subcommand help (SC-002)

---

## Implementation Deviations (recorded 2026-07-23)

Facts discovered during implementation that changed four tasks; all are
reflected in contracts/cli-audit.md (A7, A9, A13) and quickstart.md:

- **T001 reverted**: the lexer restricts task names to ASCII (grammar frozen by
  Constitution III), so the wide-rune misalignment is unreachable —
  `go-runewidth` stayed indirect; go.mod/go.sum are untouched.
- **T015/T017(d) adjusted**: instead of a display-width rewrite, the colored
  and plain `--list` branches' parity is pinned by
  `TestListTasksColoredAlignmentMatchesPlain`; no production code change.
- **T016 grew two items**: (1) `internal/diag` gained `RenderCoded`/
  `RenderAllCoded` so analyze keeps its `error[RUNE2001]` token while reusing
  the run-path renderer; (2) `--color` was found to be root-local — unusable
  with subcommands (audit A13) — and became a persistent flag in
  cmd/rune/root.go, which C8's "governs every surface" requires.
- **T023 narrowed**: serve/version/analyze/lsp already had `Example` blocks
  (audit A9 was wrong); only `completion` was fixed (examples moved out of
  `Long` into `Example`).
- **T024 was a no-op**: no golden files cover analyze text or help — the
  reformats are held by substring/strip-equality integration tests, so the
  final diff has **zero** testdata changes (stronger than SC-002 required).

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: none — start immediately
- **Foundational (Phase 2)**: independent of Phase 1; BLOCKS T026 only (constants); stories US1–US3 don't strictly need it but it lands first because it's zero-risk
- **US1 (Phase 3)**: after Phase 2 — no dependency on other stories
- **US2 (Phase 4)**: after Phase 2; T020 touches internal/cli/watch.go after T007 (same file) — run US2 after US1, or coordinate the file
- **US3 (Phase 5)**: after Phase 2 — independent of US1/US2 (different files: cmd/rune/help.go + command metadata)
- **Polish (Phase 6)**: T026 needs T002; T027–T029 need all stories done

### Story Dependency Notes

- US1 is the MVP and the bug fix — ship-ready alone.
- US2's only cross-story file overlap is watch.go (T007 in US1 vs T020); everything else is disjoint.
- US3 is fully disjoint from US1/US2 (help layer only) and can run in parallel with either.

### Within Each Story

Tests (T003–T006, T011–T015, T021) MUST be written and failing/baselined before their implementation tasks; golden regenerations (T016, T024) are deliberate, reviewed steps, never a side effect.

## Parallel Example: User Story 1

```bash
# All US1 test tasks in parallel (different files):
T003 test/integration/error_banner_test.go
T004 test/integration/watch_styling_test.go
T005 test/integration/minimum_version_test.go
T006 test/integration/secret_masking_cli_test.go
# Then implementation: T007 + T009 in parallel; T008 after T007 (shared banner format); T010 last.
```

## Implementation Strategy

**MVP first**: Phase 1 → 2 → 3 (US1), validate quickstart Scenarios 1/2/7, ship — the user-reported bug is fixed at this point. Then US2 (Scenarios 3/4/6/8), then US3 (Scenario 5), then Polish. Each story leaves the corpus green and plain output frozen, so any checkpoint is releasable.
