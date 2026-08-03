// Package styletest provides shared test helpers for asserting terminal
// styling. It is imported only by test files, so it never ships in the binary
// (unlike a helper exported from the production style package), while keeping a
// single definition of the escape grammar for every styled/plain parity test.
package styletest

import "regexp"

// sgrPattern matches SGR (Select Graphic Rendition) escape sequences — the only
// escapes Rune's themed output emits.
var sgrPattern = regexp.MustCompile("\x1b\\[[0-9;]*m")

// StripSGR removes SGR escape sequences, recovering the visible text. Because
// the theme's emphasis is zero-width, stripping a styled string must reproduce
// its plain form byte-for-byte — the invariant every styled/plain parity test
// asserts.
func StripSGR(s string) string { return sgrPattern.ReplaceAllString(s, "") }
