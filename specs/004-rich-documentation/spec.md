# Feature Specification: Rich, Example-Driven Documentation & Easy-Start Contributing

**Feature Branch**: `004-rich-documentation`

**Created**: 2026-06-09

**Status**: Draft

**Input**: User description: "rich doc about project, use technical-writer during plan, use best practices. I want understandable documentation: how to use, the main idea, and a lot of examples — thinking about different use cases and adding to examples. Update CONTRIBUTING.md with an understandable, easy-to-start guide to what to contribute."

## Overview

Rune already has a working documentation set (README, a `docs/` folder of guides, a
CONTRIBUTING file, and a single getting-started example). This feature raises that set to a
**rich, example-driven body of documentation** that a stranger to the project can read and
immediately understand: what Rune *is*, why it exists, the mental model behind it, and how
to use every capability — each anchored by **many runnable examples drawn from real use
cases**. It also reworks the contributor guide so that someone who has never touched the
project can go from "I'd like to help" to a passing local change with the least possible
friction.

The unit of value here is **reader understanding**, not page count. Every claim in the
documentation is demonstrated with an example a reader can copy, run, and predict. Every
example is verified against the actual tool so the docs never drift from reality.

This is a documentation feature: it changes the project's explanatory artifacts (overview,
guides, an example library, and the contributor guide). It does **not** change Rune's
behavior, language, or CLI.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Grasp the idea and run your first task in minutes (Priority: P1)

A developer who has never heard of Rune arrives at the project. Within a couple of minutes
of reading the top-level overview, they can state in their own words what Rune is, how it
differs from `make`/`just` and from a build system, and when they would (and would not)
reach for it. They then follow a single, linear getting-started path — install, write a
tiny task file, run it — and succeed on the first attempt without hunting across pages or
guessing at missing steps.

**Why this priority**: If a newcomer cannot understand the core idea and get a first win
quickly, none of the deeper documentation matters. This is the irreducible MVP of the
feature: a clear "main idea" plus a friction-free first run. Everything else builds on it.

**Independent Test**: Give the overview and getting-started path to a developer unfamiliar
with Rune; confirm they can (a) correctly describe Rune's purpose and core distinction in
one or two sentences after reading only the overview, and (b) install the tool, author a
small task file, and run a task successfully by following only the getting-started page,
with no external help.

**Acceptance Scenarios**:

