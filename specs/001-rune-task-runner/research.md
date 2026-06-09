# Phase 0 Research: Rune Task Runner

**Date**: 2026-06-08 | **Feature**: 001-rune-task-runner

All Technical Context items are resolved — there are no remaining `NEEDS CLARIFICATION`
markers. Most foundational choices were pre-decided by the Build Brief and locked by the
project constitution; the open decisions (Brief §9) were settled in the spec's Clarifications
session. This document records each decision with rationale and the alternatives rejected,
so implementation does not re-litigate them.

## D1. Compiler front end: hand-written lexer + recursive-descent + Pratt

- **Decision**: Hand-write a Rob Pike state-function lexer, a recursive-descent parser for the
  declarative grammar, and a Pratt (top-down operator-precedence) sub-parser for the expression
  sublanguage. Tokens carry byte offset + line + column.
- **Rationale**: Constitution Principle IV mandates it. Hand-written recursive descent (as in the
  Go compiler and `just`) gives the error-message quality, recovery, and source spans that
  Principle II ("errors are a feature") requires. Pratt parsing cleanly handles concatenation
  `+`, path-join `/`, comparisons, and `if/else` precedence.
- **Alternatives rejected**: `goyacc`/ANTLR (poor error UX, heavyweight, constrains diagnostics);
  `alecthomas/participle` (acceptable for a throwaway feel-test only — no left recursion, less
  control; MUST NOT ship per Principle IV).

## D2. CLI framework: cobra + custom dynamic dispatcher

- **Decision**: Use `spf13/cobra` to parse **global flags** (`--file`, `--list`, `--dry-run`,
  `--set`, `--watch`, `mcp`/`serve` subcommands, `--version`, `--help`) and to generate shell
  completions. Route the remaining positional args through Rune's own dispatcher, since task
  names are dynamic (defined by the Runefile, not registered cobra subcommands).
- **Rationale**: Cobra is the de-facto standard, gives first-class completion generation
  (Phase 8), and its `DisableFlagParsing` / `Args` hooks let a thin custom layer own task
  dispatch. `rune name=value task --task-flag` style trailing args pass through cleanly.
- **Alternatives rejected**: `alecthomas/kong` (lighter but weaker completion story); fully
  hand-rolled flag parsing (reinvents help/completions for no benefit).

## D3. Default shell executor: `mvdan.cc/sh/v3` (never system sh)

- **Decision**: Execute `(sh)` bodies through the pure-Go `mvdan/sh` interpreter (`syntax` to
  parse, `interp` to run). Optionally wire `moreinterp/coreutils` so `cat`/`cp`/`ls` exist on
  Windows. Honor `set shell := [...]` by switching that task to temp-file + exec of the named
  shell (the override path).
- **Rationale**: Principle V — identical cross-platform behavior with no WSL/Git-Bash and no CGO.
  This is the same library `task` and Charmbracelet *Crush* use.
- **Alternatives rejected**: shelling out to system `/bin/sh` (non-portable, the very trap
  Principle V forbids).

## D4. Alternate-language bodies: temp-file + exec the real interpreter

- **Decision**: For `(python)`/`(node)`/custom executors and `[script("cmd")]`, write the
  interpolated body to a temp file (executable bit on Unix) and exec the configured interpreter
  with that path. Interpreter command lists come from `set python`/`set node`/`[script]`.
- **Rationale**: Mirrors `just`'s shebang/`[script]` model; zero embedded runtimes; Open
  Decision §9.5 resolved to "shell out." Missing interpreter → actionable error (FR-017).
- **Alternatives rejected**: embedding Python/Node runtimes (enormous binary, contradicts
  Principle V single-binary simplicity).

## D5. MCP integration: official `modelcontextprotocol/go-sdk`; stdio always, HTTP opt-in

- **Decision**: Use the official Go MCP SDK for both server and client. The **server** exposes
  non-private tasks as tools (`mcp.AddTool` with `Name`, `Description`, `Annotations`, derived
  input schema). **stdio** transport is always available; the **Streamable HTTP/SSE** transport
  is opt-in, binds `127.0.0.1` by default, and requires a bearer/token check before any
  list/call (FR-031, Clarification Q1).
- **Rationale**: Native, schema-accurate generalization of `just-mcp` (Principle VII). Token +
  localhost default = "secure by default."
- **Alternatives rejected**: hand-rolled JSON-RPC (re-implements a moving spec); always-on remote
  with no auth (rejected in Clarification Q1, option C).
- **Verify at implementation**: exact SDK surface (`mcp.AddTool` generic signature, transport
  constructors) against the installed `v1.x`; pin the version in `go.mod`.

## D6. Agent task type: drive an installed agent CLI behind a `Provider` interface

