# Go service

> **Use case:** the everyday tasks for a compiled Go service — fetch, build, test, run —
> wired together with dependencies and parameters.

**Demonstrates:** dependencies, parameters  ·  **Guide:** [Dependencies](../../runefile.md#dependencies-and-post-hooks)

**Prerequisites:** none (bodies echo the real `go` commands; swap them in for your project)

## Run it

```sh
rune test
```

## Expected output

```text
go mod download
go build -tags release -o dist/app ./cmd/app
go test ./...
```

`test` depends on `build`, which depends on `fetch`, so all three run in order (each once).
Try `rune run 9090` to pass a parameter, or `rune --list` to see every task.

## How it works

The `build target="release"` task takes a parameter with a default; `test: build` and
`build … : fetch` declare the dependency chain. See the `Runefile` in this directory.