1. **Given** the project's landing/overview documentation, **When** a newcomer reads it,
   **Then** it states the problem Rune solves, the core mental model ("named blocks of
   commands, always run when asked, shared by humans and agents"), and explicit "use it
   when / don't use it when" guidance — without requiring any other page.
2. **Given** the getting-started path, **When** a newcomer follows it top to bottom, **Then**
   each step (install → write → run) is complete and self-contained, every command and file
   shown is copy-pasteable, and the expected output is shown alongside it.
3. **Given** a reader on Linux, macOS, or Windows, **When** they follow getting-started,
   **Then** the instructions either work identically or clearly call out any per-platform
   difference at the point it matters.
4. **Given** the first task succeeds, **When** the reader reaches the end of getting-started,
   **Then** the page points them to a clear, named next step (e.g., the example library or
   the language guide) rather than dead-ending.

---

### User Story 2 - Find a runnable example that matches my use case (Priority: P2)

A task author has a concrete job to do — set up tasks for a Go service, a Node or Python
project, a monorepo, a CI pipeline, a Dockerized workflow, a polyglot repo, an AI/agent
workflow, and so on. They browse an **example library organized by use case**, find the
closest match, copy it, and adapt it. Each example is a complete, self-contained, runnable
artifact with a short explanation of what it demonstrates and why it is written that way.
The library is broad enough that most common project shapes have a near-match starting
point.

**Why this priority**: This is the heart of the request — "a lot of examples thinking about
different use cases." Examples are how most people actually learn a task runner, and a rich,
use-case-organized library is the difference between docs that *describe* features and docs
that let someone *get their own job done*. It depends on US1 (the reader must first
understand the basics) but delivers the bulk of the day-to-day value.

**Independent Test**: Pick several distinct use cases from the library, run each example
exactly as published in a clean environment, and confirm it behaves as the surrounding text
says it will; confirm a reader can locate the example relevant to a stated use case (e.g.,
"caching an expensive build step") by browsing the library's organization alone.

**Acceptance Scenarios**:

1. **Given** the example library, **When** a reader browses it, **Then** examples are grouped
   by use case / project shape (not dumped as an undifferentiated list), and each group's
   purpose is labeled so a reader can find a relevant example without reading all of them.
2. **Given** any published example, **When** a reader copies it into a clean project and runs
   the indicated task, **Then** it works as written, and the documentation states the
   expected outcome so the reader can confirm success.
3. **Given** an example that exercises a specific capability (dependencies, parameters,
   caching, parallelism, multi-language bodies, dotenv/settings, composition/imports, OS
   filtering, agent/MCP), **When** the reader reads it, **Then** a short note explains which
   capability it demonstrates and links to the deeper guide for that capability.
4. **Given** an example that requires something not always present (a language interpreter,
   an agent CLI, a container runtime), **When** the reader reads it, **Then** that
   prerequisite is stated up front so the reader is not surprised by a failure.
5. **Given** the set of headline capabilities and common project types, **When** the library
   is reviewed for coverage, **Then** each has at least one worked, runnable example.

---

### User Story 3 - Learn a specific capability in depth when I need it (Priority: P3)

A user who already knows the basics needs to understand one capability properly — how
caching decides to skip, how parameters and variadics work, how parallel prerequisites are
bounded, how the MCP/agent surface and its security model behave, how composition/imports
resolve names, how settings and dotenv loading work. They open the relevant task-oriented
guide, which explains the concept, shows the syntax, gives at least one runnable example,
and calls out the common pitfalls and edge cases — and they come away able to use the
capability correctly.

**Why this priority**: Once people are productive, their questions get specific. Deep,
task-oriented guides turn "it sort of works" into confident, correct usage and cut down
repeated questions. It builds on the overview and examples and so follows them.

**Independent Test**: For each documented capability, confirm its guide contains a concept
explanation, the relevant syntax, at least one runnable example, and an explicit
pitfalls/edge-cases note; verify a reader can answer a realistic "how do I…/what happens
when…" question for that capability using only its guide.

**Acceptance Scenarios**:

1. **Given** any headline capability, **When** a reader opens its guide, **Then** the guide
   explains the underlying idea (not just syntax), shows at least one runnable example, and
   names the common mistakes and edge cases for that capability.
2. **Given** a reader with a "what happens when X goes wrong?" question (e.g., a missing
   interpreter, a cache miss, an undefined variable, a dependency cycle), **When** they
   consult the relevant guide, **Then** the expected behavior and the resulting diagnostic
   are described.
3. **Given** the security-sensitive surface (agents/MCP, secrets, destructive tasks), **When**
   a reader consults its guide, **Then** the safe-by-default behavior and the explicit
   opt-ins required to widen access are clearly stated.
4. **Given** any guide, **When** a reader finishes it, **Then** it cross-links to related
   guides and to at least one example in the library, so navigation never dead-ends.

---

### User Story 4 - Make my first contribution without friction (Priority: P4)

A would-be contributor wants to help but does not know the project. They open the
contributor guide and find, in plain language: what kinds of contributions are welcome
(including low-risk starters like docs and examples), exactly how to set up and verify a
change locally, where things live in the repo, and how to propose the change. They follow
it from a clean clone to a passing local check and an opened change without needing to ask a
maintainer how to begin.

**Why this priority**: Easy onboarding turns interested readers into contributors and is an
explicit ask ("easy to start"). It depends on the rest of the documentation existing (so it
can point at it) and is therefore sequenced last, but it directly serves project health.

**Independent Test**: Have someone unfamiliar with the project follow only the contributor
guide from a clean clone; confirm they can install prerequisites, make a small change
(e.g., fix a doc or add an example), run the project's checks the supported way, and
understand how to submit it — without external guidance.

**Acceptance Scenarios**:

1. **Given** the contributor guide, **When** a newcomer reads it, **Then** it clearly
   describes what to contribute — including explicitly low-barrier first contributions such
   as documentation fixes and new examples — so a beginner knows where to start.
2. **Given** a clean clone, **When** a contributor follows the setup section, **Then** the
   prerequisites, the supported way to build, and the supported way to run the checks are
   stated step by step, in the order a newcomer needs them, with no assumed tribal knowledge.
3. **Given** the project's policy that the test suite runs inside a container (never directly
   on the host), **When** a contributor reads the guide, **Then** that policy and the exact
   supported commands are explained accessibly, including what to do if they don't yet have
   the runner installed.
