package cli

import (
	"encoding/json"
	"fmt"

	"github.com/rune-task-runner/rune/internal/analysis"
	"github.com/rune-task-runner/rune/internal/config"
	"github.com/rune-task-runner/rune/internal/diag"
	"github.com/rune-task-runner/rune/internal/token"
)

// Analyze statically analyzes the Runefile at path (or the discovered Runefile
// when path is empty) together with its transitive imports, printing diagnostics
// and running nothing (spec FR-023). It uses the shared analysis service, so its
// diagnostics are identical to what the language server publishes (FR-002).
//
// Exit codes follow Rune's house scheme: 0 when there are no error-severity
// diagnostics, 3 (ValidationError) when there are, and 2 (UsageError) for a
// missing/unreadable Runefile or other discovery/IO failure.
func Analyze(opts Options, path string, jsonOut bool) error {
	runefile, err := config.Resolve(path, opts.Cwd)
	if err != nil {
		return &UsageError{Err: err}
	}

	svc := analysis.NewService(analysis.DiskSourceStore{})
	snap, err := svc.Analyze(opts.ctx(), analysis.AnalyzeRequest{URI: runefile})
	if err != nil {
		return &UsageError{Err: err}
	}

	if jsonOut {
		if err := printAnalyzeJSON(opts, snap.Diagnostics); err != nil {
			return &UsageError{Err: err}
		}
	} else {
		printAnalyzeText(opts, snap.Sources, snap.Diagnostics)
	}

	if snap.Diagnostics.HasErrors() {
		return &ValidationError{Err: errorf("%d error(s)", countErrors(snap.Diagnostics))}
	}
	return nil
}

// printAnalyzeText renders diagnostics to stdout (stdout is the analysis
// product; incidental logs go to stderr) with the same renderer the run path
// uses — severity color, faint locator, caret span — keeping analyze's coded
// severity token (014 C4). The summary's severity words are emphasized only
// when their count is non-zero. It renders against src — the exact bytes the
// diagnostics were computed on (analysis.Snapshot.Sources) — so the caret can
// never point at a stale line and no file is re-read.
//
// Analyze deliberately does NOT value-mask: it is a static tool whose output
// must stay a pure, reproducible function of the Runefile (env-derived masking
// would make it depend on the caller's shell and break CI goldens), and
// length-changing substitutions would misalign the carets it renders. Secret
// values reach output only if hardcoded into the Runefile, which the feature
// documents as out of scope (secrets come from the environment, never a
// Runefile); the source it echoes is on-disk repo content either way.
func printAnalyzeText(opts Options, src diag.SourceProvider, diags diag.List) {
	out := opts.Stdout
	th := opts.themeStdout()
	if len(diags) > 0 {
		fmt.Fprintln(out, diag.RenderAllCoded(diags, src, th))
	}
	errs, warns := countBySeverity(diags)
	eWord, wWord := plural(errs, "error"), plural(warns, "warning")
	if errs > 0 {
		eWord = th.Error.Render(eWord)
	}
	if warns > 0 {
		wWord = th.Warning.Render(wWord)
	}
	fmt.Fprintf(out, "%s, %s\n", eWord, wWord)
}

// --- JSON output ---

type jsonPos struct {
	Line   int `json:"line"`   // 1-based
	Column int `json:"column"` // 1-based (byte)
	Offset int `json:"offset"` // 0-based byte offset
}

type jsonRange struct {
	Start jsonPos `json:"start"`
	End   jsonPos `json:"end"`
}

type jsonRelated struct {
	File    string    `json:"file"`
	Message string    `json:"message,omitempty"`
	Range   jsonRange `json:"range"`
}

type jsonDiagnostic struct {
	File     string        `json:"file"`
	Code     string        `json:"code,omitempty"`
	Severity string        `json:"severity"`
	Message  string        `json:"message"`
	Range    jsonRange     `json:"range"`
	Related  []jsonRelated `json:"related,omitempty"`
}

type jsonReport struct {
	Diagnostics []jsonDiagnostic `json:"diagnostics"`
	Errors      int              `json:"errors"`
	Warnings    int              `json:"warnings"`
}

func printAnalyzeJSON(opts Options, diags diag.List) error {
	report := jsonReport{Diagnostics: make([]jsonDiagnostic, 0, len(diags))}
	for _, d := range diags {
		report.Diagnostics = append(report.Diagnostics, jsonDiagnostic{
			File:     d.Span.File,
			Code:     d.Code,
			Severity: d.Severity.String(),
			Message:  d.Message,
			Range:    spanToJSON(d.Span),
			Related:  relatedToJSON(d.Related),
		})
	}
	report.Errors, report.Warnings = countBySeverity(diags)

	enc := json.NewEncoder(opts.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

func spanToJSON(s token.Span) jsonRange {
	return jsonRange{
		Start: jsonPos{Line: s.Start.Line, Column: s.Start.Col, Offset: s.Start.Offset},
		End:   jsonPos{Line: s.End.Line, Column: s.End.Col, Offset: s.End.Offset},
	}
}

func relatedToJSON(rel []diag.RelatedLocation) []jsonRelated {
	if len(rel) == 0 {
		return nil
	}
	out := make([]jsonRelated, 0, len(rel))
	for _, r := range rel {
		out = append(out, jsonRelated{File: r.Span.File, Message: r.Message, Range: spanToJSON(r.Span)})
	}
	return out
}

// --- helpers ---

func countBySeverity(diags diag.List) (errors, warnings int) {
	for _, d := range diags {
		if d.Severity == diag.Error {
			errors++
		} else {
			warnings++
		}
	}
	return errors, warnings
}

func plural(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}
