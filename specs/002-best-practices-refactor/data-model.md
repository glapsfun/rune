# Phase 1 Data Model — Deliverable Artifacts

This feature has **no runtime/domain data model** (it changes no Rune behavior and stores no
data). The "entities" are the **deliverable artifacts** the feature produces or modifies, plus
their validation rules and lifecycle states. This is the structured form of the spec's *Key
Entities* section, used to drive `tasks.md`.

## Artifact inventory

| # | Artifact | Path | New/Mod | Story | Validated by |
|---|----------|------|---------|-------|--------------|
| A1 | README | `README.md` | new | US1 | SC-001, FR-001 |
| A2 | Install guide | `docs/installation.md` | new | US1 | FR-002 |
| A3 | Getting-started | `docs/getting-started.md` | new | US1 | SC-001, FR-003 |
| A4 | CLI reference | `docs/cli.md` | new | US1 | SC-010, FR-004 |
| A5 | Runefile language guide | `docs/runefile.md` | new | US1 | SC-010, FR-004 |
| A6 | MCP/agent guide | `docs/mcp.md` | new | US1 | FR-005 |
| A7 | Docker guide | `docs/docker.md` | new | US1/US3 | FR-018, FR-019 |
| C1 | CI workflow | `.github/workflows/ci.yml` | mod | US2 | SC-002/003/004, FR-007…015 |
| C2 | Lint/format config | `.golangci.yml` | mod | US2/US4 | SC-008, FR-008, FR-022 |
| D1 | Production image | `Dockerfile` | new | US3 | SC-005, FR-016/017/019 |
| D2 | Docker ignore | `.dockerignore` | new | US3 | FR-017 |
| D3 | Test harness | `Dockerfile.test`, `docker-compose.yml` | keep | US3 | FR-020 |
| S1 | License | `LICENSE` | new | US4 | FR-024, FR-027, SC-006 |
| S2 | Contributor guide | `CONTRIBUTING.md` | new | US4 | FR-024 |
| S3 | Security policy | `SECURITY.md` | new | US4 | FR-024 (Principle VII) |
| S4 | Editor config | `.editorconfig` | new | US4 | FR-022 |
| S5 | Dogfood Runefile | `Runefile` | new | US4 | research §9 |
| S6 | Code hygiene pass | `internal/**`, `cmd/**`, `mcpserver/**` | mod | US4 | SC-008, SC-009, FR-021/022/023 |
| R1 | Release workflow | `.github/workflows/release.yml` | new | US5 | SC-007, FR-025/026 |
| R2 | Release config | `.goreleaser.yaml` | mod | US5 | FR-027, SC-006 |

## Per-artifact rules & states

### Documentation set (A1–A7)
- **Rules**: every CLI command/flag (A4) and every Runefile construct (A5) has ≥1 runnable
  example (SC-010); examples use forward-slash paths (Principle V); `GRAMMAR.md` stays the
  formal grammar (A5 links, does not duplicate); MCP doc (A6) states the secure-by-default
  posture (read-only default, env-only secrets, destructive opt-in).
- **States**: `absent → drafted → example-verified` (getting-started example executed in CI /
  `Runefile docs-check`) → `linked from README`.

### CI workflow (C1) & lint config (C2)
- **Rules**: jobs = lint, test (race matrix, CGO on), build (CGO off matrix), golden, fuzz-smoke,
  release-dryrun (research §2); Go version from `go-version-file: go.mod` (FR-014); a failing
  job blocks merge and names the gate (SC-003); C2 enables `gofumpt`+`goimports` (FR-008).
- **States**: `current → augmented → verified` (a deliberately-broken PR is blocked per gate; a
  clean PR passes — Independent Test for US2).

### Container images (D1–D3)
- **Rules**: D1 is multi-stage, final base `distroless/static-debian12:nonroot`, `CGO_ENABLED=0`
  binary, non-root, `ENTRYPOINT ["/rune"]`, multi-arch (amd64+arm64), OCI labels; image **< 30
  MB** (SC-005); `sh`-executor parity with native (FR-019); missing-runtime tasks fail with a
  clear documented error. D3 retained as the in-container test path (FR-020).
- **States**: `test-harness-only → production-image-built → size-verified → published (R1)`.

### Repo hygiene & refactor (S1–S6)
- **Rules**: `LICENSE` + `README` exist so `.goreleaser` refs resolve (SC-006); code is
  `gofumpt`/`goimports`/`golangci-lint`-clean with **zero** violations (SC-008); the hygiene
  pass is **behavior-preserving** — full suite green with **no** golden `-update` (SC-009);
  package layout matches the constitution (FR-021, already true — verify only); `Runefile`
  uses only existing DSL features.
- **States**: `gaps-present → files-added → formatted → lint-clean → behavior-verified`.

### Release pipeline (R1–R2)
- **Rules**: tag `v*` → archives (linux/macos/windows × amd64/arm64) + `checksums.txt` +
  GHCR image tagged with the version (SC-007, FR-025/026); archives include LICENSE+README
  (FR-027); a snapshot dry-run fails on any unresolved file ref before publish (SC-006);
  GHCR auth via `GITHUB_TOKEN` only — no added secrets (Principle VII).
- **States**: `config-present-untriggered → workflow-wired → dry-run-green → tag-publishes`.

## Relationships / ordering constraints

- **S1 (LICENSE) precedes R2/R1**: release archives reference `LICENSE`/`README`, so S1+A1 must
  exist before the release dry-run can pass (SC-006).
- **C2 precedes the C1 lint job and S6**: the formatter set must be defined before the lint gate
  and the repo-wide formatting pass are meaningful.
- **D1 precedes R1's image-publish step**: the production image must build before it can be
  pushed.
- **A3/A4/A5 underpin SC-001/SC-010**: docs completeness gates the docs success criteria; A7
  depends on D1 existing (documents how to run the image).
- **S6 is gated by the existing test suite** (the behavior-preservation oracle), not by new tests.
