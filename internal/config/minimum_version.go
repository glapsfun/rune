package config

import (
	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/diag"
	"github.com/rune-task-runner/rune/internal/semver"
	"github.com/rune-task-runner/rune/internal/token"
)

// UpgradeURL is where an incompatibility diagnostic points users to obtain a
// newer Rune binary.
const UpgradeURL = "https://github.com/glapsfun/rune/releases"

// minimumVersionSetting is the public name of the setting that pins the minimum
// required Rune binary release. It is distinct from `rune_version`, which pins
// the Runefile *language* version (see version.go).
const minimumVersionSetting = "minimum_version"

// MinimumRequirement is the minimum Rune version a root Runefile requires.
type MinimumRequirement struct {
	Present bool           // whether the root file declared minimum_version
	Raw     string         // the raw literal value, e.g. "0.8.0"
	Version semver.Version // parsed requirement (valid only when the diags were empty)
	Span    token.Span     // source span of the value literal, for diagnostics
}

// MinimumVersion extracts and validates the root Runefile's `minimum_version`
// setting. It MUST be called on the root file BEFORE imports are spliced, so a
// child/imported file can never impose or relax the requirement (FR-012).
//
// The value must be a static string literal holding a valid single semantic
// version. A non-literal value or an invalid semantic version is returned as a
// diagnostic (with the caret on the offending value); in that case the returned
// requirement's Version is unusable. Absence of the setting yields
// Present=false and no diagnostics.
func MinimumVersion(f *ast.File) (MinimumRequirement, diag.List) {
	s := findSetting(f, minimumVersionSetting)
	if s == nil {
		return MinimumRequirement{}, nil
	}

	lit, ok := s.Value.(*ast.StringLit)
	if !ok {
		// Non-literal (variable/expr), list-valued, or bare boolean form.
		span := s.Sp
		if s.Value != nil {
			span = s.Value.Span()
		}
		var diags diag.List
		diags.Errorf(span, "minimum_version must be a static semantic version")
		return MinimumRequirement{Present: true, Span: span}, diags
	}

	v, err := semver.Parse(lit.Value)
	if err != nil {
		var diags diag.List
		diags.Errorf(lit.Sp, "minimum_version must be a valid semantic version, got %q", lit.Value)
		return MinimumRequirement{Present: true, Raw: lit.Value, Span: lit.Sp}, diags
	}

	return MinimumRequirement{Present: true, Raw: lit.Value, Version: v, Span: lit.Sp}, nil
}

// Satisfied reports whether the installed version meets the requirement. A build
// whose version is not a recognized semantic version — notably a local "dev"
// build — is treated as a development build (dev=true) that is not blocked, so
// `go build` binaries keep working against Runefiles that pin a minimum version;
// integration tests inject a concrete version to exercise real comparisons.
func (req MinimumRequirement) Satisfied(installed string) (ok, dev bool) {
	got, err := semver.Parse(installed)
	if err != nil {
		return true, true
	}
	return got.Satisfies(req.Version), false
}

// findSetting returns the first setting with the given name, or nil.
func findSetting(f *ast.File, name string) *ast.Setting {
	for _, s := range f.Settings {
		if s.Name == name {
			return s
		}
	}
	return nil
}
