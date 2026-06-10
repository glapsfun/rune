package cli

import (
	"os"
	"runtime"

	"github.com/rune-task-runner/rune/internal/config"
	"github.com/rune-task-runner/rune/internal/parser"
)

// TaskCandidate is a completable task: its name plus the first line of its doc
// comment, used as the shell-completion description.
type TaskCandidate struct {
	Name string
	Doc  string
}

// TaskCandidates returns the non-private, OS-matching tasks of the resolved
// Runefile, each with its first doc line, for dynamic shell completion. It is
// deliberately tolerant: any failure (no Runefile, read error, parse/compose
// error) yields nil so completion degrades gracefully and never disrupts the
// shell session. The analyzer is intentionally skipped — task names should still
// complete when the Runefile has semantic (not syntactic) errors.
func TaskCandidates(opts Options) []TaskCandidate {
	runefile, err := config.Resolve(opts.File, opts.Cwd)
	if err != nil {
		return nil
	}
	source, err := os.ReadFile(runefile)
	if err != nil {
		return nil
	}
	file, diags := parser.Parse(runefile, string(source))
	diags = append(diags, config.Compose(file, newSourceProvider(runefile, source))...)
	if diags.HasErrors() {
		return nil
	}

	var out []TaskCandidate
	for _, t := range file.Tasks {
		if t.IsPrivate() || !osMatches(t, runtime.GOOS) {
			continue
		}
		out = append(out, TaskCandidate{Name: t.Name, Doc: firstLine(t.Doc)})
	}
	return out
}
