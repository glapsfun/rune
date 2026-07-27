# Feature Specification: Secret Masking & Sanitization

**Feature Branch**: `013-secret-masking`

**Created**: 2026-07-21

**Status**: Draft

**Input**: User description: "Secret Masking & Sanitization: Implement task's logic for masking secret variables in logs. This ensures that even if an agent runs a task that outputs environment variables, sensitive credentials never end up in the agent's chat history or long-term memory"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Agent runs a task that leaks credentials, output arrives masked (Priority: P1)

An AI agent, connected to Rune as a tool server, invokes a task whose command
prints environment variables — deliberately (`env`, a debug dump) or
accidentally (a tool that echoes its configuration on failure). The task output
returned to the agent has every known secret value replaced with a mask, so no
credential ever enters the agent's chat history, context window, or long-term
memory.

**Why this priority**: This is the core promise of the feature and the reason it
exists. Agent transcripts are frequently logged, cached, and used for training
or memory; a single leaked credential there is unrecoverable. Constitution
Principle VII ("AI-Native, Secure by Default") already keeps secrets out of task
descriptions and listings — task *output* is the remaining gap.

**Independent Test**: Can be fully tested by connecting an agent (or a test
harness acting as one) to the tool server, running a task that prints a
known-secret environment variable, and asserting the returned output contains
the mask but never the secret value.

**Acceptance Scenarios**:

1. **Given** a task environment containing a variable whose name marks it as
   sensitive (e.g., `API_TOKEN`), **When** an agent runs a task that prints that
   variable's value, **Then** the output the agent receives shows a mask in
   place of the value and the raw value appears nowhere in the response.
2. **Given** a task that fails and prints a connection string containing a
   secret value to its error stream, **When** an agent runs it, **Then** both
   the normal and error output returned to the agent are masked.
3. **Given** a task whose output includes the same secret value multiple times
   and across multiple lines, **When** an agent runs it, **Then** every
   occurrence is masked, including occurrences that span buffered chunks of
   streamed output.

---

### User Story 2 - Human terminal runs are masked the same way (Priority: P2)

A developer runs the same task in their terminal. Secret values are masked in
everything Rune emits — task output passthrough, echoed command lines where a
secret was interpolated, error reports, and status/log lines — so terminal
scrollback, CI logs, and screen shares don't expose credentials either.

**Why this priority**: Consistency and defense in depth. Terminal output is
routinely captured (CI systems, `script`, pasted bug reports — and agents also
run Rune through plain shell commands, not only through the tool server). If
masking differed between surfaces, the safety guarantee would depend on *how*
the task was invoked, which authors cannot control.

**Independent Test**: Run a secret-printing task from a terminal and assert the
process's standard output and error streams contain the mask and not the value.

**Acceptance Scenarios**:

1. **Given** a task that prints a sensitive variable, **When** it is run from
   the command line, **Then** the value is masked in the emitted output.
2. **Given** a task whose command line interpolates a secret and command echoing
   is on, **When** it runs, **Then** the echoed command line shows the mask.
3. **Given** a Runefile with no sensitive variables anywhere in its task
   environments, **When** any task runs, **Then** output is byte-identical to
   today's behavior (no masking artifacts, no reformatting, no added latency a
   user would notice).

---

### User Story 3 - Author declares additional secrets beyond the defaults (Priority: P3)

A Runefile author has a credential in a variable whose name does not look
sensitive (e.g., `DEPLOY_CFG`). They declare it as secret in the Runefile, and
from then on its value is masked across all surfaces exactly like the built-in
detections. Conversely, an author with a false positive (a non-secret variable
that matches the sensitive-name patterns) can exempt it.

**Why this priority**: The built-in name patterns cannot know every project's
conventions. Author control makes the guarantee complete, but the feature
already delivers most of its value without it.

**Independent Test**: Declare a neutrally-named variable as secret, run a task
that prints it, and assert it is masked; declare a pattern-matching variable as
non-secret, and assert it is not masked.

**Acceptance Scenarios**:

1. **Given** a variable declared secret by the author, **When** any task prints
   its value, **Then** the value is masked on every output surface.
2. **Given** a variable explicitly exempted by the author, **When** a task
   prints its value, **Then** it appears unmasked.
3. **Given** a secret declaration naming a variable that never exists at run
   time, **When** tasks run, **Then** nothing breaks and no mask appears.

---

### Edge Cases

- **Secret values spanning chunk boundaries**: streamed output is delivered in
  buffers; a value split across two chunks must still be masked, never
  half-revealed.
- **Very short values**: masking a 1–3 character value (e.g., `PORT=1`) would
  mangle unrelated output wherever that character sequence appears; values below
  a documented minimum length are not value-masked, and this limitation is
  documented.
- **Empty or unset secrets**: a declared secret with an empty value must not
  cause every position in the output to be masked or any error.
- **Multi-line secret values** (e.g., PEM keys): each line of the value is
  masked wherever it appears.
- **One secret value contained inside another**: overlapping matches must not
  reveal any part of either value.
- **Secret value equals common output text**: if a secret's value happens to be
  a common word, that word is masked everywhere — accepted cost, favoring
  safety; authors can rename/exempt.
- **Transformed values** (base64, URL-encoded, JSON-escaped occurrences of a
  secret): exact-occurrence masking cannot catch transformations; the
  guarantee covers verbatim occurrences only, and the documentation states
  this boundary clearly.
