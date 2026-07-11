package language

// scopeOf returns the scope a name reference belongs to. A task's parameters
// live in that task's scope (keyed by task name); module variables live in
// ModuleScope. This is the minimal scope model the MVP needs — completion and
// definition use it to decide which variables/parameters are visible. A full
// ScopeTree (for nested scopes) is deferred until needed.
func (ix *Index) VisibleVariables(scope ScopeID) []Symbol {
	var out []Symbol
	// Module variables are visible everywhere.
	for _, s := range ix.byKind(SymbolVariable) {
		out = append(out, s)
	}
	// Parameters of the enclosing task, when in a task scope.
	if scope != ModuleScope {
		for _, s := range ix.byKind(SymbolParameter) {
			if s.Scope == scope {
				out = append(out, s)
			}
		}
	}
	return out
}

// byKind returns all indexed symbols of a kind (declaration order preserved by
// insertion into ByDocument).
func (ix *Index) byKind(kind SymbolKind) []Symbol {
	var out []Symbol
	for _, syms := range ix.ByDocument {
		for _, s := range syms {
			if s.Kind == kind {
				out = append(out, s)
			}
		}
	}
	return out
}
