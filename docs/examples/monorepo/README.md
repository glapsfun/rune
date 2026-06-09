# Monorepo

> **Use case:** one Runefile for a repo of several services — shared steps via `import`,
> per-service tasks isolated in `mod` namespaces.

**Demonstrates:** imports, modules  ·  **Guide:** [Imports and modules](../../runefile.md#imports-and-modules)

**Prerequisites:** none

## Run it

```sh
rune api::build
```

## Expected output

```text
build api service
```

`rune web::build` builds the web service; `rune release` runs the shared `package` task
imported from `shared.rune`. Run `rune --list` to see both the root and namespaced tasks.

## How it works

- `import "shared.rune"` splices shared tasks (like `package`) into the root.
- `mod api "services/api.rune"` loads a service as the `api` namespace, called as `api::build`.

This directory ships `shared.rune` and `services/{api,web}.rune` so the example is runnable
as-is.
