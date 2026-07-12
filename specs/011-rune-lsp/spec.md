# Feature Specification: Rune Language Server Protocol

**Feature Branch**: `011-rune-lsp`

**Created**: 2026-07-10

**Status**: Draft

**Input**: User description: "Rune Language Server Protocol Feature — provide first-class IDE support for Runefiles through `rune lsp` (real-time diagnostics, autocompletion, go-to-definition, hover documentation, document symbols, formatting) plus a standalone `rune analyze` command, reusing Rune's existing parser, AST, analyzer, import resolver, diagnostics, and formatter. Never create a second implementation of the Runefile language."

## Clarifications

### Session 2026-07-10

- Q: Are the specific RUNE#### diagnostic codes a stable public contract, or is only the detection coverage required? → A: Codes are the contract — each listed condition maps to exactly its RUNE#### code; codes are published to editors, printed by `rune analyze`, and documented, and golden tests assert them exactly.
- Q: When analyzing a root Runefile that imports others, should diagnostics found inside imported files be reported? → A: Yes — report full transitive diagnostics, attributed to their own file:line, in both `rune analyze` and the LSP.
- Q: How should `[private]` tasks appear in completion and go-to-definition? → A: Private tasks are valid dependency completions only within the same file; they are excluded from other files' completion, but go-to-definition and hover always resolve them.
- Q: Should the "public task lacks documentation" warning be part of the first release? → A: Yes — emit it as a warning-level diagnostic that does not affect the exit-code-3 error gate.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Live diagnostics while editing (Priority: P1)

A developer opens a Runefile in an LSP-compatible editor (VS Code, Neovim, Zed, Helix, or similar) and edits it. As they type, the editor continuously highlights problems in place — an unknown dependency, a duplicate task, a dependency cycle, an undefined variable, a wrong argument count, an invalid attribute or setting, an unsupported executor, a syntax or indentation error, an unresolved import, or an incompatible Rune version — each with a precise underline and a human-readable message. Nothing in the Runefile is executed while this happens.

**Why this priority**: This is the headline value of the feature. Catching mistakes at authoring time, before any task runs, is the single reason a developer installs the language server. Every other capability builds on the same analysis. It also delivers the MVP by itself: a developer gets useful feedback the moment they start typing.

**Independent Test**: Open a Runefile with a deliberately broken dependency (`deploy: missing`) in an editor and confirm an error appears on the `missing` token; fix it and confirm the error clears — all without any task or shell command running.

**Acceptance Scenarios**:

1. **Given** a Runefile where task `deploy` depends on `missing` which is not defined, **When** the file is opened, **Then** an error diagnostic is published on the `missing` token stating the dependency is unknown.
2. **Given** the same file, **When** the developer edits `missing` to a valid task name, **Then** the error diagnostic is withdrawn within the debounce window and no stale error remains.
3. **Given** a Runefile containing a dependency cycle across two files (`build → test → generate → build`), **When** analyzed, **Then** a single cycle error is reported with related locations naming every task and file in the cycle.
4. **Given** a Runefile with incomplete syntax (an unterminated string or an open bracket), **When** the developer is mid-edit, **Then** the still-valid declarations around the broken region continue to be analyzed and reported, and the server does not crash or hang.
5. **Given** rapid consecutive edits, **When** several changes arrive within the debounce window, **Then** only diagnostics for the most recent document version are published and results computed for a superseded version are discarded.

---

### User Story 2 - Standalone analysis command (Priority: P1)

A developer (or a CI pipeline) runs `rune analyze` to check a Runefile for problems without executing anything. It prints each diagnostic with `file:line:col`, a severity, a stable diagnostic code, and a message, followed by a summary count, and exits with a code that reflects whether errors were found. `rune analyze --json` emits the same diagnostics as machine-readable output. The analysis is identical to what the editor shows and to what Rune already checks before execution.

**Why this priority**: `rune analyze` is the same analysis service the language server uses, exposed as a command. It is the foundation both technically (a single shared analysis layer) and practically (CI gating, scripting, headless environments where no editor exists). Shipping it first proves the analysis layer in isolation.

**Independent Test**: Run `rune analyze Runefile` against a file with one error and one warning; confirm the printed diagnostics, the summary line, and an exit code of 3; run the same with `--json` and confirm structured output containing the same diagnostics.

**Acceptance Scenarios**:

