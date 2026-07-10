# Quickstart & Validation: Minimum Rune Version

End-to-end validation scenarios that prove the feature works. Detailed behavior lives in
[contracts/](./contracts/) and [data-model.md](./data-model.md); this guide is the runnable
proof.

## Prerequisites

- Repo checked out on branch `010-minimum-rune-version`.
- Docker available (all Go tests run in the Docker harness per project policy).
- Feature implemented (see `tasks.md` after `/speckit-tasks`).

## Test commands (Docker only)

```sh
# Full suite
docker-compose run --rm test go test ./...

# Race + targeted packages
docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./internal/semver/... ./internal/config/... ./test/integration/...
```

Integration tests inject the installed version via the test-only env hook:

```sh
# Simulate an old binary against a Runefile requiring >= 0.8.0
RUNE_TEST_VERSION=0.7.2 rune build     # expect: incompatibility error, exit 3, nothing executed
RUNE_TEST_VERSION=0.8.0 rune build     # expect: runs normally
RUNE_TEST_VERSION=0.9.1 rune build     # expect: runs normally
```

## Scenario 1 — Reject an incompatible binary (US1 / P1)

1. Write a Runefile:
   ```rune
   set minimum_version := "0.8.0"

   build:
       echo built
   ```
2. Run with a simulated old binary: `RUNE_TEST_VERSION=0.7.2 rune build`.
3. **Expect**: the incompatibility diagnostic from [contracts/diagnostics.md](./contracts/diagnostics.md)
   — caret on `"0.8.0"`, installed/required/upgrade lines, exit code 3, `echo built` never runs.
4. Re-run with `RUNE_TEST_VERSION=0.8.0` and `0.9.1` → task runs; with no `minimum_version` → unchanged behavior.

## Scenario 2 — Static-value guard (US2 / P1)

1. Runefile with a dynamic value:
   ```rune
   required := env("RUNE_VERSION")
   set minimum_version := required
   ```
2. Run any task → **expect** `minimum_version must be a static semantic version`, caret on the value, exit 3.
3. Try `set minimum_version := "0.8"` and `set minimum_version := ">=0.8,<1.0"` → rejected as invalid semantic version.

## Scenario 3 — `rune version` / `--check` / `--json` (US3 / P2)

```sh
rune version                          # -> "rune <v>" + "runefile language 1"
RUNE_TEST_VERSION=0.8.3 rune version --check     # compatible status, exit 0
RUNE_TEST_VERSION=0.7.2 rune version --check     # incompatible, exit non-zero, no task run
RUNE_TEST_VERSION=0.8.3 rune version --check --json
# -> {"installed":"0.8.3","required":"0.8.0","compatible":true,"runefile":"/…/Runefile"}
```

Also verify `rune version --check` in a directory with no Runefile → reports "no requirement declared", exit 0.

## Scenario 4 — Override (US4 / P3)

```sh
RUNE_TEST_VERSION=0.7.2 rune --ignore-version build
# stderr: warning: ignoring Runefile minimum Rune version 0.8.0; running 0.7.2
# build runs
```

- Confirm no Runefile setting can enable the override.
- Confirm the MCP/agent path refuses by default when incompatible, and only proceeds when
  the operator sets `AllowIgnoreVersion`.

## Scenario 5 — Root ownership (edge / FR-012)

1. Root Runefile imports a child that declares `set minimum_version := "9.9.9"`; root declares `"0.8.0"` (or none).
2. Run with `RUNE_TEST_VERSION=0.8.0`:
   - Root declares `0.8.0` → runs (child's `9.9.9` ignored).
   - Root declares none → no gate (child's value does not impose a requirement).

## Cross-platform & gates

- Confirm CI runs the new tests on Linux, macOS, Windows.
- Confirm `docs-verify` passes after documenting the setting in `docs/GRAMMAR.md` and settings docs.
- Confirm golden diagnostic fixture matches a fresh regeneration.

## Definition of done (validation)

- [ ] Old/equal/new installed versions behave per Scenario 1.
- [ ] Non-static and invalid-semver values rejected per Scenario 2.
- [ ] `rune version`, `--check`, `--json` behave per Scenario 3.
- [ ] `--ignore-version` warns and proceeds; cannot be set from a Runefile; MCP default-refuses (Scenario 4).
- [ ] Imported files cannot change the effective requirement (Scenario 5).
- [ ] Unit + integration + golden + cross-platform tests pass; docs updated and docs-verify green.
