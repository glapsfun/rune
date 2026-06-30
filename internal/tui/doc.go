// Package tui implements Rune's interactive task picker (the `--choose`
// experience) as a Bubble Tea program. It is a pure presentation layer: it
// renders a filterable list of tasks and reports the user's selection, but it
// never loads a Runefile or executes a task. The caller (internal/cli) projects
// tasks into [PickerItem] values, runs the program, and delegates the selected
// task name to the existing execution path. Keeping execution out of this
// package preserves Rune's locked engine layout and keeps Update a pure,
// table-testable function of (Model, Msg).
package tui