- **Decision**: The `(agent)` executor resolves `{{…}}` in the prompt, starts the MCP server
  in-process to offer the project's allowed tasks as tools, then invokes a **configured agent
  CLI** (`claude`, `codex`, `copilot`, …) via a configurable command, handing it the prompt and
  the MCP endpoint. The CLI owns its own authentication. This sits behind a vendor-neutral
  `Provider` interface so other backends (direct hosted APIs) can be added later.
- **Rationale**: Clarification Q2 (option A). Reuses already-authenticated local agent tools, no
  API keys in core, vendor-neutral (Principle VII / FR-030). Missing/unauthenticated CLI →
  actionable error (FR-027).
- **Alternatives rejected**: hard-coding a hosted-API SDK in core (vendor lock-in, key
  management); supporting both CLI + API in v1 (more surface than needed — deferred).

## D7. Agent authorization model: author-declared destructiveness + read-only default

- **Decision**: Private tasks are never exposed. Other non-private tasks are agent-callable.
  A task is gated (requires explicit opt-in/approval) **iff** its author marks it
  (`[confirm]`/destructive attribute) → mapped to MCP `DestructiveHint`. No content heuristics.
  Operators MAY layer a stricter allow-list. Agent tasks default to calling only non-gated tasks.
- **Rationale**: Clarification Q3 (option A). Predictable, author-controlled, matches the
  `[confirm]`→`DestructiveHint` mapping (FR-028).
- **Alternatives rejected**: heuristic destructiveness detection (fragile, false sense of safety);
  full default-deny allow-list as the only mode (more friction than the brief's "expose
  non-private tasks").

## D8. Caching: SHA-256 fingerprint, project-local `.rune/cache/`

- **Decision**: For `[cache(inputs=…, outputs=…)]` tasks, compute a SHA-256 fingerprint over
  (sorted hashed input files + task body + resolved variables + executor identity). Skip iff the
  fingerprint matches the stored one **and** all declared outputs exist; otherwise run and update.
  Store fingerprints as JSON under `.rune/cache/` at the Runefile root (advise `.gitignore`).
- **Rationale**: Open Decision §9.7. SHA-256 is in the stdlib (zero dependency, no CGO,
  deterministic) and fast enough for v1. A visible "cached"/"running" log line satisfies
  Principle I (no silent skipping).
- **Alternatives rejected**: BLAKE3 (faster, but adds a dependency — deferred as an optimization
  once profiling shows hashing is a bottleneck); XDG cache dir (project-local is simpler to
  reason about and clean; revisit if cross-project sharing is wanted).

## D9. Parallel execution: `errgroup` bounded by CPU count

- **Decision**: `[parallel]` runs independent prerequisites concurrently using
  `errgroup.Group` with `SetLimit(num_cpus())`. The scheduler memoizes `(task, resolved-args)`
  so a node runs at most once even under concurrency.
- **Rationale**: FR-019/FR-005/SC-009. `errgroup` gives first-error cancellation matching
  fail-fast semantics.
- **Alternatives rejected**: unbounded goroutines (resource thrash); a custom worker pool
  (errgroup already fits).

## D10. Supporting choices

- **`.env` loading**: parse via `mvdan/sh`'s `shell` package (consistent expansion semantics
  with the executor) rather than a separate dotenv lib. *Alternative*: `joho/godotenv` —
  rejected to keep one expansion engine.
- **Terminal output**: `fatih/color` for diagnostics/listings (simple, NO_COLOR-aware).
  *Alternative*: `lipgloss` — richer but heavier; revisit for `--choose` UI.
- **`--watch`**: `fsnotify/fsnotify`. **`--choose`**: shell out to `fzf` if present, else a
  minimal built-in picker (Phase 8).
- **Body syntax**: significant indentation, consistent within a task; tab/space mixing within
  one body is a located error (Clarification Q4 / FR-002).
- **Backward compatibility**: breaking changes are opt-in **per file** via a version pragma
  (e.g. a `set rune_version := "x"` style marker); default interpretation never changes under a
  user (Principle Governance / FR-033). Exact pragma syntax finalized during Phase 6.
- **Release**: `goreleaser`, `CGO_ENABLED=0`, build matrix Linux/macOS/Windows × amd64/arm64.

## Open items deferred to later phases (not blocking)

- Exact built-in function signatures/edge semantics (`replace_regex`, `datetime` format) —
  finalized in Phase 3 with golden tests.
- Final product name / file name / extension (Brief §9.1) — "Rune"/"Runefile"/".rune" provisional;
  swap is a mechanical rename, not an architectural change.
- Version pragma concrete syntax — Phase 6.
- BLAKE3 adoption — only if SHA-256 hashing is profiled as a bottleneck.
