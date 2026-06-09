# What is Rune?

> **New here? Start with this page**, then walk through [Getting started](getting-started.md).
> This page explains *what* Rune is, the idea behind it, and *when* to reach for it.

Rune is a **command runner**: one readable file at your repo root, the `Runefile`, that
captures every project command — `lint`, `test`, `build`, `deploy`, and the long tail of
scripts — so anyone (or anything) can run them by name.

```sh
rune test        # run the "test" task
rune --list      # show every documented task
```

## The main idea

A `Runefile` is **named blocks of commands**. You run a block by name; you declare that some
blocks depend on others. That's the whole model:

```rune
set default := "build"

# Build the project.
build:
    @echo "building…"

# Run the tests. `build` runs first, because `test` depends on it.
test: build
    @echo "testing…"
```

Run `rune` with no arguments and the default (`build`) runs. Run `rune test` and Rune runs
`build` first, then `test`. Run `rune --list` and you see both, with their descriptions.

Two things make this more than a pile of shell scripts:

1. **The same tasks serve humans *and* AI agents.** A person runs `rune test` in a terminal;
   an AI agent or IDE runs the *same* `test` task through the [Model Context Protocol](mcp.md).
   There's no second copy of "how to build this project" to drift out of date.
2. **Mistakes are caught before anything runs.** Reference a task that doesn't exist, a
   variable you never defined, or create a dependency cycle, and Rune refuses to run — pointing
   at the exact `file:line:column`.

## How it's different

| | Rune | A build system (`make`) |
|---|---|---|
| Runs a task when you ask | **Always** | Only if it thinks output is "out of date" |
| Skipping work | Opt-in per task via `[cache(...)]`, and logged | Implicit, timestamp-based |
| Cross-platform shell | **Built in** (pure-Go shell; same on Linux/macOS/Windows) | Depends on the system shell |
| AI agents run your tasks | First-class (MCP) | Not a concept |

If you've used [`just`](https://github.com/casey/just), Rune will feel familiar — with a
cross-platform shell, static validation, and the agent/MCP layer added.

## Use Rune when…

- You want one obvious place for "how do I build / test / deploy this?"
- You're replacing a `Makefile` full of `.PHONY` targets or a folder of ad-hoc scripts.
- You want the **same** commands to work for teammates, CI, and AI agents.
- You need tasks that behave identically on Linux, macOS, and Windows.

## Don't reach for Rune when…

- You need a true **build system** with file-dependency graphs and incremental rebuilds — Rune
  deliberately always runs a task when asked (caching is an explicit opt-in, not the default).
- Your logic needs loops, recursion, or general programming *in the config* — Rune's task
  language is intentionally small; real logic lives in task **bodies** (shell, Python, Node).

## Next steps

- **[Getting started](getting-started.md)** — install Rune and run your first task in minutes.
- **[Examples](examples/README.md)** — runnable starting points for real project shapes.
- **[Runefile language guide](runefile.md)** — the full syntax, by example.
