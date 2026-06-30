// Package style is the single source of truth for Rune's non-interactive
// terminal styling. It maps semantic roles (error, warning, task name, …) to
// Lip Gloss styles so every output surface draws from one restrained palette.
//
// The package never decides *whether* to colorize — that decision is made once
// per stream at the command boundary (cmd/rune) and passed in as the enabled
// flag. When disabled, every role is the zero Lip Gloss style, so Render returns
// its input unchanged and output is byte-for-byte identical to plain text.
package style

import (
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Semantic palette, defined once. 256-color codes reuse the picker's accent
// (170) and muted greys (245/241) so the whole CLI reads as one theme.
const (
	colorError   = lipgloss.Color("1")   // red
	colorWarning = lipgloss.Color("3")   // yellow
	colorSuccess = lipgloss.Color("2")   // green
	colorAccent  = lipgloss.Color("170") // task names, headings
	colorMuted   = lipgloss.Color("245") // docs, echo, cache notices
)

// Theme holds one Lip Gloss style per semantic role. The zero Theme is valid and
// renders every role as plain text (no ANSI), which is exactly the disabled
// state produced by New(false, …).
type Theme struct {
	Error    lipgloss.Style // error severity / failures
	Warning  lipgloss.Style // warning severity, cache-write warnings
	Success  lipgloss.Style // active/running accents
	TaskName lipgloss.Style // task identifiers in --list
	Heading  lipgloss.Style // group headers, help section titles
	Muted    lipgloss.Style // docs, command echo, cached/would-run notices
	Locator  lipgloss.Style // file:line:col emphasis in diagnostics
	Caret    lipgloss.Style // diagnostic caret span
}

// New builds the theme for one output stream. When enabled is false it returns
// the zero Theme (all roles plain). When true, styles are bound to an explicit
// Lip Gloss renderer whose color profile is forced on, so color is emitted even
// when the stream is a pipe (honoring --color=always) rather than being silently
// stripped by termenv's TTY auto-detection.
func New(enabled bool, w io.Writer) Theme {
	if !enabled {
		return Theme{}
	}
	r := lipgloss.NewRenderer(w)
	r.SetColorProfile(termenv.ANSI256)

	bold := r.NewStyle().Bold(true)
	return Theme{
		Error:    bold.Foreground(colorError),
		Warning:  bold.Foreground(colorWarning),
		Success:  r.NewStyle().Foreground(colorSuccess),
		TaskName: bold.Foreground(colorAccent),
		Heading:  bold.Foreground(colorAccent),
		Muted:    r.NewStyle().Foreground(colorMuted),
		Locator:  r.NewStyle().Faint(true),
		Caret:    bold.Foreground(colorError),
	}
}
