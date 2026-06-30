package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	// minHeightForDetail is the terminal height below which the detail pane is
	// hidden so the list stays usable on tiny terminals (FR-017).
	minHeightForDetail = 12
	// detailRows is the number of rows reserved for the detail pane when shown.
	detailRows = 5
)

// Model is the picker state. Update is a pure function of (Model, Msg): it
// changes only the returned Model and emits commands, never touching I/O.
type Model struct {
	list     list.Model
	styles   Styles
	width    int
	height   int
	selected string // chosen task name; "" means cancelled / nothing selected
	done     bool   // true once the user has confirmed or cancelled
}

// New builds a picker over items. color toggles styling (FR-015/FR-021).
func New(items []PickerItem, color bool) Model {
	listItems := make([]list.Item, len(items))
	for i, it := range items {
		listItems[i] = it
	}
	styles := newStyles(color)
	l := list.New(listItems, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Tasks"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	if color {
		l.Styles.Title = styles.Title
	}
	return Model{list: l, styles: styles}
}

// Selected returns the chosen task name, or "" if the user cancelled.
func (m Model) Selected() string { return m.selected }

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.resize()
		return m, nil
	case tea.KeyMsg:
		// While the user is typing a filter, defer every key to the list so
		// characters (including 'q') edit the filter instead of quitting.
		if m.list.SettingFilter() {
			break
		}
		switch msg.String() {
		case "ctrl+c", "q":
			m.selected = ""
			m.done = true
			return m, tea.Quit
		case "enter":
			if it, ok := m.list.SelectedItem().(PickerItem); ok {
				m.selected = it.Name
				m.done = true
				return m, tea.Quit
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m Model) View() string {
	if m.done {
		return ""
	}
	// Collapse the detail pane when the terminal is too short (FR-017).
	if m.height > 0 && m.height < minHeightForDetail {
		return m.list.View()
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.list.View(), m.detailView())
}

// detailView renders the highlighted task's full documentation (FR-004).
func (m Model) detailView() string {
	doc := "(no description)"
	if it, ok := m.list.SelectedItem().(PickerItem); ok && it.Doc != "" {
		doc = it.Doc
	}
	return m.styles.Detail.Render(doc)
}

// resize recomputes the list size, reserving space for the detail pane unless
// the terminal is too short.
func (m *Model) resize() {
	listHeight := m.height
	if m.height >= minHeightForDetail {
		listHeight = m.height - detailRows
	}
	if listHeight < 1 {
		listHeight = 1
	}
	m.list.SetSize(m.width, listHeight)
}
