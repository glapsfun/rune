# Node / JavaScript project

> **Use case:** capture the npm workflow for a Node project, and run a JavaScript task body
> directly under the node executor.

**Demonstrates:** dependencies, executors (node)  ·  **Guide:** [Executors](../../runefile.md#executors)

**Prerequisites:** node

## Run it

```sh
rune version
```

## Expected output

```text
node v20.x.x
```

(The exact version reflects your installed Node.) The `install`/`build`/`test` tasks echo the
npm commands — swap in the real ones for your project.

## How it works

`version (node)` runs its body as JavaScript via the node executor (Node must be installed).
The other tasks use the default shell executor.
