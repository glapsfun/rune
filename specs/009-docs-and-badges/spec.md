# Feature Specification: Modern, Example-Rich Documentation & README Status Badges

**Feature Branch**: `009-docs-and-badges`

**Created**: 2026-07-01

**Status**: Draft

**Input**: User description: "feature moder fancy docs ( examples , how to, userguid , exaples with fature + use cases like mangin python project, node project , examples with mcp) + update README.md with badge abot ci tag , release etc need to do deep reserch."

## Overview

Rune already ships a working documentation set — a top-level `README`, a `docs/` folder of
guides, and a 15-entry example library covering common project shapes. This feature raises
that set to a **modern, example-rich handbook that reads beautifully on GitHub** and gives
the project a credible **front door**: a `README` topped with live status badges (build,
release, license, Go version, code quality, docs, and Go Reference).

The documentation lives as **in-repo markdown** rendered on GitHub (no hosted website). The
"modern/fancy" bar is met through *structure and presentation*, not a site generator:
goal-oriented **how-to recipes**, a readable **user guide**, and **use-case walkthroughs**
for the shapes people actually manage — a **Python project**, a **Node project**, and
**AI-agent access via MCP** — each anchored to a runnable example whose output is shown.

The unit of value is **reader understanding and first-glance credibility**. A stranger
should size up the project's health in seconds from the README badges, then find the exact
page for their goal and copy-run their way to success without hunting across the repo.

This is a documentation-and-README feature. It changes explanatory artifacts only. It does
**not** change Rune's behavior, language, CLI output, or the shipped binary.

## Clarifications

### Session 2026-07-01

- Q: How should the new `how-to/`, `user-guide/`, `use-cases/` sections relate to the existing `docs/` layout (flat pages + `docs/guides/`)? → A: Reorganize in place — fold `guides/` into the new `user-guide/` + `how-to/` layout, keep flat pages, update all internal links into one unified structure.
- Q: Should docs accuracy (examples run, internal links resolve) be enforced going forward or verified once? → A: Enforce as an ongoing CI gate — extend the existing `docs-check` gate to run referenced examples and check internal links on every change.
- Q: How far should the README "modern/fancy" makeover go beyond badges? → A: Badges + a light header refresh — centered title/tagline, badge row, quick-nav links; keep existing prose and tables; no image/logo assets.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Size up the project in seconds from the README (Priority: P1)

An evaluator (a developer deciding whether to adopt Rune, or a maintainer scanning the repo)
opens the `README` on GitHub. Above the fold, a row of status badges tells them at a glance
that the build is green, what the latest released version is, the license, the Go version,
the code-quality grade, and where the docs and API reference live — each badge links to its
source. They form an accurate impression of the project's health and maturity before reading
a single paragraph.

**Why this priority**: Badges are the smallest, most self-contained, highest-visibility
slice of the request and the project's credibility signal. They ship independently of any
docs restructuring and deliver value immediately. This is the irreducible front door.

**Independent Test**: Open the rendered `README` on GitHub with no prior context; confirm the
badge row is present above the fold, every badge resolves to live state (no placeholder or
broken image), each badge links to the correct source, and a reader can state the build
status, latest version, and license from the badges alone.

**Acceptance Scenarios**:

1. **Given** the rendered `README`, **When** a reader looks at the top, **Then** a badge row
   shows CI/build status, latest release/version (git tag), license, Go version, code-quality
   grade, a docs link, and a Go Reference link — each as a clickable badge.
2. **Given** any status badge, **When** the underlying state changes (CI run completes, a new
   release is tagged), **Then** the badge reflects the new state automatically without a
   manual `README` edit.
3. **Given** a status/CI/release/license badge, **When** a reader clicks it, **Then** it
   navigates to the correct source for the canonical repo (Actions run list, releases page,
   `LICENSE`, etc.).
4. **Given** the `README` viewed in GitHub's light or dark theme, **When** it renders,
   **Then** all badges remain legible and correctly aligned.

---

### User Story 2 - Follow a use-case walkthrough for my kind of project (Priority: P2)

