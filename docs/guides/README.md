# Guides

> Task-oriented deep dives — one capability per page. Each explains the concept, shows the
> syntax, links a runnable example, and calls out the common pitfalls.

New to Rune? Start with **[What is Rune?](../overview.md)** and **[Getting started](../getting-started.md)**.
For the full syntax in one place, see the **[Runefile language guide](../runefile.md)**.

| Guide | Capability | Example |
|-------|-----------|---------|
| [Dependencies & post-hooks](dependencies-and-hooks.md) | Order work; run things before/after | [dependencies](../examples/dependencies/README.md) |
| [Parameters](parameters.md) | Defaulted, required, variadic inputs | [parameters](../examples/parameters/README.md) |
| [Caching](caching.md) | Opt-in content-hash skipping | [caching](../examples/caching/README.md) |
| [Parallelism](parallelism.md) | Run independent prerequisites concurrently | [parallel](../examples/parallel/README.md) |
| [Executors](executors.md) | Shell, Python, Node, agent bodies | [polyglot](../examples/polyglot/README.md) |
| [Settings & dotenv](settings-and-dotenv.md) | Project settings and `.env` | [settings-dotenv](../examples/settings-dotenv/README.md) |
| [Imports & modules](imports-and-modules.md) | Compose and namespace task files | [monorepo](../examples/monorepo/README.md) |
| [OS filtering](os-filtering.md) | Tasks scoped to an operating system | [os-filtering](../examples/os-filtering/README.md) |
| [AI agents (MCP)](../mcp.md) | Expose tasks to agents; security model | [agent-driven](../examples/agent-driven/README.md) |

Hit an error? See **[Troubleshooting](../troubleshooting.md)**.
