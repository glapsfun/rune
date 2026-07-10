package cli

import (
	"fmt"

	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/config"
	"github.com/rune-task-runner/rune/internal/diag"
)

// enforceMinimumVersion checks the root Runefile's minimum_version requirement
// against the installed version. It MUST run after parsing but BEFORE imports
// are spliced, analysis, or any execution (FR-004): reading the root file's own
// settings pre-compose is what guarantees a child/imported file can never impose
// or relax the requirement (FR-012).
//
// A malformed requirement (a non-static value or an invalid semantic version) is
// always a hard validation error. An installed version older than a valid
// requirement is an incompatibility: when allowIgnore is true it is downgraded to
// a warning and execution proceeds; otherwise it aborts with a rendered
// diagnostic and exit 3, having run nothing.
func enforceMinimumVersion(opts Options, file *ast.File, src diag.SourceProvider, allowIgnore bool) error {
	req, diags := config.MinimumVersion(file)
	if diags.HasErrors() {
		renderDiags(opts, diags, src)
		return &ValidationError{Err: errorf("%d static error(s)", countErrors(diags))}
	}
	if !req.Present {
		return nil
	}
	if ok, _ := req.Satisfied(opts.Version); ok {
		return nil
	}
	if allowIgnore {
		fmt.Fprintf(opts.Stderr, "warning: ignoring Runefile minimum Rune version %s; running %s\n", req.Raw, opts.Version)
		return nil
	}
	renderVersionMismatch(opts, req, src)
	return &ValidationError{Err: errorf("this Runefile requires Rune >= %s", req.Raw)}
}

// renderVersionMismatch prints the incompatibility diagnostic: a caret-anchored
// header pointing at the minimum_version value, followed by the installed,
// required, and upgrade notes and a "nothing was executed" trailer.
func renderVersionMismatch(opts Options, req config.MinimumRequirement, src diag.SourceProvider) {
	d := diag.New(req.Span, fmt.Sprintf("this Runefile requires Rune >= %s", req.Raw))
	fmt.Fprintln(opts.Stderr, diag.RenderAll(diag.List{d}, src, opts.themeStderr()))
	fmt.Fprintf(opts.Stderr, "    installed version: %s\n", opts.Version)
	fmt.Fprintf(opts.Stderr, "    required version:  %s\n", req.Raw)
	fmt.Fprintf(opts.Stderr, "    upgrade:           %s\n", config.UpgradeURL)
	fmt.Fprintln(opts.Stderr, "\nnothing was executed")
}