- **Interrupted tasks**: output already emitted before a cancellation/timeout
  must have been masked at emission time — masking cannot be a post-processing
  step that a crash can skip.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST build, per task run, a set of secret values to
  mask, composed of: (a) values of environment variables in the task's
  environment whose names match a built-in, documented list of sensitive-name
  patterns — names containing `TOKEN`, `SECRET`, `PASSWORD`, `PASSWD`,
  `APIKEY`, `API_KEY`, `PRIVATE_KEY`, `ACCESS_KEY`, `CREDENTIAL`, or `AUTH`
  (the normative list lives in the feature contract) — and
  (b) values of variables the Runefile author explicitly declares as secret.
- **FR-002**: The system MUST replace every verbatim occurrence of each secret
  value with a fixed mask placeholder in all output it emits or returns:
  task standard output and standard error, echoed command lines, Rune's own
  status/log/error messages, and tool-server responses returned to agents.
- **FR-003**: Masking MUST be applied before output leaves the system on any
  surface — an agent or terminal consumer MUST never be able to observe a
  window of unmasked output, including when a task is interrupted mid-stream.
- **FR-004**: Masking MUST handle secret values that are split across streaming
  buffer boundaries and values that span multiple lines.
- **FR-005**: Runefile authors MUST be able to declare additional variables as
  secret, and to exempt specific variables from the built-in name patterns; the
  declaration mechanism MUST live in the Runefile so the behavior is versioned
  with the project.
- **FR-006**: Masking MUST be on by default in every execution mode, with no
  global off switch exposed through the agent-facing surface; any opt-out is an
  explicit per-variable exemption by the Runefile author (FR-005).
- **FR-007**: Secret values shorter than a documented minimum length MUST be
  excluded from value-based masking to avoid corrupting unrelated output; the
  minimum MUST be documented user-facing.
- **FR-008**: A Runefile whose task environments contain no secrets (by pattern
  or declaration) MUST produce output identical to current behavior.
- **FR-009**: Secret declarations and exemptions MUST be validated statically
  like the rest of the language: a malformed declaration fails before execution
  with a positioned diagnostic (file, line, column), per Constitution
  Principle II.
- **FR-010**: The built-in sensitive-name patterns, the mask placeholder, the
  minimum-length rule, and the boundaries of the guarantee (verbatim
  occurrences; transformed values not covered) MUST be documented user-facing.
- **FR-011**: Masking MUST apply to what the system *emits*, never to what the
  task *receives*: the task's own environment and inter-process plumbing keep
  real values, so task behavior is unchanged.

### Key Entities

- **Secret value set**: the per-run collection of values to mask; derived from
  the task's effective environment via name patterns plus author declarations,
  minus exemptions.
- **Secret declaration / exemption**: an author-facing statement in the Runefile
  marking a variable name as secret or as exempt from pattern matching.
- **Output surface**: any channel on which Rune emits text — terminal
  passthrough of task output, echoed commands, Rune's own messages, and
  tool-server responses. All surfaces share one masking behavior.
- **Mask placeholder**: the fixed replacement text shown instead of a secret
  value; identical on every surface.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In a test suite that prints known secrets on every output surface
  (task output, error output, echoed commands, agent tool responses, failure
  diagnostics), 100% of secret occurrences — including chunk-spanning and
  multi-line cases — are masked; zero raw values reach any consumer.
- **SC-002**: An agent connected to the tool server that runs an
  environment-dumping task receives a transcript containing zero credential
  values, verified end-to-end against the real tool-server responses.
- **SC-003**: For Runefiles with no secrets present, output is byte-identical to
  the previous release across the existing golden-file corpus.
- **SC-004**: A task producing 10 MB of output completes within 10% of its
  unmasked wall-clock time.
- **SC-005**: Authors can declare or exempt a secret with a single line in the
  Runefile, and the change takes effect on the next run with no other setup.

## Assumptions

- **Secret identification is name-based plus declared, not content-scanned.**
  Rune identifies secrets from the *names* of variables in the task's effective
  environment (built-in sensitive patterns) plus explicit author declarations.
  It does not scan output for credential-shaped strings (entropy/regex content
  detection) — that approach is heuristic, false-positive-prone, and can never
  be a guarantee. This mirrors established CI systems, which mask registered
  secret values.
- **The whole effective environment is inspected, not just `.env`.** Credentials
  frequently arrive via the host environment (session-injected keys), so name
  patterns apply to every variable the task will see, wherever it came from.
  Conversely, `.env` values with non-sensitive names (e.g., `GREETING`) are not
  masked automatically — treating all configuration as secret would mangle
  ordinary output.
- **Verbatim-occurrence guarantee.** Masking covers exact verbatim occurrences
  of the value; transformed or encoded output (base64 of a secret, URL-encoded
  or JSON-escaped forms, a secret split by the task itself) is out of scope for
  v1 and documented as such.
- **One mask placeholder, no per-variable labeling.** Showing which variable was
  masked (e.g., naming it in the placeholder) is a nice-to-have deferred until
  requested; a fixed placeholder is simpler and leaks nothing.
- **No new secret storage.** This feature masks output only; secrets continue to
  come exclusively from the environment/`.env` per Constitution Principle VII.
  Rune does not become a secret manager.
- **Cache behavior is unchanged.** Task output caching (if any output is stored)
  stores what was emitted — i.e., masked text — so replayed output is safe by
  construction.
