package cli

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// chooseAndRun presents an interactive task picker (--choose), then runs the
// selected task. It uses fzf when available, else a minimal built-in picker.
func chooseAndRun(opts Options, runefile string, args []string) error {
	mod, err := loadModule(opts, runefile)
	if err != nil {
		return err
	}
	var names []string
	for _, t := range mod.file.Tasks {
		if !t.IsPrivate() {
			names = append(names, t.Name)
		}
	}
	if len(names) == 0 {
		return usagef("no tasks to choose from")
	}
	picked, err := pickTask(opts, names)
	if err != nil {
		return err
	}
	if picked == "" {
		return nil // nothing selected
	}
	return execute(opts, runefile, append([]string{picked}, args...))
}

func pickTask(opts Options, names []string) (string, error) {
	if path, err := exec.LookPath("fzf"); err == nil {
		return pickWithFzf(path, opts, names)
	}
	return pickBuiltin(opts, names)
}

func pickWithFzf(path string, opts Options, names []string) (string, error) {
	cmd := exec.CommandContext(opts.ctx(), path, "--prompt", "task> ")
	cmd.Stdin = strings.NewReader(strings.Join(names, "\n"))
	cmd.Stderr = opts.Stderr
	out, err := cmd.Output()
	if err != nil {
		return "", nil // user aborted the picker
	}
	return strings.TrimSpace(string(out)), nil
}

func pickBuiltin(opts Options, names []string) (string, error) {
	if opts.Stdin == nil {
		return "", usagef("--choose requires an interactive terminal")
	}
	fmt.Fprintln(opts.Stderr, "Select a task:")
	for i, n := range names {
		fmt.Fprintf(opts.Stderr, "  %d) %s\n", i+1, n)
	}
	fmt.Fprint(opts.Stderr, "> ")
	line, _ := bufio.NewReader(opts.Stdin).ReadString('\n')
	idx, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || idx < 1 || idx > len(names) {
		return "", usagef("invalid selection")
	}
	return names[idx-1], nil
}
