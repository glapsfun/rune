# Quickstart: Validate Modern Docs & README Badges

How to prove this feature works end-to-end. All tests run **inside Docker** (project policy);
use `go run ./cmd/rune <task>` if `rune` isn't installed on the host.

## Prerequisites

- Repo checked out on branch `009-docs-and-badges`.
- Docker + standalone `docker-compose` (no compose plugin in this project).

## 1. Run the docs verification gate

The existing harness (extended by this feature) is the primary gate:

```sh
rune docs-check
# → docker-compose run --rm test go test ./test/docs/...
```

**Expected**: PASS. This asserts (see [`contracts/docs-structure.md`](./contracts/docs-structure.md)):
- every internal link in `docs/**/*.md` + `README.md` resolves (reorg + redirect stubs);
- fenced `rune` blocks on self-contained pages validate;
- every backing example validates (Tier A) and runs where its interpreter exists (Tier B);
- example READMEs satisfy the example contract;
- **badge integrity** (`badges_test.go`): canonical repo/module targeting, no placeholders,
  alt text present, image wrapped in the correct link, real `ci.yml` referenced.

## 2. Prove the accuracy gate actually fails on drift (US4)

Confirm enforcement is real, not decorative:

```sh
# Temporarily break an internal link in any docs page, then:
rune docs-check      # → FAIL (broken internal link reported)
# Revert the change; re-run → PASS.
```

Repeat with a badge URL pointed at a wrong owner (e.g. `USER/REPO`) → `badges_test.go` FAILS.

## 3. Non-regression: CLI output unchanged (SC-008)

```sh
docker-compose run --rm test go test ./...        # full suite incl. golden
docker-compose run --rm -e CGO_ENABLED=1 test go test -race ./...
```

**Expected**: PASS with **no golden updates** — this feature changes docs + `README` only.

## 4. Manual render checks (GitHub-only surfaces)

These can't be unit-tested; verify visually on GitHub (push the branch and open it):

- **US1 badges**: the `README` badge row renders above the fold; every badge shows live state
  (CI green, tag `v0.1.0`, MIT, Go version, Report Card grade, Go Reference, Docs); each links
  to the right source; legible in light **and** dark theme; scannable on mobile width.
  - First visit: click the **Go Report Card** and **Go Reference** links once so those
    providers index/compute (they populate on first request).
- **US2 use-cases**: open `docs/use-cases/python-project.md` (then node, then mcp-agents) and,
  following only that page, copy the example and run its task — succeeds first try, output
  matches (SC-003). MCP page states the security posture.
- **US3 navigation**: from `docs/README.md`, reach any target page in ≤2 clicks (SC-007);
  spot-check GitHub Alerts and `<details>` render correctly; every page ends with a
  next-steps footer.
- **Stubs**: opening an old path (e.g. `docs/guides/caching.md`) shows the "Moved" pointer.

## Success = 

`rune docs-check` and the full test suite pass with zero golden changes; the four manual
render checks above hold; and the accuracy gate demonstrably fails when a link or badge is
intentionally broken. Maps to SC-001..SC-008 in [`spec.md`](./spec.md).
