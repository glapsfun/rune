package language

import "testing"

func TestOutlineGroupsAndEntries(t *testing.T) {
	src := "set shell := [\"sh\", \"-c\"]\noutput_dir := \"dist\"\nimport \"common.rune\"\nmod backend \"backend.rune\"\n# Build.\nbuild:\n    @echo b\n# Test.\ntest: build\n    @echo t\n"
	f, _ := resolve(t, src)
	outline := Outline(f, "Runefile")

	byName := map[string]OutlineGroup{}
	for _, g := range outline.Groups {
		byName[g.Name] = g
	}
	// Import composition tries to read common.rune/backend.rune which don't exist,
	// but the import/mod DECLARATIONS still appear in the outline.
	for _, want := range []string{"settings", "variables", "imports", "modules", "tasks"} {
		if _, ok := byName[want]; !ok {
			t.Errorf("outline missing group %q (groups: %v)", want, groupNames(outline))
		}
	}
	if g := byName["tasks"]; len(g.Entries) != 2 {
		t.Errorf("tasks group has %d entries, want 2", len(g.Entries))
	}
	if g := byName["settings"]; len(g.Entries) != 1 || g.Entries[0].Name != "shell" {
		t.Errorf("settings = %+v, want [shell]", g.Entries)
	}
	if g := byName["variables"]; g.Entries[0].Name != "output_dir" {
		t.Errorf("variables = %+v, want output_dir", g.Entries)
	}
}

func TestOutlineExcludesEmptyGroups(t *testing.T) {
	src := "# B.\nbuild:\n    @echo b\n"
	f, _ := resolve(t, src)
	outline := Outline(f, "Runefile")
	if len(outline.Groups) != 1 || outline.Groups[0].Name != "tasks" {
		t.Errorf("outline = %v, want only tasks group", groupNames(outline))
	}
}

func groupNames(o DocumentOutline) []string {
	var out []string
	for _, g := range o.Groups {
		out = append(out, g.Name)
	}
	return out
}
