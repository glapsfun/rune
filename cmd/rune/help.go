package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rune-task-runner/rune/internal/style"
)

// applyHelp installs Rune's friendly, grouped help on the root command. Section
// headings are colorized when stdout is a color terminal (resolved here because
// Cobra does not run PersistentPreRunE for --help); the body stays plain so
// piped help is ANSI-free and informative (FR-019..FR-021). Subcommands keep
// Cobra's default help (which already carries their flags and examples).
func applyHelp(root *cobra.Command) {
	defaultHelp := root.HelpFunc()
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if cmd.HasParent() {
			defaultHelp(cmd, args)
			return
		}
		out := cmd.OutOrStdout()
		// Resolve color for help output. An invalid --color value is tolerated
		// here (falls back to auto) rather than erroring: PersistentPreRunE — the
		// normal FR-009 validation path — does not run for --help, and refusing
		// to print help on a bad flag would be hostile.
		mode := colorAuto
		if v, err := cmd.Flags().GetString("color"); err == nil {
			if m, perr := parseColorMode(v); perr == nil {
				mode = m
			}
		}
		th := style.New(resolveColor(mode, streamIsTTY(out)), out)
		fmt.Fprint(out, rootHelp(cmd, th))
	})
}

// rootHelp renders the grouped root help. The plain form (disabled theme) is the
// reviewed baseline for this feature.
func rootHelp(cmd *cobra.Command, th style.Theme) string {
	h := th.Heading.Render
	var b strings.Builder

	fmt.Fprintln(&b, "Rune — a shared task runner for humans and AI agents.")
	fmt.Fprintln(&b, "It runs tasks from a Runefile on the CLI, and exposes them to AI agents over MCP.")
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, h("Usage:"))
	fmt.Fprintln(&b, "  rune [global flags] [VAR=VALUE ...] [TASK [ARGS...]]")
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, h("Tasks:"))
	fmt.Fprintln(&b, "  Tasks are defined in your Runefile and run dynamically — they are not")
	fmt.Fprintln(&b, "  listed below. Run 'rune --list' to see them, then 'rune <task> [args]' to")
	fmt.Fprintln(&b, "  run one. A task whose name collides with a command stays reachable via")
	fmt.Fprintln(&b, "  'rune -- <task>'.")
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, h("Commands:"))
	width := 0
	for _, c := range cmd.Commands() {
		if c.Hidden {
			continue
		}
		if len(c.Name()) > width {
			width = len(c.Name())
		}
	}
	for _, c := range cmd.Commands() {
		if c.Hidden {
			continue
		}
		fmt.Fprintf(&b, "  %-*s  %s\n", width, c.Name(), c.Short)
	}
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, h("Examples:"))
	for _, ex := range []struct{ cmd, note string }{
		{"rune --list", "show the tasks in your Runefile"},
		{"rune build", "run the 'build' task"},
		{"rune build --watch", "flags after the task name go to the task"},
		{"rune --choose", "pick a task interactively"},
		{"rune -- test", "run a task whose name shadows a command"},
		{"rune serve", "expose tasks to AI agents over MCP"},
	} {
		fmt.Fprintf(&b, "  %-26s # %s\n", ex.cmd, ex.note)
	}
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, h("Flags:"))
	b.WriteString(cmd.Flags().FlagUsages())

	return b.String()
}
