# Contract — Production Container Image

The official runtime image's "interface": how it is built, how users invoke it, and its
guarantees. Realizes FR-016/017/018/019, SC-005. Implemented in `Dockerfile` + `.dockerignore`,
documented in `docs/docker.md`.

## Build shape (multi-stage)

```text
Stage build:  golang:1.25-bookworm (matches go.mod major.minor)
  CGO_ENABLED=0  GOOS=linux
  go build -trimpath -ldflags "-s -w -X main.version=$VERSION -X main.commit=$COMMIT" \
           -o /out/rune ./cmd/rune
Stage final:  gcr.io/distroless/static-debian12:nonroot
  COPY --from=build /out/rune /rune
  USER nonroot
  WORKDIR /work
  ENTRYPOINT ["/rune"]
```

- Multi-arch: built for `linux/amd64` and `linux/arm64` via `docker buildx`.
- OCI labels: `org.opencontainers.image.{source,version,revision,licenses,title,description}`.
- No build toolchain, package manager, or shell in the final image (FR-017).

## Runtime interface

| Aspect | Contract |
|--------|----------|
| Entrypoint | `/rune` — args pass straight to the CLI (e.g. `docker run … rune --list`) |
| Working dir | `/work` — users bind-mount their project: `-v "$PWD":/work` |
| User | non-root (`nonroot`, uid 65532) |
| Stdout/stderr | identical streams/semantics to native (Rune messages → stderr; task output → stdout) |
| Exit codes | identical to native (CLI contract unchanged) |
| Version | `docker run … rune --version` prints `version (commit …)` via ldflags |

**Canonical usage** (documented in `docs/docker.md`):

```sh
docker run --rm -v "$PWD":/work ghcr.io/rune-task-runner/rune --list
docker run --rm -v "$PWD":/work ghcr.io/rune-task-runner/rune build
```

## Guarantees

- **Size**: final image **< 30 MB** (SC-005). Verified in CI/quickstart via
  `docker image inspect … --format '{{.Size}}'`.
- **`sh`-executor parity**: tasks whose bodies use only the pure-Go shell (`mvdan/sh`) and its
  builtins behave identically to a native run (FR-019, Principle V) — no system shell needed.
- **Static binary**: the embedded binary is `CGO_ENABLED=0`/static (Principle V); CGO is never
  enabled for the image build.

## Documented limitations (FR-019)

The minimal image intentionally omits external runtimes and system tools. Tasks that:
- shell out to external programs not present (e.g. `git`, `curl`), or
- use the `python` / `node` / agent executors,

will fail with a **clear, expected error** (missing executable), not a crash. Guidance in
`docs/docker.md`: bind-mount/native-install those tools, or build a fuller image FROM a
distro base — explicitly out of scope for the official minimal image (spec assumption).

## Test-harness image (unchanged, FR-020)

`Dockerfile.test` + `docker-compose.yml` remain the supported way to run the suite **inside a
container** (global policy: never on host): `docker compose run --rm test`. Distinct from the
production image (the harness intentionally includes the Go toolchain + python/node).
