package analysis

import "path/filepath"

// Workspace is the scope of analysis: a project root, its entry file, and the
// import graph relating its documents. Each workspace folder is an independent
// project in the first release (spec FR-021).
type Workspace struct {
	Root      DocumentURI
	EntryFile DocumentURI
	Imports   ImportGraph
}

// ImportGraph records which files import which, in both directions, so a change
// to one file can invalidate its transitive importers (spec FR-022).
type ImportGraph struct {
	ImportsByFile  map[DocumentURI][]DocumentURI
	ImportedByFile map[DocumentURI][]DocumentURI
}

// NewImportGraph returns an empty graph.
func NewImportGraph() ImportGraph {
	return ImportGraph{
		ImportsByFile:  map[DocumentURI][]DocumentURI{},
		ImportedByFile: map[DocumentURI][]DocumentURI{},
	}
}

// AddEdge records that importer imports imported.
func (g *ImportGraph) AddEdge(importer, imported DocumentURI) {
	g.ImportsByFile[importer] = appendUnique(g.ImportsByFile[importer], imported)
	g.ImportedByFile[imported] = appendUnique(g.ImportedByFile[imported], importer)
}

// TransitiveImporters returns every file that imports uri directly or
// indirectly (the set to re-analyze when uri changes, FR-022).
func (g ImportGraph) TransitiveImporters(uri DocumentURI) []DocumentURI {
	seen := map[DocumentURI]bool{}
	var out []DocumentURI
	var walk func(u DocumentURI)
	walk = func(u DocumentURI) {
		for _, importer := range g.ImportedByFile[u] {
			if seen[importer] {
				continue
			}
			seen[importer] = true
			out = append(out, importer)
			walk(importer)
		}
	}
	walk(uri)
	return out
}

func appendUnique(list []DocumentURI, v DocumentURI) []DocumentURI {
	for _, x := range list {
		if x == v {
			return list
		}
	}
	return append(list, v)
}

// DetectRoot resolves a workspace root using the spec's order (FR-021): an
// explicit client workspace folder, then the nearest ancestor directory
// containing a Runefile, then the nearest containing a .git, then the document's
// own directory. exists probes whether a path is present (injected so this is
// testable without a real filesystem).
func DetectRoot(explicit DocumentURI, docPath string, exists func(path string) bool) DocumentURI {
	if explicit != "" {
		return explicit
	}
	docDir := filepath.Dir(docPath)
	if dir, ok := nearestContaining(docDir, "Runefile", exists); ok {
		return dir
	}
	if dir, ok := nearestContaining(docDir, ".git", exists); ok {
		return dir
	}
	return docDir
}

// nearestContaining walks up from dir looking for a directory that contains
// name, returning that directory.
func nearestContaining(dir, name string, exists func(string) bool) (string, bool) {
	for {
		if exists(filepath.Join(dir, name)) {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false // reached filesystem root
		}
		dir = parent
	}
}
