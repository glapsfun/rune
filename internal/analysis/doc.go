// Package analysis is the reusable analysis service that wraps Rune's parse →
// compose → analyze pipeline behind a source-overlay abstraction. It produces
// immutable Snapshots consumed by the CLI, `rune analyze`, the language server,
// and MCP task discovery, so every interface reports identical diagnostics
// (spec FR-002). Nothing in this package executes tasks, spawns processes, opens
// sockets, or writes project files (FR-028): it imports neither internal/runtime
// nor os/exec nor any network package.
package analysis