4. **Given** a first-time contributor, **When** they reach the end of the guide, **Then** they
   know where the main pieces of the project live, how to propose a change, and what the
   review/quality gates will check — so submitting feels predictable rather than risky.
5. **Given** the contributor guide and the user documentation, **When** both are read
   together, **Then** they are consistent with each other and with the project's governing
   principles (no contradictory instructions about commands, policies, or terminology).

---

### Edge Cases

- **Documentation drift**: an example or command shown in the docs no longer matches how the
  tool behaves. The documentation set MUST be verifiable against the real tool so drift is
  caught rather than silently shipped.
- **Reader lands mid-documentation**: someone arrives directly on a deep guide via a search
  engine or deep link, with no prior context. Each page must orient the reader (what it is,
  where it sits, where to go next) rather than assuming they read from the top.
- **Cross-platform divergence**: an instruction or example behaves differently on Windows
  vs. macOS/Linux. Such differences must be called out at the point they matter, not left
  for the reader to discover by failure.
- **Missing prerequisite for an example**: a reader runs an example that needs an
  interpreter, agent CLI, or container runtime they don't have installed. The example must
  state its prerequisites up front.
- **Broken or stale links**: an internal link points at a moved or renamed page. Internal
  navigation must resolve; broken links are a defect.
- **Example that could be destructive or surprising**: an example that demonstrates a gated
  or destructive task must make its nature obvious and must not encourage running something
  harmful unattended.
- **Beginner without the project's tooling**: a contributor or reader who does not yet have
  the tool (or container runtime) installed must still find a path forward, not a dead end.
- **Terminology inconsistency**: the same concept is named two different ways across pages,
  confusing the reader. Terminology must be consistent across the whole set.

## Requirements *(mandatory)*

### Functional Requirements

**Understanding & the main idea**

- **FR-001**: The documentation MUST include an overview that explains, in plain language and
  without requiring any other page, the problem Rune solves, its core mental model, and how
  it differs from a build system and from comparable command runners.
- **FR-002**: The overview MUST give explicit "use Rune when… / don't use Rune when…"
  guidance so a reader can decide whether the tool fits their need.
- **FR-003**: The documentation MUST present a single, linear getting-started path that takes
  a newcomer from nothing to a successfully run task, with every step self-contained and
  every command/file copy-pasteable and accompanied by its expected output.
- **FR-004**: The getting-started path MUST end by directing the reader to a clearly named
  next step rather than dead-ending.

**Examples (the core of this feature)**

- **FR-005**: The documentation MUST include an example library organized by use case /
  project shape, with each group labeled by purpose so a reader can locate a relevant example
  by browsing the organization alone.
- **FR-006**: Every example MUST be self-contained and runnable as published, and MUST state
  the expected outcome so a reader can confirm success.
