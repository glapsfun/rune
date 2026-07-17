package language

import (
	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/token"
)

// Definition resolves the declaration location(s) for the symbol at offset. It
// returns the declaration span(s) and true when a navigable definition exists.
// Attributes and built-ins have no source location (their docs surface via
// hover), so they return false.
func Definition(ix *Index, f *ast.File, file string, offset int) ([]token.Span, bool) {
	target, ok := TargetAt(f, file, offset)
	if !ok {
		return nil, false
	}
	switch target.Kind {
	case TargetDependency:
		if sym, ok := lookupTask(ix, target.Name); ok {
			return []token.Span{sym.Selection}, true
		}
	case TargetVarRef:
		if sp, ok := resolveVarOrParam(ix, target.Name, target.Scope); ok {
			return []token.Span{sp}, true
		}
	case TargetTaskName:
		if sym, ok := lookupTask(ix, target.Name); ok {
			return []token.Span{sym.Selection}, true
		}
	case TargetParamDecl:
		return []token.Span{target.Span}, true
	}
	return nil, false
}

// lookupTask finds a task by qualified name, then by base name.
func lookupTask(ix *Index, name string) (Symbol, bool) {
	if s, ok := ix.ByQualified[name]; ok {
		return s, true
	}
	for _, s := range ix.ByName[baseName(name)] {
		if s.Kind == SymbolTask {
			return s, true
		}
	}
	return Symbol{}, false
}

// resolveVarOrParam resolves a name to a parameter in the given task scope, or
// falls back to a module variable.
func resolveVarOrParam(ix *Index, name string, scope ScopeID) (token.Span, bool) {
	if scope != ModuleScope {
		for _, s := range ix.byKind(SymbolParameter) {
			if s.Name == name && s.Scope == scope {
				return s.Definition, true
			}
		}
	}
	for _, s := range ix.ByName[name] {
		if s.Kind == SymbolVariable {
			return s.Definition, true
		}
	}
	return token.Span{}, false
}
