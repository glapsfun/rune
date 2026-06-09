# Quickstart — Validating the Best-Practices Refactor

Runnable scenarios that prove each user story is delivered. These are **validation/run
scenarios**, not implementation. Per global policy, the Go test suite runs **inside Docker**
(`docker compose run --rm test …`), never on the host. Commands assume repo root.

Cross-references: gate semantics → `contracts/ci-gates.md`; image → `contracts/docker-image.md`;
release → `contracts/release-pipeline.md`; docs coverage → `contracts/docs-structure.md`.

---

## Prerequisites

- Docker (with buildx) — for the test harness, the production image, and release dry-run.
- `git`. (A local Go toolchain is optional; CI derives it from `go.mod`.)

---

## US1 — Adopt Rune from documentation (P1)

**Goal**: a newcomer reaches a running task using only docs (SC-001), and every command/construct
is documented (SC-010).

```sh
# 1. README exists and links the docs set
test -f README.md && grep -q "docs/getting-started.md" README.md

# 2. All required docs pages exist
for f in installation getting-started cli runefile mcp docker; do
  test -f "docs/$f.md" || echo "MISSING docs/$f.md"
done

# 3. The getting-started example actually runs (freshness, FR-006) — via dogfood task
docker compose run --rm test go run ./cmd/rune --file docs/examples/getting-started/Runefile <task>
#   → exits 0 and prints the documented output

# 4. Coverage spot-check: every global flag from main.go appears in docs/cli.md
#   (CI/docs-check enforces; manual: compare flag names to docs/cli.md)
```

**Expected**: all files present; the example task exits 0 with the documented output; no flag or
construct is undocumented.

---

## US2 — Constitutional CI gates (P1)

**Goal**: every change is gated; violations are blocked and named (SC-002/003/004).

```sh
# Reproduce each gate locally (the same commands CI runs):
gofmt -l .                                   # → empty
golangci-lint run                            # → no findings (incl. gofumpt/goimports)
go vet ./...                                  # → clean
docker compose run --rm test go test -race ./...        # → PASS, 0 races (CGO on in container)
CGO_ENABLED=0 go build ./...                 # → builds (static)
# Golden guard:
go test ./internal/diag ./internal/lexer ./internal/parser ./internal/cli ./test/corpus -update
git diff --exit-code testdata/ docs/GRAMMAR.md          # → no diff
# Release config resolves:
goreleaser release --snapshot --clean        # → succeeds; archives include LICENSE+README
```

**Independent test (US2)**: open a PR that (a) misformats a file, (b) adds a lint violation, (c)
breaks a test, (d) introduces a data race, (e) hand-edits a golden — each must turn the matching
job red and **block merge**; a clean PR passes all jobs. Verify the failing job name identifies
the gate (SC-003).

---

## US3 — Run Rune via Docker (P2)

**Goal**: install-free usage; minimal image; documented limits (SC-005, FR-016/017/019).

```sh
# Build the production image (multi-stage, distroless static)
docker buildx build -t rune:local --load .

# Size budget < 30 MB (SC-005)
docker image inspect rune:local --format '{{.Size}}'   # → < 30000000

# Runs against a mounted project, parity with native sh executor
docker run --rm -v "$PWD":/work rune:local --version    # → version (commit …)
docker run --rm -v "$PWD":/work rune:local --list        # → same listing as native

# Documented limitation: a python/node/missing-tool task fails CLEANLY (not a crash)
docker run --rm -v "$PWD":/work rune:local <python-task> # → clear "executable not found" error

# Test harness still works in-container (FR-020)
docker compose run --rm test go test ./...
```

**Expected**: image builds, is <30 MB, runs as non-root, `sh`-executor output matches native, and
missing-runtime tasks fail with a clear documented message.

---

## US4 — Best-practice structure & code (P2)

**Goal**: hygiene files present, zero lint violations, behavior preserved (SC-008/009, FR-024).

```sh
# Hygiene files exist (and resolve .goreleaser refs)
for f in LICENSE CONTRIBUTING.md SECURITY.md .editorconfig .dockerignore Runefile; do
  test -e "$f" || echo "MISSING $f"
done

# Zero discipline violations (SC-008)
golangci-lint run                            # → no issues

# Behavior preserved (SC-009): full suite green WITHOUT regenerating any golden
docker compose run --rm test go test ./...   # → PASS
git status --porcelain testdata/             # → empty (no golden changed)

# Dogfood: the project runs its own tasks
docker compose run --rm test go run ./cmd/rune --list      # shows fmt/lint/test/build/docker…
```

**Expected**: all hygiene files present; lint clean; entire existing suite passes with **no**
golden file modified (proves zero behavior change); the root `Runefile` lists the dev tasks.

---

## US5 — One-tag release (P3)

**Goal**: a single tag yields binaries + checksums + published image (SC-007, FR-025/026/027).

```sh
# Dry-run locally (no publish) — proves the artifact set & file refs
goreleaser release --snapshot --clean
ls dist/                                     # → archives per OS/arch + checksums.txt
tar tzf dist/rune_*_linux_amd64.tar.gz | grep -E 'LICENSE|README'   # → both present

# Real release (maintainer): pushing a tag triggers .github/workflows/release.yml
git tag v0.1.0 && git push origin v0.1.0
#   → GitHub Release with archives + checksums, and ghcr.io/rune-task-runner/rune:v0.1.0 (+latest)
```

**Expected**: dry-run produces the full artifact set with LICENSE+README in each archive; the
tag-triggered workflow publishes the GitHub Release and the multi-arch GHCR image with no manual
assembly. A missing referenced file fails the dry-run before any publish (SC-006).

---

## Done-when (maps to Success Criteria)

- [ ] SC-001 newcomer → first task in <10 min via docs · [ ] SC-002 all changes gated
- [ ] SC-003 violations blocked + gate named · [ ] SC-004 `-race` green on 3 OSes
- [ ] SC-005 image <30 MB, install-free · [ ] SC-006 release dry-run resolves all refs
- [ ] SC-007 one tag → full release · [ ] SC-008 zero lint/discipline violations
- [ ] SC-009 suite green, **no** golden regenerated · [ ] SC-010 every command/construct documented
