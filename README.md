# Rune

**A shared task runner for humans and AI agents.**

Rune runs your project's commands — `lint`, `test`, `build`, `deploy` — from one readable
file, the `Runefile`. Humans run tasks from the CLI; AI agents and IDEs run the *same*
tasks through the Model Context Protocol (MCP). It ships as a single static binary with no
runtime dependencies.

```rune
set default := "build"

# Build the project.
build target="release": fetch
    go build -tags {{target}} ./...

# Fetch dependencies.
fetch:
    @go mod download

# Run tests.
test: build
    go test ./...
```

```sh
rune              # runs the default task (build)
rune test         # runs `build` (a dependency), then `test`
rune --list       # shows documented tasks
```

## Why Rune?

If you've used `make` or `just`, Rune will feel familiar — with a few deliberate choices:

- **Command runner, not a build system.** Tasks always run when asked. No `.PHONY` traps,
  no timestamp-based "up-to-date" skipping. Content-hash caching exists, but only as an
  explicit per-task `[cache(...)]` opt-in.
- **Errors are a feature.** Undefined variables, unknown dependencies, dependency cycles,
  and arity mismatches are caught *before anything runs*, with precise `file:line:col` +
  caret-underlined diagnostics.
- **Boringly portable.** One static binary. The default shell executor is a pure-Go,
  cross-platform shell (`mvdan.cc/sh`) — your `(sh)` tasks behave the same on Linux, macOS,
  and Windows, with no WSL or Git-Bash required.
- **AI-native and secure by default.** Tasks are first-class MCP tools. Agent access is
  read-only by default, destructive tasks are gated, and secrets come from the environment
  only — never the Runefile.
- **Multi-language bodies.** Write task bodies in shell, Python, Node, or drive an AI agent.

## Install

```sh
go install github.com/rune-task-runner/rune/cmd/rune@latest
```

Or grab a prebuilt binary from [Releases](https://github.com/rune-task-runner/rune/releases),
or run it in a container:

```sh
docker run --rm -v "$PWD":/work ghcr.io/rune-task-runner/rune --list
```

Full instructions: **[Installation guide](docs/installation.md)**.

## Documentation

| Guide | What's inside |
|-------|---------------|
| [What is Rune?](docs/overview.md) | The idea, the mental model, and when to use it |
| [Getting started](docs/getting-started.md) | Zero → your first task in minutes |
| [Examples](docs/examples/README.md) | Runnable starting points for real project shapes |
| [Installation](docs/installation.md) | Binary, source, and container install per OS |
| [CLI reference](docs/cli.md) | Every command, flag, and exit code |
| [Runefile language](docs/runefile.md) | Tasks, deps, params, caching, executors, dotenv |
| [Guides](docs/guides/README.md) | Task-oriented deep dives, one per capability |
| [AI agents (MCP)](docs/mcp.md) | Expose tasks to agents; the security model |
| [Docker](docs/docker.md) | Run Rune install-free in a container |
| [Troubleshooting](docs/troubleshooting.md) | Errors, diagnostics, and exit codes |
| [Grammar](docs/GRAMMAR.md) | The formal Runefile grammar |

## Contributing

See **[CONTRIBUTING.md](CONTRIBUTING.md)**. In short: tests run inside Docker
(`docker-compose run --rm test go test ./...`), and every change must pass the CI gates
(format, lint, cross-platform race tests, fuzz smoke, golden consistency, build). Rune
dogfoods itself — run `rune --list` in the repo to see the development tasks.

## Security

Found a vulnerability? Please follow the process in [SECURITY.md](SECURITY.md).

## License

[MIT](LICENSE) © The Rune Authors.