1. **Given** a valid Runefile, **When** `rune analyze` runs, **Then** it prints no error diagnostics and exits 0, having executed no task.
2. **Given** a Runefile containing at least one error diagnostic, **When** `rune analyze` runs, **Then** it prints each diagnostic as `file:line:col: severity[CODE]: message`, prints a summary count, and exits 3.
3. **Given** any Runefile, **When** `rune analyze --json` runs, **Then** it emits machine-readable diagnostics (code, message, severity, range, related locations) on stdout.
4. **Given** the same Runefile, **When** it is analyzed via `rune analyze` and via the editor, **Then** both report the identical set of diagnostics.
5. **Given** an internal failure unrelated to Runefile content, **When** `rune analyze` runs, **Then** it exits 1 (distinct from the "errors found" exit code 3).

---

### User Story 3 - Context-aware autocompletion (Priority: P2)

While editing, a developer triggers completion and receives suggestions relevant to the cursor's position rather than an undifferentiated list of every symbol. Typing a dependency after a task header suggests task names (including imported, namespaced ones); typing inside an interpolation suggests in-scope variables and parameters; typing after `set` suggests setting names; typing inside `[...]` suggests attributes; typing in the executor position suggests executors; typing a built-in prefix suggests built-in functions. Each suggestion carries a short signature and documentation, and task suggestions also show parameters and their source module.

**Why this priority**: Completion removes the need to memorize task names, setting names, attributes, and built-ins. It is high-value but depends on the analysis layer and symbol index already existing, so it follows the P1 stories.

**Independent Test**: In a Runefile with a `build` task, type `deploy env: bu` and trigger completion; confirm `build` is suggested with its parameters and documentation; select it and confirm the text resolves to `deploy env: build`.

**Acceptance Scenarios**:

1. **Given** a task `build` exists, **When** completion is requested after `deploy: bu`, **Then** `build` (and any other matching task names, including namespaced module tasks like `backend::build`) is suggested with parameters, documentation, and source module.
2. **Given** the cursor is inside a task body interpolation, **When** completion is requested, **Then** only variables and parameters visible in that scope are suggested, and the task's own parameters rank above global variables.
3. **Given** the cursor follows `set `, **When** completion is requested, **Then** valid setting names (e.g. `working-directory`) are suggested.
4. **Given** the cursor is inside an attribute bracket `[`, **When** completion is requested, **Then** valid attributes are suggested (e.g. `confirm`, `private`, `parallel`, `group`, `cache`, platform selectors).
5. **Given** the cursor is in the executor position of a task header, **When** completion is requested, **Then** valid executors (e.g. `sh`, `python`, `node`, `agent`) are suggested.
6. **Given** the cursor follows a built-in prefix in an expression, **When** completion is requested, **Then** matching built-in functions (e.g. `os_family()`) are suggested with signature and documentation.

---

### User Story 4 - Go-to-definition and cross-file navigation (Priority: P2)

A developer holds Ctrl/Cmd and clicks (or invokes "Go to definition") on a symbol and the editor jumps to its declaration: a dependency to its task definition, a variable reference to its assignment, a parameter interpolation to the parameter declaration, a module namespace (`mod backend`) to the module file, and a namespaced module task (e.g. `backend::build`) to its declaration inside the module file. Flat-imported tasks (`import "…"`) carry no prefix and resolve by their bare name to their origin file. Definitions for settings and attributes resolve to their language documentation.

**Why this priority**: Navigation makes multi-file Runefile projects tractable. It relies on the symbol index and import graph from the analysis layer, so it follows diagnostics.

**Independent Test**: In a Runefile where `deploy` depends on `build`, invoke go-to-definition on `build` and confirm the cursor lands on the `build` task declaration; repeat for a `backend::build` module dependency (from `mod backend`) and confirm it opens the module file at the `build` declaration.

**Acceptance Scenarios**:

1. **Given** `deploy: build`, **When** go-to-definition is invoked on `build`, **Then** the editor navigates to the `build` task declaration.
2. **Given** a variable reference in an interpolation, **When** go-to-definition is invoked, **Then** the editor navigates to the variable's assignment.
3. **Given** a parameter used in a task body, **When** go-to-definition is invoked, **Then** the editor navigates to that parameter's declaration in the task header.
4. **Given** `deploy: backend::build` where `backend` is declared with `mod backend`, **When** go-to-definition is invoked on `backend::build`, **Then** the editor opens the module file at the `build` declaration.
5. **Given** an unsaved edit exists in an imported file open in the editor, **When** go-to-definition resolves into that file, **Then** navigation uses the editor's overlay content, not the on-disk version.

