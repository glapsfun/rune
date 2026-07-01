# Use case: managing a Node / JavaScript project

> Capture the npm workflow for a Node project — `install`, `build`, `test` — and run a task
> body *in JavaScript* under the node executor. This walkthrough assumes you already have a
> Node project; adapt the echoed npm commands to your real ones.

**Backing example:** [`examples/node-project`](../examples/node-project/README.md) ·
**Features:** [dependencies](../how-to/dependencies-and-hooks.md), [executors (node)](../how-to/executors.md) ·
**Prerequisites:** `node`

## The Runefile

```rune
# Tasks for a Node/JavaScript project. The `version` task body runs under the
# node executor (real JavaScript); the rest echo the npm commands to adapt.

# Install dependencies.
install:
    @echo "npm install"

# Build the project (depends on install).
build: install
    @echo "npm run build"

# Run the test suite.
test: install
    @echo "npm test"

# Print the Node.js version — body runs under the node executor.
version (node):
    console.log("node " + process.version)
```

## Run it

```sh
rune version
```

```text
node v20.x.x
```

The exact version reflects your installed Node. `rune build` and `rune test` each run
`install` first, because both declare it as a dependency.

## Why it's written this way

- **`build: install`** and **`test: install`** share the `install` dependency, so either
  entry point sets up `node_modules` first. See [Dependencies & post-hooks](../how-to/dependencies-and-hooks.md).
- **`version (node)`** uses the **node executor**: the body is real JavaScript run through
  Node. Reach for it when a task is cleaner in JS than in shell. See [Executors](../how-to/executors.md).
- The `install`/`build`/`test` bodies **echo** their npm commands so the example runs
  anywhere; replace them with your real scripts.

> [!TIP]
> Prefer wiring your existing `package.json` scripts as the task bodies (`npm run build`)
> rather than duplicating logic — Rune becomes the single, discoverable entry point.

---

**Next:** [AI agents (MCP)](mcp-agents.md) · [Python project](python-project.md) · [Executors](../how-to/executors.md)
