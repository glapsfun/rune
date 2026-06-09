# Running Rune in Docker

You can run Rune with no local Go toolchain using the official container image. The image
is a minimal, non-root, static-binary container (distroless base) — just the `rune` binary,
no shell or package manager.

## Quick start

Mount your project at `/work` (the image's working directory) and pass any normal Rune
arguments after the image name:

```sh
# List tasks in the current project:
docker run --rm -v "$PWD":/work ghcr.io/rune-task-runner/rune --list

# Run a task:
docker run --rm -v "$PWD":/work ghcr.io/rune-task-runner/rune build

# Pass arguments / variable overrides:
docker run --rm -v "$PWD":/work ghcr.io/rune-task-runner/rune greet Ada
docker run --rm -v "$PWD":/work ghcr.io/rune-task-runner/rune build target=release
```

The container's `ENTRYPOINT` is `/rune`, so everything after the image name goes straight to
the CLI. Behavior for the default `sh` executor is identical to a native run — the shell is
the pure-Go `mvdan.cc/sh` interpreter baked into the binary.

## Tags

- `ghcr.io/rune-task-runner/rune:latest` — latest release
- `ghcr.io/rune-task-runner/rune:<version>` — a specific release (e.g. `:1.2.3`)

## Building it yourself

```sh
docker buildx build -t rune:local --load .

# With version metadata (so `rune --version` reports it):
docker buildx build \
  --build-arg VERSION="$(git describe --tags --always)" \
  --build-arg COMMIT="$(git rev-parse --short HEAD)" \
  -t rune:local --load .
```

## Limitations of the minimal image

The image intentionally contains **only** the `rune` binary — no system shell, interpreters,
or tools. That keeps it tiny and secure, but it means:

- **`sh`-executor tasks** that use only the built-in pure-Go shell work as usual.
- **Tasks that shell out to external programs** (e.g. `git`, `curl`, compilers) will fail
  with a clear "executable not found" error — those programs are not in the image.
- **`python` / `node` / `agent` executor tasks** require their runtimes, which are not
  present. They will fail with a clear error rather than crash.

If you need system tools or language runtimes, either:

1. Run Rune **natively** (see the [installation guide](installation.md)), or
2. Build a fuller image `FROM` a distro base (e.g. `debian:bookworm-slim`) and copy the
   `rune` binary into it alongside the tools you need.

## Running Rune's own test suite (contributors)

The repository ships a separate Docker-based **test harness** (not the production image).
Per project policy the Go test suite runs inside the container, never on the host:

```sh
docker-compose run --rm test go test ./...
docker-compose run --rm test go test -race ./...
docker-compose run --rm test go vet ./...
```

See [CONTRIBUTING.md](../CONTRIBUTING.md) for the full development workflow.
