package tui

import (
	"bytes"
	"testing"

	"github.com/charmbracelet/bubbles/list"
)

func toListItems(items []PickerItem) []list.Item {
	out := make([]list.Item, len(items))
	for i, it := range items {
		out[i] = it
	}
	return out
}

// TestSectionDelegate_HeaderOnlyAtSectionStart asserts headerLine returns the
// group name exactly at each section's first item (by adjacency among the
// list's items) and "" everywhere else, including every item of the
// ungrouped ("") section (FR-001, FR-003).
func TestSectionDelegate_HeaderOnlyAtSectionStart(t *testing.T) {
	items := []PickerItem{
		{Name: "hello"},
		{Name: "compile", Section: "build"},
		{Name: "lint", Section: "build"},
		{Name: "unit", Section: "test"},
	}
	listItems := toListItems(items)
	l := list.New(listItems, list.NewDefaultDelegate(), 80, 24)
	d := newSectionDelegate(newStyles(false), true)

	want := []string{"", "build", "", "test"}
	for i, w := range want {
		if got := d.headerLine(l, i, listItems[i]); got != w {
			t.Errorf("headerLine(%d) = %q, want %q", i, got, w)
		}
	}
}

// TestSectionDelegate_UngroupedNeverGetsAHeader covers the all-ungrouped
// case: no item, regardless of position, ever gets a header line.
func TestSectionDelegate_UngroupedNeverGetsAHeader(t *testing.T) {
	items := []PickerItem{{Name: "a"}, {Name: "b"}, {Name: "c"}}
	listItems := toListItems(items)
	l := list.New(listItems, list.NewDefaultDelegate(), 80, 24)
	d := newSectionDelegate(newStyles(false), true)

	for i := range items {
		if got := d.headerLine(l, i, listItems[i]); got != "" {
			t.Errorf("headerLine(%d) = %q, want empty", i, got)
		}
	}
}

// TestSectionDelegate_NotGroupedMatchesDefaultDelegate proves FR-004/SC-002:
// when grouping is off, the delegate's Height and Render output must be
// byte-identical to the plain list.DefaultDelegate used before this feature.
func TestSectionDelegate_NotGroupedMatchesDefaultDelegate(t *testing.T) {
	items := []PickerItem{{Name: "build", Desc: "compile the binary"}}
	listItems := toListItems(items)
	l := list.New(listItems, list.NewDefaultDelegate(), 80, 24)

	def := list.NewDefaultDelegate()
	sec := newSectionDelegate(newStyles(false), false)

	if sec.Height() != def.Height() {
		t.Fatalf("Height() = %d, want %d (DefaultDelegate's)", sec.Height(), def.Height())
	}

	var wantBuf, gotBuf bytes.Buffer
	def.Render(&wantBuf, l, 0, listItems[0])
	sec.Render(&gotBuf, l, 0, listItems[0])
	if gotBuf.String() != wantBuf.String() {
		t.Fatalf("grouped=false render diverges from DefaultDelegate:\n got:  %q\n want: %q", gotBuf.String(), wantBuf.String())
	}
}

// TestSectionDelegate_GroupedAddsExactlyOneLine proves the grouped delegate's
// Height is the default plus one reserved line, and that Render's output
// starts with that extra line (header or blank) followed by the item's
// normal (unmodified) row.
func TestSectionDelegate_GroupedAddsExactlyOneLine(t *testing.T) {
	items := []PickerItem{{Name: "compile", Desc: "compile the binary", Section: "build"}}
	listItems := toListItems(items)
	l := list.New(listItems, list.NewDefaultDelegate(), 80, 24)

	def := list.NewDefaultDelegate()
	sec := newSectionDelegate(newStyles(false), true)

	if want := def.Height() + 1; sec.Height() != want {
		t.Fatalf("Height() = %d, want %d", sec.Height(), want)
	}

	var wantBuf, gotBuf bytes.Buffer
	def.Render(&wantBuf, l, 0, listItems[0])
	sec.Render(&gotBuf, l, 0, listItems[0])
	if got, want := gotBuf.String(), "build\n"+wantBuf.String(); got != want {
		t.Fatalf("grouped render = %q, want %q", got, want)
	}
}
