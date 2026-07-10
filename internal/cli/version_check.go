package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/rune-task-runner/rune/internal/config"
	"github.com/rune-task-runner/rune/internal/parser"
)

// CompatibilityResult is the machine-readable result of `rune version --check`.
type CompatibilityResult struct {
	Installed  string `json:"installed"`
	Required   string `json:"required"` // empty when no requirement is declared
	Compatible bool   `json:"compatible"`
	// Development is true when the installed version is not a recognized semantic
	// version (a local "dev" build): the requirement is waved through rather than
	// enforced, so a machine consumer can tell this apart from a genuine match.
	Development bool   `json:"development"`
	Runefile    string `json:"runefile"` // resolved path, empty when none found
}

// VersionCheck implements `rune version --check [--json]`: it resolves the
// applicable Runefile, reads its minimum_version, and reports compatibility of
// the installed binary. It runs no task. When no Runefile or no requirement is
// present it reports "no requirement declared" (exit 0). An installed version
// that does not satisfy a valid requirement exits non-zero (ExitValidation).
func VersionCheck(opts Options, asJSON bool) error {
	runefile, err := config.Resolve(opts.File, opts.Cwd)
	if err != nil {
		// Discovery found no Runefile: no requirement to check (FR-022). Any other
		// failure (e.g. an explicit --file that doesn't exist) is a usage error,
		// mirroring the run path rather than falsely reporting "compatible".
		if errors.Is(err, config.ErrNotFound) {
			return reportNoRequirement(opts, asJSON, "")
		}
		return &UsageError{Err: err}
	}
	source, err := os.ReadFile(runefile)
	if err != nil {
		return &UsageError{Err: err}
	}
	file, _ := parser.Parse(runefile, string(source))
	req, diags := config.MinimumVersion(file)
	if diags.HasErrors() {
		renderDiags(opts, diags, newSourceProvider(runefile, source))
		return &ValidationError{Err: errorf("invalid minimum_version")}
	}
	if !req.Present {
		return reportNoRequirement(opts, asJSON, runefile)
	}

	ok, dev := req.Satisfied(opts.Version)
	res := CompatibilityResult{
		Installed:   opts.Version,
		Required:    req.Raw,
		Compatible:  ok,
		Development: dev,
		Runefile:    runefile,
	}
	if asJSON {
		if err := printJSON(opts, res); err != nil {
			return err
		}
	} else {
		status := "compatible"
		switch {
		case dev:
			status = "compatible (development build)"
		case !ok:
			status = "incompatible"
		}
		fmt.Fprintf(opts.Stdout, "Runefile requires >= %s\n", req.Raw)
		fmt.Fprintf(opts.Stdout, "Installed Rune: %s\n", opts.Version)
		fmt.Fprintf(opts.Stdout, "Status: %s\n", status)
	}
	if !ok {
		return &ValidationError{Err: errorf("installed Rune %s does not satisfy >= %s", opts.Version, req.Raw)}
	}
	return nil
}

// reportNoRequirement prints the "no requirement declared" result (exit 0).
func reportNoRequirement(opts Options, asJSON bool, runefile string) error {
	if asJSON {
		return printJSON(opts, CompatibilityResult{
			Installed:  opts.Version,
			Compatible: true,
			Runefile:   runefile,
		})
	}
	fmt.Fprintln(opts.Stdout, "no requirement declared")
	fmt.Fprintf(opts.Stdout, "Installed Rune: %s\n", opts.Version)
	return nil
}

// printJSON writes res as indented JSON, mapping a marshal failure to a
// UsageError so callers can simply return it.
func printJSON(opts Options, res CompatibilityResult) error {
	data, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return &UsageError{Err: err}
	}
	fmt.Fprintln(opts.Stdout, string(data))
	return nil
}
