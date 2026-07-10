package analyzer

import (
	"fmt"

	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/diag"
)

// CheckDocumentation reports a warning (RUNE2010) for every public task that has
// no documentation comment (spec FR-008a).
//
// It is intentionally SEPARATE from Analyze: this is a warning surfaced by
// `rune analyze` and the language server, not during normal execution. Keeping
// it out of Analyze means it never gates a run (warnings never set HasErrors)
// and never adds noise to `rune` output. The analysis service combines Analyze
// and CheckDocumentation into a single snapshot.
func CheckDocumentation(f *ast.File) diag.List {
	var diags diag.List
	for _, t := range f.Tasks {
		if t.IsPrivate() || t.Doc != "" {
			continue
		}
		diags.Add(diag.Warn(t.Sp, fmt.Sprintf("public task %q has no documentation", t.Name)).
			WithCode(diag.CodeUndocumentedTask))
	}
	return diags
}