- **FR-007**: The example library MUST cover, at minimum, one worked example for each headline
  capability (task dependencies, parameters/variadics, content-hash caching, parallel
  prerequisites, multi-language bodies, settings/dotenv, composition/imports, OS filtering,
  and the AI-agent/MCP surface) and for the common project shapes (a compiled-language
  service, a Node/JavaScript project, a Python project, a monorepo, a CI/CD pipeline, a
  containerized workflow, a polyglot repository, and an agent-driven workflow).
- **FR-008**: Each example MUST state any prerequisite it requires (e.g., a language
  interpreter, an agent CLI, or a container runtime) before the reader runs it.
- **FR-009**: Each example MUST note which capability it demonstrates and link to the deeper
  guide for that capability.

**Depth & reference**

- **FR-010**: Each headline capability MUST have a task-oriented guide that explains the
  concept (not only the syntax), shows at least one runnable example, and names the common
  pitfalls and edge cases for that capability.
- **FR-011**: The documentation MUST describe expected behavior for the common failure modes
  (e.g., missing interpreter, cache miss, undefined variable, dependency cycle, arity
  mismatch), including the diagnostic the reader should expect to see.
- **FR-012**: The documentation for the agent/MCP surface MUST clearly state the
  safe-by-default behavior (read-only agent access, env-only secrets, gated destructive
  tasks) and the explicit opt-ins required to widen access.
- **FR-013**: A complete reference for the command-line surface (every command, flag, and
  documented exit code) MUST be present and consistent with the tool's actual behavior.

**Navigation, consistency & accuracy**

- **FR-014**: Every page MUST orient a reader who arrives directly on it (what the page is and
  where to go next), and pages MUST cross-link to related guides and to at least one relevant
  example so navigation does not dead-end.
- **FR-015**: All internal documentation links MUST resolve to existing targets; broken
  internal links MUST be treated as defects.
- **FR-016**: Terminology MUST be used consistently across the entire documentation set (one
  name per concept).
- **FR-017**: The runnable examples and commands shown in the documentation MUST be verifiable
  against the actual tool so that documentation drift is detected rather than shipped, and the
  project MUST provide a repeatable way to perform that verification.
- **FR-018**: Cross-platform differences in any instruction or example MUST be called out at
  the point they matter (Linux, macOS, Windows).

**Contributing (easy to start)**

- **FR-019**: The contributor guide MUST describe what to contribute, explicitly including
  low-barrier first contributions (such as documentation fixes and new examples), so a
  beginner knows where to start.
- **FR-020**: The contributor guide MUST provide a step-by-step path from a clean clone to a
  verified local change: prerequisites, the supported way to build, and the supported way to
  run the project's checks, in the order a newcomer needs them.
- **FR-021**: The contributor guide MUST explain, accessibly, the project's policy that the
  test suite runs inside a container (never directly on the host), including the exact
  supported commands and what to do if the reader does not yet have the tool installed.
- **FR-022**: The contributor guide MUST orient a newcomer to where the major pieces of the
  project live, how to propose a change, and what the review/quality gates will check.
- **FR-023**: The contributor guide MUST be consistent with the user documentation and with
  the project's governing principles (no contradictory commands, policies, or terminology).

**Scope boundary**

- **FR-024**: This feature MUST NOT change Rune's runtime behavior, task-file language, or CLI;
  it changes only the project's documentation, examples, and contributor guide. If the
  documentation work reveals an actual product defect, that defect MUST be recorded
  separately rather than fixed by quietly editing the docs to match wrong behavior.

### Key Entities

- **Overview / landing documentation**: The first thing a newcomer reads; conveys the problem,
  the mental model, the differentiators, and the fit ("when to use / not use").
- **Getting-started path**: The single linear install → write → run journey to a first success.
- **Example**: A self-contained, runnable artifact demonstrating one or more capabilities or a
  project shape, with a stated purpose, prerequisites, and expected outcome.
- **Example library**: The use-case-organized collection of examples, grouped and labeled so
  readers can find a relevant starting point.
