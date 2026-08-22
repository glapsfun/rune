# Research: Agent Context Prep (`[context]`)

**Feature**: 021-agent-context-prep | **Date**: 2026-08-21

All unknowns were resolved interactively during brainstorming (2026-08-20)
and by direct codebase/SDK verification. No NEEDS CLARIFICATION markers
remain in the Technical Context.

## R1. Which agent surfaces receive the context

- **Decision**: Both — the MCP server (`rune mcp` / `rune serve`) and the
  `agent` executor, from one hook definition.
- **Rationale**: "Before it even chooses a task" applies to both places an
  agent meets a Runefile; one concept injected consistently everywhere beats
  two divergent mechanisms.
- **Alternatives considered**: MCP-only (leaves `(agent)` tasks blind);
  agent-executor-only (leaves the primary external-agent surface blind).

## R2. Declaration syntax

- **Decision**: A new bare task attribute `[context]` (`ast.AttrContext`),
  at most one per composed Runefile.
- **Rationale**: Attributes are Rune's established metadata surface
  (`[private]`, `[network]`, `[no-exit-message]` precedents); the task stays
  a real, CLI-runnable, doc-commented task; parser change is one case label;
  LSP/formatter support falls out of existing machinery.
- **Alternatives considered**: reserved task name (magic naming, collision
  risk, implicit); new top-level block (new grammar production, parser,
  formatter, and LSP work disproportionate to the feature — and Principle
  III treats grammar growth as constitutional).

## R3. MCP delivery channel

- **Decision**: The `instructions` field of the initialize result, via
  `mcp.ServerOptions.Instructions`.
- **Rationale**: It is the protocol's designated slot for model-facing
  context; mainstream clients (Claude Code, IDE integrations) append it to
  the system context; zero extra round-trips; verified present in the vendored
  SDK (`go-sdk v1.6.1`, `ServerOptions.Instructions`, client-readable via
  `ClientSession.InitializeResult().Instructions`).
- **Alternatives considered**: MCP resource (many clients never auto-fetch —
  agents would not see it); per-tool description preamble (duplicates the
  text N times, pollutes schemas).

## R4. Freshness model

- **Decision**: Gather once per MCP server start; gather per invocation for
  `(agent)` tasks.
- **Rationale**: The SDK fixes `Instructions` at server construction (no
  per-session hook exists in `ServerOptions`), which equals per-session
  freshness on stdio (one session per process — the dominant deployment) and
  shared instructions across HTTP sessions of one process. `(agent)` tasks
  construct a fresh provider run each invocation, so per-invocation gathering
  is natural there. Spec FR-002 was amended to "once per server start" to
  match (2026-08-21).
- **Alternatives considered**: refresh tool (more agent-callable surface —
  deferred until demanded); TTL cache (cache machinery + config knob for
  marginal benefit; violates YAGNI and NFR-002's no-config stance).

## R5. Failure and timeout policy

- **Decision**: Best-effort with a hard 10-second timeout; on non-zero exit,
  timeout, or error: one-line notice in place of the context —
  `(context hook %q failed; proceeding without project context)` — plus a
  warning on Rune's stderr. Never blocks a session or task.
- **Rationale**: An informational feature must not become an availability
  risk (a broken linter must not brick agent access). A hook that "fails
  because the health check found problems" follows normal task semantics —
  authors append `|| true` to deliver findings as output.
- **Alternatives considered**: fail-closed (bricks all agent access on a
  broken health command); injecting failure stdout/stderr as context (noisy
  tracebacks in every prompt, murkier masking story).

## R6. Output processing order and limits

- **Decision**: Mask first (inherent — gathering reuses `mcpAdapter.Call`,
  whose buffers only ever hold masked text), then truncate to 8 KiB with a
  trailing `[truncated]` marker. Constants, not configuration.
- **Rationale**: Reusing the existing masking choke point (Principle VII)
  means no new leak path and no duplicated policy; mask-before-truncate
  guarantees a secret can never straddle the cap into visibility. 8 KiB is
  ample for branch + status + lint summary while bounding prompt cost.
- **Alternatives considered**: configurable limits (NFR-002 forbids; no
  demonstrated need); no cap (a chatty linter could flood every prompt).

## R7. Tool-surface exclusion

- **Decision**: The `[context]` task is never exposed as an MCP tool
  (filtered in `mcpAdapter.Tasks()`, independent of `[private]`), but stays
  CLI-runnable by name; `[private]` composes as usual for listings.
- **Rationale**: Agents already receive the hook's output; a callable
  duplicate is redundant surface. The filter must live in the MCP adapter,
  not the shared `visibleOn` predicate, or `--list` would wrongly hide a
  non-private hook.
- **Alternatives considered**: implying `[private]` (would hide it from
  `--list`, hurting debuggability); exposing it as a tool (redundant, and a
  destructive-adjacent hook would widen the callable surface for no value).

## R8. Static validation set

- **Decision**: Analyzer errors (code RUNE2007 `CodeInvalidAttribute`) for:
  duplicate `[context]` (post-import, both offenders named via a related
  location), `[context]`+`[confirm]`, any parameter lacking a default, and
  the `agent` executor on the hook itself. OS attributes compose; an
  OS-mismatched hook is silently absent at runtime (existing `AvailableOn`
  semantics, FR-008).
- **Rationale**: The hook must run unattended and non-recursively; every
  misuse is statically detectable, and Principle II demands analyzer-time
  errors with spans. Reusing RUNE2007 avoids minting a code for what is an
  attribute-usage error.
- **Alternatives considered**: runtime rejection (violates Principle II);
  new diagnostic code (no added value; RUNE2007 already covers attribute
  misuse).

## R9. Prompt assembly format (agent executor)

- **Decision**: Fixed delimiters, verbatim:
  `## Project context (generated by Runefile [context] task)` then the text,
  a blank line, `## Task`, then the task body. Pass-through (byte-identical
  prompt) when no hook exists.
- **Rationale**: Models need an unambiguous boundary between gathered state
  and the author's instruction; naming the source ("generated by Runefile
  [context] task") tells the model the provenance. Pass-through preserves
  FR-009/NFR-001 for existing Runefiles.
- **Alternatives considered**: XML-style tags (foreign to the Markdown-first
  prompt conventions of current agent CLIs); no delimiter (model cannot
  distinguish instruction from state).
