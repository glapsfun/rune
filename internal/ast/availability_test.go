package ast

import (
	"reflect"
	"testing"
)

func taskWithAttrs(kinds ...string) *Task {
	t := &Task{Name: "t"}
	for _, k := range kinds {
		t.Attributes = append(t.Attributes, &Attribute{Kind: k})
	}
	return t
}

// TestAvailableOn covers the full truth table from the spec's data model.
// Attributes written on separate lines in a Runefile land as separate
// entries in Task.Attributes, so the multi-entry cases below cover both
// `[linux, windows]` and stacked single-attribute lines.
func TestAvailableOn(t *testing.T) {
	tests := []struct {
		name                   string
		attrs                  []string
		linux, darwin, windows bool
	}{
		{"no attributes", nil, true, true, true},
		{"linux", []string{AttrLinux}, true, false, false},
		{"macos", []string{AttrMacos}, false, true, false},
		{"windows", []string{AttrWindows}, false, false, true},
		{"unix", []string{AttrUnix}, true, true, false},
		{"linux or windows", []string{AttrLinux, AttrWindows}, true, false, true},
		{"unix or windows", []string{AttrUnix, AttrWindows}, true, true, true},
		{"non-OS attrs ignored", []string{AttrGroup, AttrPrivate}, true, true, true},
		{"os attr among non-OS attrs", []string{AttrGroup, AttrLinux, AttrConfirm}, true, false, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			task := taskWithAttrs(tc.attrs...)
			for goos, want := range map[string]bool{
				"linux":   tc.linux,
				"darwin":  tc.darwin,
				"windows": tc.windows,
			} {
				if got := task.AvailableOn(goos); got != want {
					t.Errorf("%s: AvailableOn(%q) = %v, want %v", tc.name, goos, got, want)
				}
			}
		})
	}
}

func TestOSFilters(t *testing.T) {
	tests := []struct {
		name  string
		attrs []string
		want  []string
	}{
		{"empty for unrestricted", nil, nil},
		{"empty for non-OS attrs", []string{AttrGroup, AttrParallel}, nil},
		{"single", []string{AttrWindows}, []string{AttrWindows}},
		{"source order preserved", []string{AttrWindows, AttrLinux, AttrUnix}, []string{AttrWindows, AttrLinux, AttrUnix}},
		{"non-OS attrs skipped in place", []string{AttrGroup, AttrMacos, AttrConfirm, AttrLinux}, []string{AttrMacos, AttrLinux}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := taskWithAttrs(tc.attrs...).OSFilters(); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("OSFilters() = %v, want %v", got, tc.want)
			}
		})
	}
}