---

### User Story 5 - Hover documentation (Priority: P3)

A developer hovers over a symbol and sees a concise documentation panel: for a task, its signature (name and parameters with defaults), its doc comment, its executor and group, and where it is defined; for a parameter, its type, default, and declaring task; for an attribute, what it does; for a built-in, its signature and behavior.

**Why this priority**: Hover is a convenience layered on the same symbol index and shared documentation registry. Valuable but not required for the MVP navigation/diagnostics loop.

**Independent Test**: Hover over a `build target="debug"` task and confirm the panel shows the task signature, its doc comment, executor, group, and definition location.

**Acceptance Scenarios**:

1. **Given** a documented task, **When** hovered, **Then** the panel shows the task signature, documentation, executor, group, and definition location.
2. **Given** a task parameter, **When** hovered, **Then** the panel shows its type, default value, and declaring task.
3. **Given** an attribute in a `[...]` block, **When** hovered, **Then** the panel shows the attribute's description from the shared language registry.
4. **Given** a built-in function, **When** hovered, **Then** the panel shows its signature and documentation from the shared language registry.

---

### User Story 6 - Document outline / symbols (Priority: P3)

A developer opens the editor's outline view and sees the Runefile's structure grouped into settings, variables, imports, and tasks (and modules), so they can navigate the file at a glance.

**Why this priority**: The outline is a straightforward projection of the symbol index. Useful for orientation but not essential to the core editing loop.

**Independent Test**: Open a Runefile with settings, a variable, an import, and several tasks; confirm the outline lists each under its category and clicking an entry navigates to its declaration.

**Acceptance Scenarios**:

1. **Given** a Runefile with settings, variables, imports, and tasks, **When** the outline is requested, **Then** each symbol is listed under its category with a location that navigates to its declaration.
2. **Given** a Runefile with imported modules, **When** the outline is requested, **Then** imports/modules are represented distinctly from local tasks.

---

### User Story 7 - Canonical formatting (Priority: P3)

A developer formats the current document and the editor replaces it with Rune's canonical formatting. Formatting works on the unsaved editor content, and the editor — not the server — decides whether to apply the returned edit. Formatting never shells out to run the CLI.

**Why this priority**: Formatting reuses Rune's existing formatter directly. It is valuable polish but independent of the diagnostics/navigation core.

**Independent Test**: Take an unsaved, poorly formatted Runefile in the editor, invoke format, and confirm the returned edit matches Rune's canonical formatter output for that content.

**Acceptance Scenarios**:

1. **Given** unsaved editor content, **When** formatting is requested, **Then** the server returns a text edit whose result equals Rune's canonical formatter output for that exact content.
2. **Given** formatting is requested, **When** the server produces the edit, **Then** no child process is spawned and no file is written by the server.
3. **Given** content that is already canonically formatted, **When** formatting is requested, **Then** the result is unchanged (idempotent).

---

### User Story 8 - Editor setup out of the box (Priority: P3)

A developer installs a VS Code extension or follows short configuration snippets for Neovim, Zed, or Helix, points the editor at `rune lsp`, and gets the above capabilities with minimal setup.

**Why this priority**: Distribution is what makes the server reachable for real users, but it depends on the server existing and being stable first.

**Independent Test**: Follow the documented setup for one editor and confirm the server starts, initializes, and publishes diagnostics for an opened Runefile.

**Acceptance Scenarios**:

1. **Given** the documented setup for a supported editor, **When** a Runefile is opened, **Then** the language server starts and diagnostics appear.
2. **Given** a supported editor with the server configured, **When** the developer requests completion, definition, hover, symbols, or formatting, **Then** each advertised capability responds.

---

### Edge Cases

