package analyzer

import (
	"fmt"

	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/diag"
	"github.com/rune-task-runner/rune/internal/language"
)

// CheckSettings reports RUNE2008 for any setting whose name is not a recognized
// Rune setting, consulting the shared language registry so validation and
// completion agree by construction (spec FR-027 / R9).
//
// Like CheckDocumentation, it is SEPARATE from Analyze: it is surfaced by
// `rune analyze` and the language server, but not on the execution path, where
// an unknown setting has always been silently ignored. Keeping it out of Analyze
// preserves backward compatibility (an existing Runefile with a stray setting
// keeps running) while still flagging the likely typo in the editor.
func CheckSettings(f *ast.File) diag.List {
	var diags diag.List
	if f == nil {
		return diags
	}
	for _, s := range f.Settings {
		if !language.IsValid(language.BuiltinSetting, s.Name) {
			diags.Add(diag.Diagnostic{
				Severity: diag.Error,
				Span:     s.Sp,
				Message:  fmt.Sprintf("invalid setting %q", s.Name),
				Code:     diag.CodeInvalidSetting,
			})
		}
	}
	return diags
}