A task author has a concrete job: wire up tasks for a **Python project**, a **Node project**,
or **expose tasks to an AI agent over MCP**. They open the matching use-case walkthrough, see
a complete, self-contained scenario written around a real project shape, copy the example
`Runefile`, run the commands shown, and get the predicted output — succeeding on the first
attempt without jumping to other pages. Each walkthrough pairs concrete **features** (params,
caching, dependencies, executors, MCP exposure) with the **use case** so the reader learns
why the file is written the way it is.

**Why this priority**: This is the heart of the request — "user guide + use cases like
managing a Python project, Node project, examples with MCP." Use-case walkthroughs are how
people actually adopt a task runner. It builds on the front door (P1) but delivers standalone
value: a Python/Node/MCP user can succeed from these pages alone.

**Independent Test**: Give a developer only the Python (or Node, or MCP) use-case page;
confirm they can copy the example, run the tasks, and reach the shown output on the first
try, and can explain which Rune features the walkthrough used and why.

**Acceptance Scenarios**:

1. **Given** the documentation set, **When** a reader looks for their project type, **Then**
   there are use-case walkthroughs for at least a Python project, a Node project, and MCP /
   AI-agent access, each linked from a discoverable index.
2. **Given** a use-case walkthrough, **When** the reader follows it top to bottom, **Then**
   every command and file shown is copy-paste runnable, is anchored to a runnable example in
   the example library, and shows the expected output alongside it.
3. **Given** the MCP use-case walkthrough, **When** a reader follows it, **Then** it shows how
   tasks become agent-callable tools and states the default security posture (read-only by
   default, gated destructive tasks, env-only secrets) at the point it matters.
4. **Given** a walkthrough, **When** the reader finishes, **Then** it links to the relevant
   deeper guide and to related examples rather than dead-ending.

---

### User Story 3 - Learn any capability from a how-to recipe or the user guide (Priority: P2)

