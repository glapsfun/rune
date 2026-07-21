package tui

// PickerItem is one selectable task row, projected from an ast.Task by the
// caller. It implements bubbles/list.Item (FilterValue) and list.DefaultItem
// (Title, Description) so the default delegate can render it.
type PickerItem struct {
	Name    string // task name; what gets executed on selection
	Desc    string // one-line description (first line of the task doc)
	Doc     string // full documentation, shown in the detail pane
	Section string // group name this task belongs to; "" if ungrouped
}

// PickerSection groups a set of tasks under a shared label for New. Name is
// the group("...") attribute's value, or "" for the ungrouped section; it is
// never rendered as a header of its own (see sectionDelegate).
type PickerSection struct {
	Name  string
	Items []PickerItem
}

// Title is the primary label shown in the list (the task name).
func (i PickerItem) Title() string { return i.Name }

// Description is the secondary label shown under the title.
func (i PickerItem) Description() string { return i.Desc }

// FilterValue spans both the name and the description so incremental filtering
// matches either (FR-003).
func (i PickerItem) FilterValue() string { return i.Name + " " + i.Desc }
