package tui

import (
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

// newSizedModel returns a picker that has received a window size, so the list
// is laid out and navigable.
func newSizedModel(t *testing.T) Model {
	t.Helper()
	m := New(testItems(), false)
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
	m := New(testItems(), false)
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