A reader who knows what Rune is wants to accomplish a specific thing ("how do I cache a slow
task?", "how do I pass parameters?", "how do I run tasks in parallel?") or read a coherent
top-to-bottom tour of the tool. They find a **how-to** section of short, goal-titled recipes
for quick answers, and a **user guide** that presents the capabilities as a readable,
ordered narrative. Both are consistently structured, cross-linked, and never leave the reader
guessing at the next step.

**Why this priority**: Rounds out "how to" + "user guide" from the request and turns the
scattered guides into a navigable handbook. Independently valuable even without the use-case
walkthroughs, but lower priority than the specific project walkthroughs people asked for.

**Independent Test**: Pick three arbitrary goals; confirm each maps to a how-to recipe or a
user-guide section reachable from the docs index in at most two clicks, each with runnable,
output-showing snippets and links onward.

**Acceptance Scenarios**:

1. **Given** the docs index, **When** a reader has a goal in mind, **Then** they can reach a
   how-to recipe or user-guide section for it in ≤2 clicks via a goal-oriented ("I want to…")
   index.
2. **Given** any documentation page, **When** it is read, **Then** it follows a consistent
   structure (clear title, short intro, a contents list on long pages, and a next-steps
   footer) and cross-links related pages and examples.
3. **Given** any code snippet whose result is meaningful, **When** it appears in the docs,
   **Then** the expected output is shown next to it.
4. **Given** notes, warnings, or tips, **When** they appear, **Then** they use GitHub-rendered
   callout/admonition markdown so they stand out from body text.

---

### User Story 4 - Trust that the docs are accurate and won't rot (Priority: P3)

A reader (and a maintainer) can trust that what the docs claim matches the real tool: every
runnable example actually runs, every internal link resolves, and every badge points at the
right repository. Stale output, broken links, and dead badges are caught rather than shipped.

**Why this priority**: Protects the value delivered by P1–P3 over time. Independently
valuable as a quality gate, but lower priority than producing the content itself.

**Independent Test**: Run the docs-check gate (extended to execute the documentation's example
set against the actual tool and link-check the docs); confirm all examples succeed, all
internal links resolve, and every badge URL targets the correct repo/module — and that the
gate fails when an example or link is intentionally broken.

**Acceptance Scenarios**:

1. **Given** every runnable example referenced by the docs, **When** it is executed against
   the actual tool, **Then** it succeeds and produces the output shown in the docs.
2. **Given** the documentation set, **When** internal links are checked, **Then** none are
   broken.
3. **Given** the badge row, **When** each badge's target is checked, **Then** it points at the
   canonical repository/module and not a placeholder.

---

### Edge Cases

- **No release / red CI**: badges must show the *real* state (e.g. "no releases yet", a red
  failing badge) rather than a broken image or a stale hardcoded value.
- **Provider not yet indexed**: code-quality and Go Reference badges must degrade to a
  pending/neutral state (not a broken image) if the module isn't indexed yet.
- **Badge provider unavailable / rate-limited**: a badge that fails to load must fall back to
  readable alt text and still link to its source.
- **Repo identity split**: the public repo (`glapsfun/rune`) and the Go import path
  (`rune-task-runner/rune`) differ; each badge must target the correct one for its purpose.
- **README viewed off GitHub** (raw file, IDE preview, package mirrors): badges degrade to alt
  text/links and the page stays readable.
- **Light vs dark theme**: badges and any inline images stay legible in both GitHub themes.
- **Example drift**: an example `Runefile` changes and the output shown in a walkthrough goes
  stale — this must be catchable, not silently wrong.
- **Long pages**: a page long enough to require scrolling provides an in-page contents list so
  readers don't lose their place.

## Requirements *(mandatory)*

### Functional Requirements

#### README & status badges

- **FR-001**: The `README` MUST display, near the top, a badge row covering: continuous-
  integration/build status, latest release/version (git tag), license, Go version, code-
  quality grade, a documentation link, and a Go Reference (package documentation) link.
- **FR-002**: Status badges (CI, release/version) MUST reflect live state automatically as CI
  runs complete and releases are tagged — no manual `README` edit per change.
- **FR-003**: Each badge MUST link to its authoritative source (CI → workflow runs, release →
  releases page, license → `LICENSE`, code quality → the grading report, Go Reference → the
  package docs, docs → the in-repo documentation index).
- **FR-004**: Badges MUST target the project's canonical GitHub repository for repo-scoped
  badges and the canonical Go module path for module-scoped badges (see Assumptions).
- **FR-005**: Every badge MUST degrade gracefully when its state is empty or its provider is
  unavailable (readable alt text, valid link, no broken-image placeholder).
- **FR-006**: The `README` MUST render correctly on GitHub (badges, tables, callouts,
  collapsible sections) in both light and dark themes, and remain scannable above the fold.
- **FR-006a**: The `README` MUST receive a light header refresh — a centered title/tagline, the
  badge row, and quick-navigation links to the docs — while keeping the existing prose and
  tables. No image, logo, or banner assets are introduced (keeps rendering theme- and
  mobile-safe with nothing to maintain).

#### Documentation content

- **FR-007**: The documentation MUST include a **how-to** section of short, goal-titled
  recipes ("how do I …") that give quick, task-focused answers.
- **FR-008**: The documentation MUST include a **user guide** that presents Rune's
  capabilities as a coherent, ordered, readable narrative.
- **FR-009**: The documentation MUST include **use-case walkthroughs** for at least: managing
  a **Python project**, managing a **Node project**, and **exposing tasks to AI agents via
  MCP**.
- **FR-010**: Each use-case walkthrough MUST be anchored to a runnable example in the example
  library, be copy-paste runnable, and show the expected output alongside the commands.
- **FR-011**: Each use-case walkthrough MUST pair concrete Rune features (e.g. parameters,
  caching, dependencies, executors, MCP exposure) with the use case and explain *why* the
  example is written the way it is.
- **FR-012**: The MCP use-case walkthrough MUST show how tasks become agent-callable tools and
  state the default security posture (read-only by default, gated destructive tasks, env-only
  secrets) where it is relevant.

#### Navigation & presentation

- **FR-013**: The documentation MUST provide a goal-oriented index ("I want to…") from which a
  reader can reach the right how-to, guide, use case, or example in ≤2 clicks.
- **FR-014**: Every documentation page MUST follow a consistent structure: a clear title, a
  short intro, an in-page contents list on long pages, and a next-steps footer that links
  onward rather than dead-ending.
- **FR-015**: Every code snippet whose result is meaningful MUST show its expected output next
  to it, and related pages/examples MUST be cross-linked.
- **FR-016**: Notes, warnings, and tips MUST use GitHub-rendered callout/admonition markdown so
  they are visually distinct from body text.

#### Accuracy & non-regression

- **FR-017**: Every runnable example referenced by the docs MUST execute successfully against
  the actual tool and produce the output shown, and this MUST be enforced as an ongoing gate
  (extending the existing `docs-check` verification) so failures are caught on every change.
- **FR-018**: All internal documentation links MUST resolve (no broken links), enforced as an
  ongoing gate (extending the existing `docs-check` verification) rather than a one-time check.
- **FR-019**: This feature MUST NOT change Rune's behavior, language, CLI output, or the
  shipped binary — it changes documentation and the `README` only; existing golden/byte-exact
  CLI output stays identical.

### Key Artifacts *(documentation-only; no runtime data)*

- **README**: The repo front door; a badge row plus the value proposition and links into the
  docs.
- **Status badge**: A single live indicator (CI, release, license, Go version, code quality,
  docs, Go Reference), each linking to its source.
- **How-to recipe**: A short, goal-titled page answering one "how do I …" question.
- **User-guide chapter**: One section of the ordered, readable capability tour.
- **Use-case walkthrough**: A complete, project-shaped scenario (Python, Node, MCP) anchored
  to a runnable example.
- **Example**: A self-contained, runnable `Runefile` (+ notes) in the example library.
- **Docs index / navigation**: The goal-oriented map that connects all of the above.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A first-time reader can state the build status, latest version, and license of
  the project within 10 seconds of opening the `README`, using the badges alone.
- **SC-002**: Every `README` badge resolves without error (zero placeholder or broken-image
  badges), and the status badges (CI, release/version) reflect live state — consistent with
  FR-002 (the static license and docs badges have no changing state).
- **SC-003**: A Python, Node, or MCP user can go from opening the relevant use-case
  walkthrough to a successfully running task by copy-pasting, on the first attempt, without
  consulting any other page.
- **SC-004**: 100% of runnable examples referenced by the docs execute successfully against
  the tool, and every output-bearing snippet shows its expected output.
- **SC-005**: Zero broken internal documentation links.
- **SC-006**: The `README` and docs render correctly on GitHub across desktop and mobile
  widths and in both light and dark themes (no broken tables, badges, or layout).
- **SC-007**: A reader can reach the correct page for a stated goal (e.g. "manage a Python
  project", "expose tasks to an agent") from the docs index in ≤2 clicks.
- **SC-008**: The feature introduces zero changes to tool behavior or CLI output — the
  existing golden-output test suite passes unchanged.

## Assumptions

- **Rendering surface**: Documentation is in-repo markdown viewed on GitHub; there is **no
  hosted documentation website** (per clarification). The "docs" badge therefore links to the
  in-repo documentation index.
- **Badge scope**: The requested badge set is CI/build status, latest release/tag, license,
  Go version, Go Report Card (code quality), a docs link, and Go Reference (per clarification —
  all sets selected).
- **Canonical targets**: Repo-scoped badges (CI, release, license) target
  `github.com/glapsfun/rune` (the git remote). Module-scoped badges (Go Reference, Go Report
  Card) use the module path `github.com/rune-task-runner/rune`. An existing repo/module naming
  split in the current docs is treated as intentional and preserved.
- **Starting content**: The existing `docs/` markdown and the 15-entry example library are the
  starting point; this feature **reorganizes them in place** — the existing `docs/guides/` is
  folded into the new `user-guide/` + `how-to/` layout, flat pages (overview, cli, runefile,
  etc.) are kept, and all internal links are updated so the docs form one unified structure
  with the goal-oriented index as its navigation surface. It does not start from scratch, and
  Node, Python, MCP, and agent-driven examples already exist and will be built upon.
- **Badge tooling**: Badge images come from standard third-party shield/badge providers chosen
  at planning time; this is a documentation concern only and does not affect the shipped
  binary or add any runtime dependency.
- **Scope boundary**: This is a docs-and-README change only. Rune's code, CLI, DSL, and golden
  output are untouched; "deep research" applies to documentation quality and badge/provider
  selection, not to tool changes.
- **License**: The project is MIT-licensed (`LICENSE` present), which the license badge
  reflects.
