// Package formatter renders a parsed Runefile back into its canonical source
// form. It is the single formatter shared by the CLI (`rune --fmt`) and the
// language server's textDocument/formatting (spec FR-020); neither shells out.
package formatter
