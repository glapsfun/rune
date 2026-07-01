package main

import (
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

// colorMode is the validated value of the global --color flag.
type colorMode string

const (
	colorAuto   colorMode = "auto"
	colorAlways colorMode = "always"
	colorNever  colorMode = "never"
)

// parseColorMode validates the --color flag value, returning a usage-style error
// for anything other than auto|always|never (FR-009).
func parseColorMode(s string) (colorMode, error) {
	switch m := colorMode(s); m {
	case colorAuto, colorAlways, colorNever:
		return m, nil
	default:
		return "", fmt.Errorf("invalid --color value %q: want auto, always, or never", s)
	}
}

// resolveColor decides whether to emit ANSI on a single stream. Precedence
// (highest first): --color=never forces off; --color=always forces on (even
// through a pipe); otherwise NO_COLOR disables, and finally the stream's own TTY
// status decides. The decision is per-stream by design, so it deliberately does
// not consult any process-global color flag (which would be derived from a
// single stream and taint the other).
func resolveColor(mode colorMode, isTTY bool) bool {
	switch mode {
	case colorNever:
		return false
	case colorAlways:
		return true
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return isTTY
}

// streamIsTTY reports whether w is a terminal-backed *os.File. A non-file writer
// (pipe, buffer, test capture) is never a TTY, so styling stays off there.
func streamIsTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}
