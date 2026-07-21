package tui

import (
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func testItems() []PickerItem {
	return []PickerItem{
		{Name: "build", Desc: "compile the binary", Doc: "Compile the static binary."},
		{Name: "test", Desc: "run the suite", Doc: "Run the test suite in Docker."},
		{Name: "lint", Desc: "static checks", Doc: "Run golangci-lint."},
	}
}

// singleSection wraps items in one unnamed section - the shape a Runefile
// with no group(...) attributes projects to.
func singleSection(items []PickerItem) []PickerSection {
	return []PickerSection{{Items: items}}
}

// groupedSections returns tasks split across two named groups, mirroring a
// Runefile with group("build") and group("test") attributes.
func groupedSections() []PickerSection {
	return []PickerSection{
		{Name: "build", Items: []PickerItem{
			{Name: "compile", Desc: "compile the binary"},
			{Name: "lint", Desc: "static checks"},
		}},
		{Name: "test", Items: []PickerItem{
			{Name: "unit", Desc: "run the suite"},
		}},
	}
}

// mixedSections mirrors spec.md US1 Acceptance Scenario 1: named groups
// alongside a section of ungrouped tasks.
func mixedSections() []PickerSection {
	return []PickerSection{
		{Items: []PickerItem{{Name: "hello", Desc: "say hi"}}},
		{Name: "build", Items: []PickerItem{{Name: "compile", Desc: "compile the binary"}}},
		{Name: "test", Items: []PickerItem{{Name: "unit", Desc: "run the suite"}}},
	}
}

// newSizedModel returns a picker that has received a window size, so the list
// is laid out and navigable.
func newSizedModel(t *testing.T) Model {
	t.Helper()
	m := New(singleSection(testItems()), false)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return next.(Model)
}

// newSizedGroupedModel is newSizedModel's grouped counterpart.
func newSizedGroupedModel(t *testing.T) Model {
	t.Helper()
	m := New(groupedSections(), false)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return next.(Model)
}

func isQuit(t *testing.T, cmd tea.Cmd) bool {
	t.Helper()
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

func TestNavigationMovesHighlight(t *testing.T) {
	m := newSizedModel(t)
	if got := m.list.Index(); got != 0 {
		t.Fatalf("initial index = %d, want 0", got)
	}
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(Model)
	if got := m.list.Index(); got != 1 {
		t.Fatalf("after down, index = %d, want 1", got)
	}
	// 'j' is bound to "down" in the default keymap.
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = next.(Model)
	if got := m.list.Index(); got != 2 {
		t.Fatalf("after j, index = %d, want 2", got)
	}
}

func TestEnterSelectsHighlightedAndQuits(t *testing.T) {
	m := newSizedModel(t)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if m.Selected() != "build" {
		t.Fatalf("selected = %q, want %q", m.Selected(), "build")
	}
	if !isQuit(t, cmd) {
		t.Fatalf("Enter did not return tea.Quit")
	}
}

func TestEnterAfterNavigationSelectsCorrectTask(t *testing.T) {
	m := newSizedModel(t)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(Model)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(Model)
	if m.Selected() != "test" {
		t.Fatalf("selected = %q, want %q", m.Selected(), "test")
	}
	if !isQuit(t, cmd) {
		t.Fatalf("Enter did not return tea.Quit")
	}
}

func TestCancelKeysLeaveNoSelection(t *testing.T) {
	cases := []struct {
		name string
		msg  tea.KeyMsg
	}{
		{"q", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}},
		{"ctrl+c", tea.KeyMsg{Type: tea.KeyCtrlC}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newSizedModel(t)
			next, cmd := m.Update(tc.msg)
			m = next.(Model)
			if m.Selected() != "" {
				t.Fatalf("selected = %q, want empty after cancel", m.Selected())
			}
			if !isQuit(t, cmd) {
				t.Fatalf("%s did not return tea.Quit", tc.name)
			}
		})
	}
}

func TestWindowSizeRendersAndCollapsesDetailWhenShort(t *testing.T) {
	m := New(singleSection(testItems()), false)
	// Tall terminal: detail pane shown alongside the list.
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	tall := next.(Model)
	if tall.View() == "" {
		t.Fatal("tall view is empty")
	}
	// Short terminal: must still render without panic (detail collapsed).
	next, _ = tall.Update(tea.WindowSizeMsg{Width: 80, Height: 6})
	short := next.(Model)
	if short.View() == "" {
		t.Fatal("short view is empty")
	}
}

func TestFilterValueSpansNameAndDescription(t *testing.T) {
	it := PickerItem{Name: "build", Desc: "compile the binary"}
	if got, want := it.FilterValue(), "build compile the binary"; got != want {
		t.Fatalf("FilterValue() = %q, want %q", got, want)
	}
}

// TestNewFlattensSectionsTaggingEachItem asserts New's flattening step
// preserves section order and tags every item with its section name,
// including the ungrouped ("") section (FR-001, FR-002, FR-003).
func TestNewFlattensSectionsTaggingEachItem(t *testing.T) {
	m := New(mixedSections(), false)
	var got []string
	for _, it := range m.list.Items() {
		pi := it.(PickerItem)
		got = append(got, pi.Name+"/"+pi.Section)
	}
	want := []string{"hello/", "compile/build", "unit/test"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("flattened items = %v, want %v", got, want)
	}
}

// TestFilterNarrowsToMatchingSectionOnly proves FR-006: once a filter leaves
// only one section with surviving tasks, no other section's tasks (and so no
// other section's header, which is derived from adjacency among visible
// items) remain visible.
func TestFilterNarrowsToMatchingSectionOnly(t *testing.T) {
	m := newSizedGroupedModel(t)
	m.list.SetFilterText("unit")

	visible := m.list.VisibleItems()
	if len(visible) != 1 {
		t.Fatalf("visible items = %d, want 1 (%v)", len(visible), visible)
	}
	pi, ok := visible[0].(PickerItem)
	if !ok || pi.Name != "unit" || pi.Section != "test" {
		t.Fatalf("visible item = %+v, want unit/test", visible[0])
	}
}

// TestSelectionAlwaysAPickerItem is a regression guard for FR-005: since
// headers are never list.Items (see delegate.go), every key that moves the
// cursor - including jump keys - can only ever land on a PickerItem. There is
// nothing to "skip" by construction, but this still guards against a future
// change reintroducing a non-task item into the list.
func TestSelectionAlwaysAPickerItem(t *testing.T) {
	m := newSizedGroupedModel(t)
	keys := []tea.KeyMsg{
		{Type: tea.KeyDown},
		{Type: tea.KeyDown},
		{Type: tea.KeyDown},
		{Type: tea.KeyUp},
		{Type: tea.KeyHome},
		{Type: tea.KeyEnd},
		{Type: tea.KeyPgDown},
		{Type: tea.KeyPgUp},
	}
	for _, k := range keys {
		next, _ := m.Update(k)
		m = next.(Model)
		if _, ok := m.list.SelectedItem().(PickerItem); !ok {
			t.Fatalf("after %v, SelectedItem() = %#v, want a PickerItem", k, m.list.SelectedItem())
		}
	}
}
