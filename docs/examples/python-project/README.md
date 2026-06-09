# Python project

> **Use case:** capture the venv/pytest workflow for a Python project, and run a Python task
> body directly under the python executor.

**Demonstrates:** dependencies, executors (python)  ·  **Guide:** [Executors](../../runefile.md#executors)

**Prerequisites:** python3

## Run it

```sh
rune version
```

## Expected output

```text
python 3.12.x
```

(The exact version reflects your installed Python.) The `install`/`test` tasks echo the
commands — swap in the real ones for your project.

## How it works

`version (python)` runs its body as Python via the python executor (python3 must be installed),
shelling out through a temp file. The other tasks use the default shell executor.
