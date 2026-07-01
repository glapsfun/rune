# Use case: managing a Python project

> Wire Rune into a Python project so `install`, `test`, and friends run by name — and write a
> task body *in Python* when shell would be awkward. This walkthrough assumes you already have
> a Python project; adapt the echoed commands to your real ones.

**Backing example:** [`examples/python-project`](../examples/python-project/README.md) ·
**Features:** [dependencies](../how-to/dependencies-and-hooks.md), [executors (python)](../how-to/executors.md) ·
**Prerequisites:** `python3`

## The Runefile

```rune
# Tasks for a Python project. The `version` task body runs under the python
# executor (real Python); the rest echo the commands to adapt.

# Create a virtualenv and install dependencies.
install:
    @echo "python3 -m venv .venv && .venv/bin/pip install -r requirements.txt"

# Run the test suite (depends on install).
test: install
    @echo "pytest -q"

# Print the Python version — body runs under the python executor.
version (python):
    import sys; print("python", sys.version.split()[0])
```

## Run it

```sh
rune version
```

```text
python 3.12.x
```

The exact version reflects your installed Python. `rune test` runs `install` first (its
dependency), then the tests.

## Why it's written this way

- **`test: install`** declares a dependency, so `rune test` always sets up the environment
  before running — no forgetting the venv. See [Dependencies & post-hooks](../how-to/dependencies-and-hooks.md).
- **`version (python)`** picks the **python executor**: the body is real Python, run through
  `python3`, not shell. Use it when a task is genuinely easier in Python (parsing, data
  munging) than in shell. See [Executors](../how-to/executors.md).
- The `install`/`test` bodies **echo** their commands here so the example runs anywhere; swap
  in your real `pip`/`pytest` invocations.

> [!TIP]
> Keep secrets (PyPI tokens, etc.) in the environment, never in the Runefile — Rune reads
> them from the environment only.

---

**Next:** [Node project](node-project.md) · [Executors](../how-to/executors.md) · [Examples](../examples/README.md)
