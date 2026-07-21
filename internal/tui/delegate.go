package tui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
)

// sectionDelegate wraps list.DefaultDelegate to draw a section's group name
// as a decorative line above whichever visible item starts that section.
//
// Headers are deliberately never their own list.Item: bubbles/list has no
// per-item variable height, and giving a header its own Item would fight the
// library's own fuzzy filter (SetItems re-triggers async filtering, which
// would immediately strip a header back out, since its FilterValue can never
// legitimately match arbitrary search text). Deriving the header purely from
// adjacency within m.VisibleItems() at render time means the cursor only
// ever lands on a real task — every navigation key, including
// Home/End/PageUp/PageDown, is correct with no special-casing — and a
// section's header disappears on its own once filtering leaves it with zero
// surviving tasks.
type sectionDelegate struct {
	list.DefaultDelegate
	styles  Styles
	grouped bool // true when the Runefile defines at least one group
}

// newSectionDelegate builds the delegate. When grouped is false the height
// is left at the embedded DefaultDelegate's default, so the no-groups picker
// renders byte-identical to before this feature (FR-004).
func newSectionDelegate(styles Styles, grouped bool) sectionDelegate {
	d := list.NewDefaultDelegate()
	if grouped {
		d.SetHeight(d.Height() + 1)
	}
	return sectionDelegate{DefaultDelegate: d, styles: styles, grouped: grouped}
}

// Render writes the optional header line, then delegates the item's own row
// to the embedded DefaultDelegate unchanged.
func (d sectionDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if d.grouped {
		fmt.Fprintln(w, d.headerLine(m, index, item))
	}
	d.DefaultDelegate.Render(w, m, index, item)
}

// headerLine returns the styled group name when item is the first visible
// item of its section, or "" (a blank line) otherwise — including when the
// item's section is the ungrouped ("") section, which never gets a header of
// its own (matching --list's treatment of ungrouped tasks).
func (d sectionDelegate) headerLine(m list.Model, index int, item list.Item) string {
	pi, ok := item.(PickerItem)
	if !ok || pi.Section == "" {
		return ""
	}
	if index == 0 {
		return d.styles.Header.Render(pi.Section)
	}
	prev, ok := m.VisibleItems()[index-1].(PickerItem)
	if !ok || prev.Section != pi.Section {
		return d.styles.Header.Render(pi.Section)
	}
	return ""
}
