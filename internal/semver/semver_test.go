package semver

import "testing"

func TestParseValid(t *testing.T) {
	cases := []struct {
		in   string
		want Version
	}{
		{"0.8.0", Version{Minor: 8}},
		{"1.2.3", Version{Major: 1, Minor: 2, Patch: 3}},
		{"0.9.0-rc.1", Version{Minor: 9, Prerelease: []string{"rc", "1"}}},
		{"0.9.0-dev+13dbf54", Version{Minor: 9, Prerelease: []string{"dev"}, Build: []string{"13dbf54"}}},
		{"1.0.0+build.5", Version{Major: 1, Build: []string{"build", "5"}}},
	}
	for _, c := range cases {
		got, err := Parse(c.in)
		if err != nil {
			t.Errorf("Parse(%q) unexpected error: %v", c.in, err)
			continue
		}
		if got.String() != c.want.String() {
			t.Errorf("Parse(%q) = %q, want %q", c.in, got.String(), c.want.String())
		}
	}
}

func TestParseInvalid(t *testing.T) {
	for _, in := range []string{
		"", "0.8", "v0.8.0", "latest", "1.2.3.4", ">=0.8,<1.0",
		"^0.8", "~0.8.1", "1.2.x", "01.2.3", "1.2.3-", "1.2.3-01",
	} {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) succeeded, want error", in)
		}
	}
}

func TestCompareCore(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"0.8.0", "0.8.0", 0},
		{"0.7.2", "0.8.0", -1},
		{"0.9.1", "0.8.0", 1},
		{"1.0.0", "0.9.9", 1},
		{"0.8.1", "0.8.0", 1},
	}
	for _, c := range cases {
		a := mustParse(t, c.a)
		b := mustParse(t, c.b)
		if got := a.Compare(b); got != c.want {
			t.Errorf("Compare(%q,%q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestPrereleaseRanksBelowRelease(t *testing.T) {
	// SemVer 2.0.0: 0.9.0-rc.1 < 0.9.0 (FR-010).
	pre := mustParse(t, "0.9.0-rc.1")
	rel := mustParse(t, "0.9.0")
	if pre.Compare(rel) != -1 {
		t.Errorf("0.9.0-rc.1 should be < 0.9.0")
	}
	if pre.Satisfies(rel) {
		t.Errorf("0.9.0-rc.1 must NOT satisfy a 0.9.0 requirement")
	}
	// A dev build is a prerelease and likewise does not satisfy the release.
	dev := mustParse(t, "0.9.0-dev+abc")
	if dev.Satisfies(rel) {
		t.Errorf("0.9.0-dev+abc must NOT satisfy a 0.9.0 requirement")
	}
}

func TestPrereleaseOrdering(t *testing.T) {
	// From the SemVer 2.0.0 spec example.
	order := []string{
		"1.0.0-alpha", "1.0.0-alpha.1", "1.0.0-alpha.beta",
		"1.0.0-beta", "1.0.0-beta.2", "1.0.0-beta.11", "1.0.0-rc.1", "1.0.0",
	}
	for i := 0; i+1 < len(order); i++ {
		a := mustParse(t, order[i])
		b := mustParse(t, order[i+1])
		if a.Compare(b) != -1 {
			t.Errorf("%q should be < %q", order[i], order[i+1])
		}
	}
}

func TestNumericPrereleaseNoOverflow(t *testing.T) {
	// Numeric prerelease identifiers larger than int64 must still order correctly
	// (compared by digit-length then lexically, never via strconv).
	small := mustParse(t, "1.0.0-1")
	huge := mustParse(t, "1.0.0-99999999999999999999")
	if small.Compare(huge) != -1 {
		t.Errorf("1.0.0-1 should be < 1.0.0-99999999999999999999")
	}
	a := mustParse(t, "1.0.0-100")
	b := mustParse(t, "1.0.0-99")
	if a.Compare(b) != 1 {
		t.Errorf("1.0.0-100 should be > 1.0.0-99")
	}
}

func TestBuildMetadataIgnored(t *testing.T) {
	a := mustParse(t, "0.9.0+abc")
	b := mustParse(t, "0.9.0")
	if a.Compare(b) != 0 {
		t.Errorf("build metadata must not affect precedence: %q vs %q", "0.9.0+abc", "0.9.0")
	}
	c := mustParse(t, "0.9.0+xyz")
	if a.Compare(c) != 0 {
		t.Errorf("differing build metadata must compare equal")
	}
}

func TestSatisfies(t *testing.T) {
	req := mustParse(t, "0.8.0")
	cases := []struct {
		installed string
		want      bool
	}{
		{"0.7.2", false},
		{"0.8.0", true},
		{"0.8.1", true},
		{"0.9.1", true},
		{"1.0.0", true},
	}
	for _, c := range cases {
		got := mustParse(t, c.installed).Satisfies(req)
		if got != c.want {
			t.Errorf("%q Satisfies 0.8.0 = %v, want %v", c.installed, got, c.want)
		}
	}
}

func mustParse(t *testing.T, s string) Version {
	t.Helper()
	v, err := Parse(s)
	if err != nil {
		t.Fatalf("Parse(%q): %v", s, err)
	}
	return v
}
