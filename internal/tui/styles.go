package tui

import "github.com/charmbracelet/lipgloss"

// Styles holds the Lip Gloss styles for the picker. When color is disabled
// (NO_COLOR or a non-color terminal), every style is the zero (plain) style so
// the picker renders without ANSI escapes while staying fully usable (FR-015).
type Styles struct {
	Title  lipgloss.Style // list title
	Detail lipgloss.Style // detail pane body
	Help   lipgloss.Style // key hint line
	Header lipgloss.Style // section header line, drawn above a section's first item
}

// newStyles builds the style set. With color==false the styles are plain.
func newStyles(color bool) Styles {
	if !color {
		return Styles{}
	}
	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")),
		Detail: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Padding(0, 1),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		// Same accent color --list uses for its "[group]" heading lines
		// (internal/style's colorAccent), so the two surfaces read as one
		// visual language.
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")),
	}
}
