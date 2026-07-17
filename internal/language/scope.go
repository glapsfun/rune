package language

// byKind returns all indexed symbols of a kind. This is the minimal scope
// support the MVP needs: completion and definition combine it with a symbol's
// Scope (a task name, or ModuleScope) to decide which variables/parameters are
// visible. A full ScopeTree for nested scopes is deferred until needed.
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
