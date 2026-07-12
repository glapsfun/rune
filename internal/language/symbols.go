package language

import (
	"sort"

	"github.com/rune-task-runner/rune/internal/ast"
	"github.com/rune-task-runner/rune/internal/token"
)

// DocumentOutline is a document's symbols grouped into the categories an editor
// outline shows: settings, variables, imports, modules, and tasks (spec User
// Story 6). Each entry carries the span to navigate to.
type DocumentOutline struct {
	Groups []OutlineGroup
}

// OutlineGroup is one named category of outline entries.
type OutlineGroup struct {
	Name    string
	Kind    SymbolKind
	Entries []OutlineEntry
}

// OutlineEntry is a single symbol in the outline.
type OutlineEntry struct {
	Name      string
	Detail    string     // e.g. a task signature
	Selection token.Span // the declaration span to navigate to
}

// Outline builds the document outline for the given file's own declarations
// (imported/module symbols are attributed to their own files, so they are
// excluded here — the outline reflects one document).
func Outline(f *ast.File, file string) DocumentOutline {
	var settings, variables, imports, modules, tasks []OutlineEntry
	if f == nil {
		return DocumentOutline{}
	}

	for _, s := range f.Settings {
		if s.Sp.File != file {
			continue
		}
		settings = append(settings, OutlineEntry{Name: s.Name, Detail: "setting", Selection: s.Sp})
	}
	for _, a := range f.Assignments {
		if a.Sp.File != file {
			continue
		}
		variables = append(variables, OutlineEntry{Name: a.Name, Detail: "variable", Selection: a.Sp})
	}
	for _, im := range f.Imports {
		if im.Sp.File != file {
			continue
		}
		imports = append(imports, OutlineEntry{Name: im.Path, Detail: "import", Selection: im.Sp})
	}
	for _, m := range f.Mods {
		if m.Sp.File != file {
			continue
		}
		modules = append(modules, OutlineEntry{Name: m.Name, Detail: "module", Selection: m.Sp})
	}
	for _, t := range f.Tasks {
		if t.Sp.File != file {
			continue
		}
		tasks = append(tasks, OutlineEntry{Name: t.Name, Detail: TaskSignature(t), Selection: t.Sp})
	}

	out := DocumentOutline{}
	add := func(name string, kind SymbolKind, entries []OutlineEntry) {
		if len(entries) > 0 {
			out.Groups = append(out.Groups, OutlineGroup{Name: name, Kind: kind, Entries: entries})
		}
	}
	add("settings", SymbolSetting, settings)
	add("variables", SymbolVariable, variables)
	add("imports", SymbolImport, imports)
	add("modules", SymbolModule, modules)
	add("tasks", SymbolTask, tasks)
	return out
}

// SortOutlineByPosition orders entries within each group by source position
// (declaration order), which editors expect for an outline.
func SortOutlineByPosition(o DocumentOutline) {
	for _, g := range o.Groups {
		sort.SliceStable(g.Entries, func(i, j int) bool {
			return g.Entries[i].Selection.Start.Offset < g.Entries[j].Selection.Start.Offset
		})
	}
}
