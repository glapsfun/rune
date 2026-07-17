// Package lsp is the protocol glue for `rune lsp`: a JSON-RPC 2.0 server over
// stdio speaking a minimal typed subset of LSP 3.17. It contains no language
// logic — it converts between LSP payloads and the analysis/language layers,
// converts byte spans to UTF-16 positions (convert.go), and enforces that stdout
// carries protocol bytes only (spec FR-011, FR-012, FR-014). It runs nothing
// (FR-028).
package lsp
