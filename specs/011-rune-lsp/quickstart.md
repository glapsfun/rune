# Quickstart & Validation: Rune Language Server Protocol

Runnable scenarios that prove the feature works end to end. All Go tests run **inside Docker Compose** (project policy), never on the host.

## Prerequisites

- Repo checked out on branch `011-rune-lsp`.
- Docker + standalone `docker-compose` (per `CLAUDE.md` / `CONTRIBUTING.md`).
- Build the binary for manual/editor checks: `go run ./cmd/rune …` or `docker-compose run --rm test go build -o /tmp/rune ./cmd/rune`.

## 1. `rune analyze` parity & exit codes (User Story 2)

Create `testdata/lsp/basic/Runefile`:

```
output := "dist"
# Build the application.
build target="debug":
    go build -tags {{target}} -o {{output}}/app ./...
deploy env: missing
    ./deploy.sh {{env}}
```

Run:

```sh
rune analyze testdata/lsp/basic/Runefile
```

**Expect**: a line `…:5:13: error[RUNE2001]: unknown dependency "missing"`, a summary count, and **exit code 3**. Fix `missing` → `build` and re-run: no errors, **exit 0**. Add `--json`: structured diagnostics with the same `code`/`range`. Confirm no task ran (no `go build` / `deploy.sh` output).

## 2. Live diagnostics in an editor (User Story 1)

Point any LSP client at `rune lsp` (see §8). Open the fixture above.

**Expect**: an error underline on `missing`. Edit it to `build`; the error clears within the debounce window. Introduce an unterminated string (`target="`) mid-line; surrounding valid tasks still analyze and the server neither crashes nor hangs.

## 3. Completion (User Story 3) — the MVP acceptance

With the file from the MVP definition open, type `deploy env: bui` and request completion.

**Expect**: `build` suggested, showing its parameters (`target="debug"`) and doc comment. Accept → text becomes `deploy env: build`. Also verify: inside `{{ }}` suggests `output`/params; after `set ` suggests `working-directory`; inside `[` suggests attributes; `(py` in executor position suggests `python`.

## 4. Go-to-definition, incl. cross-file (User Story 4)

Single file: invoke definition on `build` in `deploy: build` → jumps to the `build` task.

Cross-file: use `testdata/lsp/imports/` (a root that imports a `backend` module defining `build`). Invoke definition on `backend.build` → opens the imported file at the `build` declaration. With the imported file open and edited but unsaved, definition uses the overlay content.

## 5. Hover & symbols (User Stories 5–6)

Hover `build` → panel shows signature, doc, executor, group, and `Defined in: …`. Open the outline → settings, variables, imports, tasks listed under their categories.

## 6. Formatting (User Story 7)

Take a poorly-formatted unsaved buffer, request formatting.

**Expect**: the returned edit's result equals `internal/formatter.Format` output for that content, and formatting already-formatted content is a no-op. Confirm (via `--log-level debug`) that no child process was spawned and no file was written by the server.

## 7. Safety & robustness (SC-004, SC-005)

- Run the fuzz targets briefly (parser recovery, position conversion, edit application, malformed JSON-RPC):
  ```sh
  docker-compose run --rm test go test -run x -fuzz FuzzParseRecover -fuzztime 30s ./internal/parser
  ```
  **Expect**: no panic, termination, no out-of-bounds ranges.
- Run the no-side-effects test: it fails if any analyze/LSP operation runs a task, spawns a process, opens a socket, or writes a project file.

## 8. Protocol integration (SC-009)

Drive a real subprocess through the JSON-RPC lifecycle:

```sh
docker-compose run --rm test go test ./internal/lsp -run TestProtocol
```

Sequence exercised: `initialize` → `initialized` → `didOpen` → `completion` → `definition` → `hover` → `shutdown` → `exit`. **Expect**: correct responses, advertised capabilities match `contracts/cli-lsp.md`, clean exit.

## 9. Editor setup (User Story 8)

- **VS Code**: install/point the `editors/vscode` extension at the built binary; open a Runefile; confirm diagnostics + completion.
- **Neovim / Zed / Helix**: follow `editors/README.md`; each configures the command `rune lsp` for the `runefile` filetype.

## Full suite gate

```sh
docker-compose run --rm test go test ./...
docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...
```

All existing gates (lint, golden, fuzz-smoke, docs-verify, release-dryrun) plus the new LSP tests must pass. Reference the [diagnostic-code catalog](./contracts/diagnostic-codes.md), [CLI contracts](./contracts/), and [data model](./data-model.md) for details rather than duplicating them here.
