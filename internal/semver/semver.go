// Package semver is a small, dependency-free Semantic Versioning 2.0.0 parser
// and comparator, sufficient for Rune's `minimum_version` requirement gate. It
// deliberately implements only what the gate needs — parse, precedence compare,
// and a `>=` satisfaction test — rather than a full constraint language.
//
// Precedence follows SemVer 2.0.0: the numeric core is compared first, a
// prerelease ranks below the otherwise-equal release (so 0.9.0-rc.1 < 0.9.0),
// prerelease identifiers are compared field-by-field (numeric identifiers rank
// below alphanumeric ones), and build metadata is ignored entirely.
package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is a parsed semantic version. Build metadata is retained for
// round-tripping but never affects precedence.
type Version struct {
	Major, Minor, Patch int
	Prerelease          []string // dot-separated identifiers; empty for a release
	Build               []string // build metadata; ignored for precedence
}

// Parse parses a semantic version of the form MAJOR.MINOR.PATCH[-pre][+build].
// A leading "v", a partial core (e.g. "0.8"), leading zeros in numeric fields,
// and empty identifiers are rejected.
func Parse(s string) (Version, error) {
	if s == "" {
		return Version{}, fmt.Errorf("empty version")
	}

	core := s
	var pre, build string
	var hasPre, hasBuild bool
	// Build metadata is the tail after the first '+'.
	if i := strings.IndexByte(core, '+'); i >= 0 {
		core, build, hasBuild = core[:i], core[i+1:], true
	}
	// Prerelease is the tail after the first '-' in what remains.
	if i := strings.IndexByte(core, '-'); i >= 0 {
		core, pre, hasPre = core[:i], core[i+1:], true
	}
	// A present-but-empty prerelease/build section (a trailing '-' or '+') is
	// malformed.
	if (hasPre && pre == "") || (hasBuild && build == "") {
		return Version{}, fmt.Errorf("version %q has an empty prerelease or build section", s)
	}

	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("version %q must have MAJOR.MINOR.PATCH", s)
	}
	var v Version
	for i, dst := range []*int{&v.Major, &v.Minor, &v.Patch} {
		n, err := parseNumeric(parts[i])
		if err != nil {
			return Version{}, fmt.Errorf("version %q: %w", s, err)
		}
		*dst = n
	}

	if pre != "" {
		ids := strings.Split(pre, ".")
		for _, id := range ids {
			if id == "" || !isIdentifier(id) {
				return Version{}, fmt.Errorf("version %q: invalid prerelease identifier %q", s, id)
			}
			// A purely numeric prerelease identifier must not have leading zeros.
			if isNumeric(id) && len(id) > 1 && id[0] == '0' {
				return Version{}, fmt.Errorf("version %q: numeric prerelease identifier %q has a leading zero", s, id)
			}
		}
		v.Prerelease = ids
	}

	if build != "" {
		ids := strings.Split(build, ".")
		for _, id := range ids {
			if id == "" || !isIdentifier(id) {
				return Version{}, fmt.Errorf("version %q: invalid build identifier %q", s, id)
			}
		}
		v.Build = ids
	}

	return v, nil
}

// parseNumeric parses a non-negative integer, rejecting leading zeros.
func parseNumeric(s string) (int, error) {
	if s == "" || !isNumeric(s) {
		return 0, fmt.Errorf("numeric field %q is not a number", s)
	}
	if len(s) > 1 && s[0] == '0' {
		return 0, fmt.Errorf("numeric field %q has a leading zero", s)
	}
	return strconv.Atoi(s)
}

// Compare returns -1, 0, or +1 as v is less than, equal to, or greater than o in
// SemVer precedence. Build metadata is ignored.
func (v Version) Compare(o Version) int {
	if c := cmpInt(v.Major, o.Major); c != 0 {
		return c
	}
	if c := cmpInt(v.Minor, o.Minor); c != 0 {
		return c
	}
	if c := cmpInt(v.Patch, o.Patch); c != 0 {
		return c
	}
	return comparePrerelease(v.Prerelease, o.Prerelease)
}

// Satisfies reports whether v meets the minimum requirement min (v >= min).
func (v Version) Satisfies(min Version) bool { return v.Compare(min) >= 0 }

// String renders the version in canonical form.
func (v Version) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d.%d.%d", v.Major, v.Minor, v.Patch)
	if len(v.Prerelease) > 0 {
		b.WriteByte('-')
		b.WriteString(strings.Join(v.Prerelease, "."))
	}
	if len(v.Build) > 0 {
		b.WriteByte('+')
		b.WriteString(strings.Join(v.Build, "."))
	}
	return b.String()
}

// comparePrerelease compares two prerelease identifier lists. A release (empty
// list) outranks any prerelease.
func comparePrerelease(a, b []string) int {
	switch {
	case len(a) == 0 && len(b) == 0:
		return 0
	case len(a) == 0: // a is a release, b is a prerelease
		return 1
	case len(b) == 0:
		return -1
	}
	for i := 0; i < len(a) && i < len(b); i++ {
		if c := compareIdentifier(a[i], b[i]); c != 0 {
			return c
		}
	}
	// All shared identifiers equal: the longer list has higher precedence.
	return cmpInt(len(a), len(b))
}

// compareIdentifier compares two prerelease identifiers: numeric identifiers
// compare numerically and rank below alphanumeric ones; alphanumeric compare
// lexically in ASCII order.
func compareIdentifier(a, b string) int {
	an, bn := isNumeric(a), isNumeric(b)
	switch {
	case an && bn:
		// Numeric identifiers carry no leading zeros (enforced in Parse), so the
		// longer digit string is the larger number; equal length compares
		// lexically. This avoids strconv overflow on very large identifiers.
		if len(a) != len(b) {
			return cmpInt(len(a), len(b))
		}
		return strings.Compare(a, b)
	case an: // numeric < alphanumeric
		return -1
	case bn:
		return 1
	default:
		return strings.Compare(a, b)
	}
}

func cmpInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// isIdentifier reports whether s consists only of [0-9A-Za-z-].
func isIdentifier(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9', c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c == '-':
		default:
			return false
		}
	}
	return true
}