- **Arbitrary/garbled input**: For any byte sequence in the document (including binary noise, deeply nested brackets, or truncated declarations), analysis must terminate, must not crash the server, and every reported diagnostic range must point inside the actual document.
- **Unicode and line endings**: Ranges must be correct for documents containing multi-byte characters (e.g. Ukrainian text), emoji, combining characters, mixed CRLF/LF, empty lines, and positions at end-of-file.
- **Unsaved imports**: When an imported Runefile is open with unsaved edits, all analysis and navigation must use the editor's content for that import, not the on-disk file.
- **Imported-file change propagation**: When a file changes, every file that transitively imports it must be re-analyzed and have its diagnostics refreshed.
- **Stale results**: A diagnostic computed for an older document version must never overwrite results for a newer version.
- **Root/workspace ambiguity**: When no explicit workspace folder is provided, the server must still determine a project root deterministically (nearest Runefile, then nearest `.git`, then the document's directory).
- **Protocol hygiene**: Non-protocol output (logs, warnings) must never appear on the stdout channel reserved for protocol messages; malformed incoming protocol messages must not crash the server.
- **Multiple workspace folders**: The first release treats each workspace folder as an independent project; symbols do not leak across unrelated projects.
- **Missing documentation warning**: A public task without documentation is surfaced as a warning-level diagnostic (non-blocking) so it does not affect exit-code-3 error gating on its own. *(See Assumptions.)*

## Requirements *(mandatory)*

### Functional Requirements

#### Analysis engine reuse (foundational)

- **FR-001**: The feature MUST reuse Rune's existing parser, AST, analyzer, import resolver, diagnostics, and formatter. It MUST NOT introduce a second grammar, parser, or independent implementation of the Runefile language.
- **FR-002**: A single shared analysis service MUST produce the diagnostics consumed by CLI execution, `rune analyze`, the language server, and MCP task discovery, so all interfaces report identical results for identical input.
- **FR-003**: The analysis service MUST support an in-memory source overlay so unsaved editor content is analyzed in preference to on-disk content, and this rule MUST apply transitively to imported files.
- **FR-004**: The parser MUST support a recovery mode that continues analyzing valid declarations around invalid or incomplete regions, in addition to the existing strict mode used for execution.
- **FR-005**: For every possible input, the parser MUST terminate, MUST NOT panic, and MUST produce only diagnostic ranges that fall within the document. This invariant MUST be fuzz-tested.
- **FR-006**: Position/range conversion between Rune's source spans and editor line/character positions MUST be centralized (not implemented per handler) and MUST be correct for ASCII, multi-byte characters, emoji, combining characters, CRLF and LF line endings, empty lines, and end-of-file positions.

#### Diagnostics

- **FR-007**: Every diagnostic MUST include a stable code, a message, a severity, a source range, and any related locations.
- **FR-008**: The feature MUST detect and report, at minimum, the parser diagnostics (unexpected token, invalid indentation, unterminated string, incomplete expression, malformed task declaration), the semantic diagnostics (unknown dependency, duplicate task, dependency cycle, undefined variable, wrong argument count, duplicate parameter, invalid attribute, invalid setting, invalid executor), and the project diagnostics (unresolved import, import cycle, duplicate imported namespace, incompatible Rune version), each with a stable code (see FR-010).
- **FR-008a**: The feature MUST additionally emit a warning-level diagnostic when a public (non-private) task has no documentation. This warning MUST NOT, by itself, cause `rune analyze` to exit with the error code (see FR-025); it carries its own stable diagnostic code.
- **FR-009**: Dependency-cycle and import-cycle diagnostics MUST include every task/file involved in the cycle as related locations.
- **FR-009a**: Analysis MUST report diagnostics found in transitively imported files, not only import-level problems seen from the root. Each such diagnostic MUST be attributed to its own file, line, and column. This applies identically to `rune analyze` and to the language server.
- **FR-010**: The specific diagnostic codes enumerated in FR-008 (the RUNE1xxx/RUNE2xxx/RUNE3xxx catalog) are a stable public contract: each listed condition MUST map to exactly its assigned code. The codes MUST be published to editors via the protocol, printed by `rune analyze`, documented, and asserted exactly by golden tests. Once assigned, a code's meaning MUST NOT change.

#### Language server

- **FR-011**: The `rune lsp` command MUST run a language server over stdin/stdout using JSON-RPC and LSP 3.17.
- **FR-012**: The server MUST reserve stdout exclusively for protocol messages; all logs MUST go to stderr or to an explicit log file, controllable via options for log destination and log level.
- **FR-013**: The server MUST publish diagnostics on document open, change, and save, on relevant workspace file changes, and when imported files change.
- **FR-014**: The server MUST advertise only the capabilities it actually implements during initialization.
- **FR-015**: The server MUST support incremental document synchronization, applying edits to its document buffer and analyzing the resulting full document.
- **FR-016**: The server MUST debounce analysis after changes and MUST cancel superseded analysis, so diagnostics for an older document version are never published after a newer version has arrived.
- **FR-017**: The server MUST implement completion, go-to-definition, hover, document symbols, and document formatting as described in User Stories 3–7.
- **FR-018**: Completion results MUST be context-aware (dependency, variable/parameter, setting, attribute, executor, built-in) rather than a flat list of all symbols, and MUST include a signature and documentation per item.
- **FR-019**: Go-to-definition MUST resolve dependencies, variable references, parameter interpolations, module namespaces, and module/imported tasks, including across files. Resolution distinguishes the two cross-file mechanisms: flat `import` tasks resolve by bare name to their declaring file; `mod` tasks resolve by their `name::task` qualified name, and the `name` segment resolves to the `mod` declaration / module file.
- **FR-019a**: `[private]` tasks MUST be offered as dependency completions only within the same file that declares them (where they are callable); they MUST NOT appear in completion for other files. Go-to-definition and hover MUST always resolve private tasks regardless of file.
- **FR-020**: Formatting MUST call Rune's formatter directly (never spawn a child process) and MUST return an edit the editor may choose to apply, computed over the current (possibly unsaved) content.
- **FR-021**: The server MUST determine a workspace root using the order: explicit client workspace folder, then nearest directory containing a Runefile, then nearest directory containing `.git`, then the current document's directory. Each workspace folder MAY be treated as an independent project in the first release.
- **FR-022**: When an imported file changes, the server MUST invalidate and re-analyze all transitive importers and refresh their diagnostics. Files that are not open in the editor are tracked via workspace file-watching so on-disk edits to imported files still refresh dependents.

#### Standalone analysis command

- **FR-023**: The `rune analyze` command MUST analyze the target Runefile (defaulting to `Runefile` when no path is given) together with its transitive imports, and report diagnostics from all of them (per FR-009a), without executing any task or command.
- **FR-024**: `rune analyze` MUST print each diagnostic as `file:line:col: severity[CODE]: message` and a summary count in its default (human) mode, and MUST emit machine-readable diagnostics under `--json`.
- **FR-025**: `rune analyze` MUST exit 0 when no error-severity diagnostics are found and 3 when error-severity diagnostics are present. Discovery/IO failures (no Runefile, unreadable file) exit 2, aligning with Rune's global exit-code scheme (0 success · 2 usage · 3 validation). *(This supersedes the original "1 on internal failure", reconciled during implementation for CLI consistency — see `contracts/cli-analyze.md`.)*

#### Symbol index and language registry

- **FR-026**: The feature MUST build a symbol index over the workspace covering tasks, variables, parameters, settings, attributes, built-ins, imports, and modules, providing at least lookup by name, by qualified name, and by document.
- **FR-027**: There MUST be a single structured registry of Rune language metadata (built-ins, settings, attributes) — with name, kind, signature, and documentation — shared by hover, completion, CLI help/reference generation, documentation generation, and validation. Duplicate hard-coded lists of this metadata MUST NOT exist.

#### Safety

- **FR-028**: No analysis operation (diagnostics, completion, definition, hover, symbols, formatting, `rune analyze`) may execute a task, run a shell command, invoke Python or Node, start an agent, make a network request driven by Runefile logic, expand secret environment values into messages, or write project files.
- **FR-029**: Any future "run task" action MUST be a separate, explicit capability that invokes Rune's standard execution pipeline; it is out of scope for this feature.

### Key Entities *(include if feature involves data)*

- **Analysis Snapshot**: The immutable result of analyzing one document at one version — the parsed file, the source provider, the diagnostics, the symbol index, and the import graph. Consumed by every interface.
- **Open Document**: An editor-held document identified by URI, with a version number and current text; the overlay source that takes precedence over disk.
- **Source Store**: The abstraction that resolves a document's content, choosing the editor overlay when a document is open and falling back to disk otherwise.
- **Workspace / Project**: A root directory, an entry file, its documents, and its import graph; the unit of analysis scope.
- **Import Graph**: The relation of which files import which, in both directions, used to propagate re-analysis when a file changes.
- **Symbol**: A named language entity (task, variable, parameter, setting, attribute, built-in, import, module) with its kind, definition location, selection range, scope, documentation, signature, and export status.
- **Diagnostic**: A code, message, severity, source range, and related locations.
- **Language Registry Entry (Built-in)**: A name, kind, signature, and documentation for a built-in function, setting, or attribute — the single source of language metadata.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer editing a Runefile sees a newly introduced error (e.g. an unknown dependency) reflected in the editor without taking any explicit action beyond typing, and sees it clear once corrected.
- **SC-002**: For every diagnostic type listed in FR-008, a golden test demonstrates that both `rune analyze` and the language server report the same diagnostic (same code, message intent, and range) for the same input.
- **SC-003**: `rune analyze` returns exit code 0 for a clean Runefile, 3 for one containing errors, and 1 on internal failure, and prints a summary count — verified by tests.
- **SC-004**: Across a fuzz corpus of arbitrary byte sequences, the parser and analysis pipeline never panic, always terminate, and never emit a diagnostic range outside the document.
- **SC-005**: Zero task executions, shell invocations, network requests, or project file writes occur during any analysis, completion, definition, hover, symbol, or formatting operation — verified by a test that fails if any such side effect happens.
- **SC-006** (cumulative acceptance — satisfied once US1, US3, US4, and US5 are complete): The full MVP scenario works end to end: in a file with a `build` task, typing `deploy env: bui` and completing suggests `build`; accepting yields `deploy env: build`; go-to-definition on `build` opens its declaration; hover shows its signature and documentation; changing the dependency to `missing` immediately produces an unknown-dependency error — with no task executed at any point.
- **SC-007**: Formatting through the server produces output identical to Rune's canonical formatter for the same content and is idempotent.
- **SC-008**: Position conversion is correct for the full unicode/line-ending matrix (ASCII, Ukrainian text, emoji, combining characters, CRLF, LF, empty lines, EOF), verified by unit tests.
- **SC-009**: A protocol integration test drives a real `rune lsp` subprocess through initialize → initialized → didOpen → completion → definition → hover → shutdown → exit and verifies correct responses and lifecycle behavior.
- **SC-010 (product target, not a release blocker)**: For ordinary Runefiles, open-document analysis completes under 50 ms at P50 and under 150 ms at P95, and completion/definition/hover complete under 50 ms at P95; for a 100-Runefile project, initial workspace indexing completes under 1 second and a single imported-file update completes under 250 ms. These targets are validated by benchmarks once benchmarks exist.
- **SC-011**: At least one editor (VS Code) works via a published client extension, and Neovim, Zed, and Helix each have documented working configurations.

## Assumptions

- **Rune version reference (`0.8.0`)**: The `serverInfo.version` and the `minimum_version` examples in the source description reference `0.8.0`; the actual value MUST track Rune's real release version at ship time (currently past 0.2.0). The number in the description is illustrative, not a requirement to report `0.8.0`.
- **Diagnostic code catalog**: Per the 2026-07-10 clarification, the RUNE1xxx/RUNE2xxx/RUNE3xxx codes are a stable public contract (see FR-010), not merely a coverage guide. The one apparent mismatch in the source (a documentation warning shown under `RUNE2001`, which denotes "unknown dependency") is a typo in the source example; the "public task lacks documentation" warning is a distinct warning-level diagnostic with its own code (see FR-008a) and does not participate in error-code gating.
- **Transport**: Only stdin/stdout (JSON-RPC) transport is in scope. TCP transport is explicitly excluded from the first release.
- **Parser strategy**: Error recovery is achieved by extending Rune's existing hand-written parser (consistent with the constitution's "hand-written front end" principle). Tree-sitter or any second grammar is explicitly excluded.
- **Synchronization**: The server stores the full current document text and performs a full reparse per change; incremental AST parsing is deferred until profiling justifies it. Full reparse is acceptable because Runefiles are small.
- **Debounce interval**: A short debounce (on the order of ~100 ms) is assumed for change-triggered analysis; the exact value is tunable and not itself a hard requirement.
- **Workspace scope**: The first release maintains an independent project per workspace folder; workspace-wide symbol search across folders is out of scope.
- **Excluded from first release**: rename, find references, semantic tokens, code actions, task execution from the editor, inlay hints, workspace-wide symbol search, automatic Rune installation, TCP transport, incremental AST parsing, Tree-sitter grammar, and any third-party plugin system.
- **Testing environment**: Per project policy, the Go test suite (including the LSP protocol, fuzz, and golden tests) runs inside the Docker Compose harness, not on the host.
- **Reuse of existing version gate**: The "incompatible Rune version" diagnostic (RUNE3004) builds on the existing `minimum_version` mechanism rather than introducing a new version-checking scheme.
