package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/rune-task-runner/rune/internal/style"
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

// tolerantTheme resolves styling for the two code paths that run when cobra's
// normal flag parsing may not have completed: --help (PersistentPreRunE does
// not fire) and the post-Execute failure banner (a flag-parse error aborts
// before --color is applied). It prefers an explicit --color from the raw
// arguments — so the user's choice survives a usage error — then the parsed
// flag, then auto; an unrecognized value falls back to auto rather than
// erroring, since refusing to print help or a banner over a bad flag is hostile.
func tolerantTheme(cmd *cobra.Command, w io.Writer) style.Theme {
	mode := colorAuto
	if fl := cmd.Flag("color"); fl != nil {
		if m, err := parseColorMode(fl.Value.String()); err == nil {
			mode = m
		}
	}
	if m, ok := rawColorMode(os.Args[1:]); ok {
		mode = m
	}
	return style.New(resolveColor(mode, streamIsTTY(w)), w)
}

// rawColorMode extracts an explicit --color value from the global-flag region
// of args, used as a fallback when pflag may have aborted before binding the
// flag. Because SetInterspersed(false) makes every argument at or after the
// first positional (the task name) belong to the task, the scan stops at the
// first non-flag token as well as at the "--" separator — so a task's own
// --color is never mistaken for Rune's global flag. It honors both
// "--color=VALUE" and "--color VALUE" forms and lets the last explicit
// occurrence win; an unrecognized value is ignored.
func rawColorMode(args []string) (colorMode, bool) {
	var (
		mode  colorMode
		found bool
	)
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" || !strings.HasPrefix(a, "-") {
			break
		}
		var v string
		switch {
		case strings.HasPrefix(a, "--color="):
			v = a[len("--color="):]
		case a == "--color" && i+1 < len(args):
			v = args[i+1]
			i++
		default:
			continue
		}
		if m, err := parseColorMode(v); err == nil {
			mode, found = m, true
		}
	}
	return mode, found
}
