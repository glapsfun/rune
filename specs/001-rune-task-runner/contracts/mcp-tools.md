# Contract: MCP Tool Exposure & Agent Tasks

**Feature**: 001-rune-task-runner | **Date**: 2026-06-08

Covers FR-025…FR-031 and Clarifications Q1–Q3. Two surfaces: (A) Rune **as an MCP server**
(exposing tasks to external agents/IDEs), and (B) the **`(agent)` executor** (a task whose body
is a prompt, run by driving an installed agent CLI).

## A. MCP server — task → tool mapping (FR-025, FR-026)

Each **non-private** task becomes one MCP tool:

| Tool field | Source |
|------------|--------|
| `name` | task name; submodule tasks namespaced as `mod__task` (e.g. `docker__push`) |
| `description` | the task's doc comment (or `[doc("…")]`) |
| `inputSchema` | derived from params: each param → property; `defaulted`→optional, `required`→required, `variadic`→array |
| `annotations.destructiveHint` | `true` iff the task has `[confirm]`/destructive attribute (Q3) |
| `annotations.openWorldHint` | `true` if the task is known to touch the network |

Tool **call** handler: validate args against the schema → run the task through the **same
scheduler** the CLI uses → return `{ stdout, stderr, exitCode }` as the tool result (FR-026).
Private tasks are never listed or callable (FR-028). Secret values never appear in any name,
description, schema, or result surfaced to the agent (FR-029).

### Authorization (Q3 / FR-028)

- Private → not exposed at all.
- Non-private, **not** marked destructive → callable.
- Marked destructive (`[confirm]`) → exposed with `destructiveHint`, but a **call** requires
  explicit approval per policy (`--yes`, an allow-list, or interactive confirmation); refused
  otherwise.
- Operators MAY supply an explicit allow-list that further narrows callable tasks.

## B. Transports & security (FR-031 / Q1)

| Transport | Default | Auth |
|-----------|---------|------|
| stdio | **always available** (`rune mcp`) | inherited from the local process/session |
| Streamable HTTP/SSE | **opt-in** (`rune serve --mcp --http`) | binds `127.0.0.1` by default; **requires** a bearer token (`--token-file`/env) before any `list`/`call` |

Rules:
- The remote endpoint is OFF unless explicitly requested.
- Binding to a non-localhost address requires an explicit `--addr` and still enforces the token.
- Every list/call over HTTP without a valid token → rejected (SC-010).

## C. `(agent)` executor (FR-027 / Q2)

Execution steps for a task declared `(agent)`:

1. Resolve `{{…}}` interpolations in the prompt body.
2. Start the MCP server **in-process**, exposing only the tasks this agent is allowed to call
   (default: non-destructive, non-private; narrowed by `[agent(allow=[…])]`).
3. Invoke the configured **agent CLI** via a configurable command template, e.g.
   `set agent_cmd := ["claude", "-p"]` (or `codex`, `copilot`), handing it the prompt and the
   MCP endpoint so it can call back into project tasks.
4. The CLI manages its **own** authentication (no API keys in the Runefile, FR-029).
5. Capture final output → becomes the task's output. Success criterion (default): the CLI exits 0.

`[agent(...)]` knobs (incremental): `model`, `max_tool_calls`, `allow=[task,…]`, `readonly`
(deny destructive tools — the default).

### Failure modes (FR-027 edge case)

- Configured agent CLI not on PATH → actionable error naming the missing tool; exit 1.
- CLI present but unauthenticated → actionable error pointing to the CLI's own login; exit 1.
- No `agent_cmd`/provider configured → configuration error; never invent credentials.

## Provider interface (FR-030)

```
Provider.Run(ctx, prompt, toolSession, opts) -> (finalText, toolTrace, error)
```

v1 concrete provider = **agent-CLI provider** (drives `claude`/`codex`/`copilot`/compatible).
The interface stays open so a direct hosted-API provider can be added later without changing
core. No single vendor is hard-coded.

## Acceptance mapping

US4 scenarios 1–5 → list/describe (A), call via shared engine (A), destructive gating (B
authz + B transport), agent task drives CLI (C), zero secret leakage (FR-029 + SC-007).
