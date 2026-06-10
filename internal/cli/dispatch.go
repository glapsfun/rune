package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/rune-task-runner/rune/internal/config"
)

// errorf is a thin wrapper kept package-local so error construction is uniform.
func errorf(format string, args ...any) error { return fmt.Errorf(format, args...) }

// Options carries the resolved global CLI flags and I/O streams for one
// invocation. main.go populates it from cobra and passes it to Run.
type Options struct {
	File       string // -f/--file
	List       bool   // --list
	DryRun     bool   // --dry-run
	Summary    bool   // --summary
	Dump       bool   // --dump
	DumpFormat string // --format (with --dump)
	Set        []string
	Watch      bool
	Choose     bool
	Yes        bool
	Quiet      bool
	Fmt        bool
	ClearCache bool

	Color   bool // resolved: emit ANSI color
	Version string
	Cwd     string
	Ctx     context.Context // cancelled on SIGINT (nil => Background)
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer

	// Commands are the reserved subcommand names (and aliases). They are used
	// only to enrich the "unknown task" error with a did-you-mean suggestion.
	Commands []string
}

// ctx returns the invocation context, defaulting to Background.
func (o Options) ctx() context.Context {
	if o.Ctx != nil {
		return o.Ctx
	}
	return context.Background()
}

// Run executes one CLI invocation. args is everything after the global flags:
// VAR=VALUE overrides interleaved with task names and their arguments.
//
// The full pipeline (lex → parse → analyze → schedule → execute) is wired in
// run.go; this function resolves the Runefile and delegates.
func Run(opts Options, args []string) error {
	runefile, err := config.Resolve(opts.File, opts.Cwd)
	if err != nil {
		return &UsageError{Err: err}
	}
	if opts.Choose {
		return chooseAndRun(opts, runefile, args)
	}
	if opts.Watch {
		return watch(opts, runefile, args)
	}
	return execute(opts, runefile, args)
}