- **Capability guide**: A task-oriented deep-dive for one capability — concept, syntax, a
  runnable example, and pitfalls/edge cases.
- **CLI reference**: The complete, accurate catalog of commands, flags, and exit codes.
- **Contributor guide**: The onboarding document covering what to contribute, local setup and
  verification, repository orientation, and the change/review process.
- **Verification mechanism**: The repeatable means by which the documentation's examples and
  commands are checked against the real tool to prevent drift.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer unfamiliar with Rune can correctly state its purpose and its core
  distinction (command runner vs. build system; shared by humans and agents) in their own
  words after reading **only** the overview, in under 3 minutes.
- **SC-002**: A new user can go from a standing start (nothing installed) to a successfully
  run task by following only the getting-started path in under 5 minutes, on the first
  attempt, with no external help.
- **SC-003**: 100% of runnable code/command examples in the documentation execute successfully
  as published, verified by the project's repeatable verification mechanism (zero drifted
  examples at ship time).
- **SC-004**: Every headline capability and every targeted common project shape (per FR-007)
  has at least one worked, runnable example in the library (100% coverage of the defined set).
- **SC-005**: 100% of internal documentation links resolve to existing targets (zero broken
  internal links).
- **SC-006**: A reader can locate the example or guide relevant to a stated common task (e.g.,
  "cache an expensive step," "run prerequisites in parallel," "expose tasks to an agent")
  within two navigation steps from the documentation's entry point.
- **SC-007**: A first-time contributor, following only the contributor guide from a clean
  clone, can make a small change and run the project's checks the supported way in under 15
  minutes, without asking a maintainer how to begin.
- **SC-008**: Each headline capability's guide contains all four required elements (concept,
  syntax, at least one runnable example, and an explicit pitfalls/edge-cases note) — 100% of
  capability guides complete on this measure.
- **SC-009**: The documentation set is internally consistent: zero contradictions between the
  contributor guide, the user guides, and the project's governing principles regarding
  commands, policies, and terminology (measured by review against a consistency checklist).
- **SC-010**: Reader-facing examples that require an unusual prerequisite (interpreter, agent
  CLI, container runtime) state that prerequisite up front in 100% of cases.

## Assumptions

- **The product is stable and not changing**: this work documents and exemplifies Rune as it
  exists; it does not add or alter features. Any product defect surfaced while writing docs is
  logged separately, not masked by the documentation.
- **Audience**: the primary readers are developers comfortable with a terminal who may be new
  to Rune specifically; many will know `make` or `just`. The docs assume terminal literacy but
  not prior Rune knowledge.
- **Example scope**: "a lot of examples across different use cases" is interpreted as at least
  one runnable example per headline capability and per common project shape enumerated in
  FR-007; this is treated as the minimum coverage bar, expandable over time.
- **Examples live in the repository and are verifiable**: examples are real files a reader can
  run, kept honest by the project's existing documentation-verification mechanism (extended as
  needed), rather than illustrative snippets that can silently rot.
- **Format and tooling are an implementation choice for planning**: how the docs are authored,
  laid out, and verified (file layout, any site generator, the exact drift-check command) is
  decided during planning, not fixed by this spec. The existing `docs/` set, README, and
  CONTRIBUTING are the starting point and will be extended/reworked rather than replaced
  wholesale.
- **Cross-platform parity is a documentation obligation**: because the tool targets Linux,
  macOS, and Windows, examples and instructions account for all three, calling out any
  divergence explicitly.
- **The contributor guide reflects the real workflow**: it documents the project's actual,
  governing policies (container-only testing, the existing CI quality gates, the
  dogfooded development tasks) accurately and accessibly.
- **Planning will apply professional technical-writing practice**: per the request, the
  planning phase will engage technical-writing best practices (clear information architecture,
  task-oriented structure, consistent terminology, scannable formatting, and verified
  examples). The `technical-writer` capability is expected to be used during `/speckit-plan`.
