// Package docs holds the documentation verification harness.
//
// It treats the project's documentation as tested fixtures (Constitution
// Principle VI): every example Runefile statically validates, fenced rune code
// blocks on self-contained pages validate, example directories satisfy the
// example contract, internal links resolve, and no secret literals leak. The
// harness builds the rune binary once (see TestMain) and runs it, mirroring
// test/integration. Tests run inside Docker per project policy (CONTRIBUTING.md).
//
// There is no built-in `rune check` subcommand; static validation uses
// `rune --file <path> --list`, which forces parse+analyze and runs nothing.
package docs
