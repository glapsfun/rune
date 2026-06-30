package style

import (
	"io"
	"regexp"
	"testing"
)

// ansiSeq matches SGR escape sequences so tests can assert presence/absence of
// color and recover the visible text.
var ansiSeq = regexp.MustCompile("\x1b\\[[0-9;]*m")

func stripANSI(s string) string { return ansiSeq.ReplaceAllString(s, "") }

// roles returns each role's Render func by name for table-driven assertions.
func roles(th Theme) map[string]func(...string) string {
	return map[string]func(...string) string{
		"Error":    th.Error.Render,
		"Warning":  th.Warning.Render,
		"Success":  th.Success.Render,
		"TaskName": th.TaskName.Render,
		"Heading":  th.Heading.Render,
		"Muted":    th.Muted.Render,
		"Locator":  th.Locator.Render,
		"Caret":    th.Caret.Render,
	}
}

// FR-003: a disabled theme must emit zero ANSI and return its input byte-for-byte.
func TestDisabledThemeIsPlain(t *testing.T) {
	th := New(false, io.Discard)
	for name, render := range roles(th) {
		got := render("hello")
		if got != "hello" {
			t.Errorf("role %s disabled: got %q, want plain %q", name, got, "hello")
		}
	}
}

// FR-012: an enabled theme adds SGR escapes but never changes the visible text
// or its width (stripping ANSI must recover the original string exactly).
func TestEnabledThemeStylesButPreservesText(t *testing.T) {
	th := New(true, io.Discard)
	const in = "build"
	for name, render := range roles(th) {
		got := render(in)
		if got == in {
			t.Errorf("role %s enabled: expected ANSI styling, got plain %q", name, got)
		}
		if stripANSI(got) != in {
			t.Errorf("role %s enabled: visible text changed: stripped %q, want %q", name, stripANSI(got), in)
		}
	}
}
