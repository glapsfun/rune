# Contract: Versioning & Tags

## Format — FR-001

- Tags: `v` + **SemVer 2.0.0** → `vMAJOR.MINOR.PATCH` optionally `-rc.N` (e.g. `v0.0.1`,
  `v0.4.0-rc.2`).
- The in-binary version (`{{.Version}}`) is the tag without the leading `v`.

## Computation — FR-002/004/005

Given `latest` (highest existing `v*` tag, default `v0.0.0`) and `bump`:

| bump | result |
|------|--------|
| `major` | `vX+1.0.0` |
| `minor` | `vX.Y+1.0` |
| `patch` | `vX.Y.Z+1` |

If `prerelease`: append `-rc.N` where `N = (count of existing <next>-rc.* tags) + 1`.
The resulting tag MUST NOT already exist (refuse — FR-004).

## SemVer ↔ Conventional Commits (advisory) — research §B

- `fix:` → PATCH · `feat:` → MINOR · `feat!:`/`BREAKING CHANGE:` → MAJOR.
- **Pre-1.0 (0.x)**: a breaking change bumps **MINOR**, not MAJOR (SemVer §4 convention).
- Because the maintainer picks the bump, tooling only **warns** on an under-bump (e.g. a
  `feat!` present but `patch` chosen). It never auto-sets or blocks.

## Prerelease semantics — FR-005/020

- A `-rc.N` tag is auto-detected as a GitHub pre-release (`release.prerelease: auto`).
- A prerelease MUST NOT move `:latest`, nor update the Homebrew/Scoop stable channels.
