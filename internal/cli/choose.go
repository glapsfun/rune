package cli

import (
	"context"
	"errors"
	"os"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"

	"github.com/rune-task-runner/rune/internal/tui"
)

// chooseAndRun presents the interactive task picker (--choose), then runs the
// selected task. Order matters: the Runefile is loaded and analyzed first, so
// static errors are reported with zero side effects (Principle II) before any
// UI; a non-interactive terminal or an empty task set fails with a clear usage
// error rather than a broken UI. On selection the picker tears itself down and
// the task runs through the same execution path as a direct `rune <task>`.
func chooseAndRun(opts Options, runefile string, args []string) error {
	// --choose is an interactive CLI path, so it honors the CLI --ignore-version
	// flag (execute() applies the same flag on the run that follows selection).
	mod, err := loadModule(opts, runefile, opts.IgnoreVersion)
	if err != nil {
		return err
	}

	if !interactiveTerminal(opts) {
		return usagef("--choose requires an interactive terminal")
	}

	items := pickerItems(mod)
	if len(items) == 0 {
		return usagef("no tasks to choose from")
	}

	picked, err := runPicker(opts, items)
	if err != nil {
		return err
	}
	if picked == "" {
		return nil // nothing selected
	}
	return execute(opts, runefile, append([]string{picked}, args...))
}

// pickerItems projects the loaded module's tasks into selectable rows, applying
// the same visibility rules as `--list` and shell completion: non-private and
// matching the current OS. Order follows the Runefile.
func pickerItems(mod *loadedModule) []tui.PickerItem {
	var items []tui.PickerItem
	for _, t := range mod.file.Tasks {
		if t.IsPrivate() || !osMatches(t, runtime.GOOS) {
			continue
		}
		items = append(items, tui.PickerItem{
			Name: t.Name,
			Desc: firstLine(t.Doc),
			Doc:  t.Doc,
		})
	}
	return items
}

// interactiveTerminal reports whether both stdin and stdout are connected to an
// interactive terminal. The picker requires a real TTY; in any piped,
// redirected, or CI context this returns false so the caller errors instead of
// rendering control sequences into captured output (FR-011).
func interactiveTerminal(opts Options) bool {
	in, okIn := opts.Stdin.(*os.File)
	out, okOut := opts.Stdout.(*os.File)
	if !okIn || !okOut {
		return false
	}
	return isTTY(in.Fd()) && isTTY(out.Fd())
}

func isTTY(fd uintptr) bool {
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

// runPicker runs the Bubble Tea program to completion and returns the selected
// task name ("" if cancelled). The program renders to stderr (where Rune's own
// output goes), keeping stdout clean for the task that runs afterward. It is
// run with the invocation context so an external SIGINT cancels it cleanly.
func runPicker(opts Options, items []tui.PickerItem) (string, error) {
	prog := tea.NewProgram(
		tui.New(items, opts.ColorStderr),
		tea.WithContext(opts.ctx()),
		tea.WithAltScreen(),
		tea.WithInput(opts.Stdin),
		tea.WithOutput(opts.Stderr),
	)
	final, err := prog.Run()
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, tea.ErrProgramKilled) {
			return "", &Interrupted{}
		}
		return "", &UsageError{Err: err}
	}
	if m, ok := final.(tui.Model); ok {
		return m.Selected(), nil
	}
	return "", nil
}
