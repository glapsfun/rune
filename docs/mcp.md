# Using Rune with AI Agents (MCP)

> The agents capability guide. New here? Read [What is Rune?](overview.md) first, and see the
> runnable **[agent-driven example](examples/agent-driven/README.md)**.

Rune is AI-native: the same tasks you run from the CLI can be exposed to AI agents and IDEs
through the [Model Context Protocol (MCP)](https://modelcontextprotocol.io). An agent calls
your tasks as tools — it never needs to learn a bespoke interface.

## Starting the server

**Stdio (local agents / IDE integrations):**

```sh
rune mcp
```

Most local agent tooling launches the server this way and talks to it over stdin/stdout.

**Streamable HTTP (opt-in, for networked clients):**

```sh
rune serve --http --addr 127.0.0.1:8765 --token-file ./mcp-token.txt
```

- `--http` selects the Streamable HTTP transport (stdio is the default otherwise).
- `--addr` sets the bind address.
- `--token-file` supplies a bearer token required on every request.

A remote endpoint is **opt-in, intended to be localhost-bound, and token-gated** — it is
never enabled by default.

## How tasks become tools

- Every **non-private** task is exposed as an MCP tool. Add `[private]` to hide a task (it
  remains callable only as a dependency of another task).
- A task's **parameters** define the tool's input schema. Defaults and variadic
  parameters carry over.
- The task's **doc comment** (the comment directly above it) becomes the tool description.

```rune
# Build the project for a target.
build target="release":
    go build -tags {{target}} ./...

[private]
_internal-helper:
    @echo "not exposed to agents"
```

Here `build` is offered to the agent as a tool with a `target` argument; `_internal-helper`
is not.

## Security model (secure by default)

Exposing tasks to an agent grants execution capability, so Rune is conservative by default:

- **Read-only by default.** Agent access defaults to non-destructive tasks. Access to
  destructive tasks is an explicit, per-task opt-in.
- **Destructive tasks are gated.** A task marked `[confirm("…")]` is annotated with the
  MCP `DestructiveHint`, so clients can warn or require approval before invoking it.
- **Network tasks are flagged.** `[network]` sets the MCP `openWorldHint`.
- **Secrets come from the environment only.** API keys and secrets are read from the
  environment (or the agent CLI's own session) — **never** from the Runefile, and they are
  never included in any tool description, schema, or log.
- **Vendor-neutral.** The agent/LLM layer sits behind a provider interface; no single
  vendor is hard-coded.

```rune
# Safe: read-only, exposed to agents by default.
status:
    @git status --short

# Destructive: gated with confirm → DestructiveHint for agents.
[confirm("Delete all build output?")]
clean:
    rm -rf dist
```

## Agent task bodies

A task can itself be driven by an AI agent using the `agent` executor, which runs an
installed agent CLI (e.g. `claude`, `codex`, `copilot`) behind the vendor-neutral provider
interface:

```rune
# Summarize recent changes using the configured agent provider.
summarize (agent):
    Summarize the latest git changes in three bullet points.
```

Agent tasks default to **read-only** tool access; granting them destructive capability is
an explicit opt-in.

## See also

- [agent-driven example](examples/agent-driven/README.md) — a runnable agent task + gating
- [CLI reference](cli.md) — `rune mcp` / `rune serve` flags
- [Runefile language guide](runefile.md) — `[private]`, `[confirm]`, `[network]`, executors
- [Guides](how-to/README.md) — the rest of the capability deep dives
