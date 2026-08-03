package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rune-task-runner/rune/internal/style"
)

// applyHelp installs Rune's friendly, grouped help on the root command and —
// via Cobra's help-func inheritance — on every subcommand (014 C5). Section
// headings are colorized when stdout is a color terminal (resolved here because
// Cobra does not run PersistentPreRunE for --help); the body stays plain so
// piped help is ANSI-free and informative (FR-019..FR-021).
func applyHelp(root *cobra.Command) {
	root.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		out := cmd.OutOrStdout()
		// PersistentPreRunE (the normal color-resolution path) does not run for
		// --help, so styling is resolved tolerantly here — shared with the
		// failure-banner path (see tolerantTheme in color.go).
		th := tolerantTheme(cmd, out)
		if cmd.HasParent() {
			fmt.Fprint(out, subHelp(cmd, th))
			return
		}
		fmt.Fprint(out, rootHelp(cmd, th))
	})
}

// subHelp renders a subcommand's help in the same grouped shape as the root:
// description, Usage, Aliases/Examples where defined, then flags. The plain
// form (disabled theme) is the reviewed 014 baseline. Sections are derived
// from Cobra metadata, so new commands and flags inherit the layout — and a
// command missing its Long/Example shows up as a visibly thinner screen.
func subHelp(cmd *cobra.Command, th style.Theme) string {
	h := th.Heading.Render
	var b strings.Builder

	desc := cmd.Long
	if desc == "" {
		desc = cmd.Short
	}
	if desc != "" {
		fmt.Fprintln(&b, strings.TrimSpace(desc))
		fmt.Fprintln(&b)
	}

	fmt.Fprintln(&b, h("Usage:"))
	fmt.Fprintln(&b, "  "+cmd.UseLine())
	fmt.Fprintln(&b)

	if len(cmd.Aliases) > 0 {
		fmt.Fprintln(&b, h("Aliases:"))
		fmt.Fprintln(&b, "  "+strings.Join(append([]string{cmd.Name()}, cmd.Aliases...), ", "))
		fmt.Fprintln(&b)
	}

	if cmd.Example != "" {
		fmt.Fprintln(&b, h("Examples:"))
		fmt.Fprintln(&b, cmd.Example)
		fmt.Fprintln(&b)
	}

	// A command group lists its children so they stay discoverable from --help,
	// matching rootHelp's Commands section. Latent today (every subcommand is a
	// leaf), but keeps subHelp correct if a group is ever added.
	if cmd.HasAvailableSubCommands() {
		fmt.Fprintln(&b, h("Commands:"))
		width := 0
		for _, c := range cmd.Commands() {
			if (c.IsAvailableCommand() || c.Name() == "help") && len(c.Name()) > width {
				width = len(c.Name())
			}
		}
		for _, c := range cmd.Commands() {
			if c.IsAvailableCommand() || c.Name() == "help" {
				fmt.Fprintf(&b, "  %-*s  %s\n", width, c.Name(), c.Short)
			}
		}
		fmt.Fprintln(&b)
	}

	if cmd.HasAvailableLocalFlags() {
		fmt.Fprintln(&b, h("Flags:"))
		b.WriteString(cmd.LocalFlags().FlagUsages())
	}
	if cmd.HasAvailableInheritedFlags() {
		if cmd.HasAvailableLocalFlags() {
			fmt.Fprintln(&b)
		}
		fmt.Fprintln(&b, h("Global Flags:"))
		b.WriteString(cmd.InheritedFlags().FlagUsages())
	}

	return b.String()
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
