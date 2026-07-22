package mask

import (
	"bytes"
	"sort"
	"strings"
)

const (
	// Placeholder replaces every occurrence of a tracked secret value on every
	// output surface (contract §2.2).
	Placeholder = "***"
	// MinLen is the minimum tracked value (or value-line) length in bytes;
	// shorter values are never value-masked to avoid corrupting unrelated
	// output (FR-007).
	MinLen = 4
)

// namePatterns are the built-in sensitive-name patterns, matched
// case-insensitively as substrings of a variable NAME (contract §2.1).
var namePatterns = []string{
	"TOKEN", "SECRET", "PASSWORD", "PASSWD", "APIKEY", "API_KEY",
	"PRIVATE_KEY", "ACCESS_KEY", "CREDENTIAL", "AUTH",
}

// Set is an immutable collection of secret values to mask. It is safe for
// concurrent readers.
type Set struct {
	entries [][]byte  // unique, sorted longest-first (leftmost-longest matching)
	maxLen  int       // length of the longest entry (carry bound is maxLen-1)
	first   [256]bool // fast reject: bytes that can start an entry
}

// NewSet derives the secret-value set from KEY=value environment pairs.
// A variable is tracked when its name matches a built-in pattern and is not
// exempted, or when it is explicitly declared. Values are split per line;
// lines shorter than MinLen are dropped (data-model derivation steps 1–4).
func NewSet(env, declared, exempt []string) *Set {
	decl := toNameSet(declared)
	ex := toNameSet(exempt)
	s := &Set{}
	seen := map[string]struct{}{}
	for _, pair := range env {
		eq := strings.IndexByte(pair, '=')
		if eq <= 0 {
			continue
		}
		if !isSecretName(pair[:eq], decl, ex) {
			continue
		}
		for _, line := range strings.Split(pair[eq+1:], "\n") {
			line = strings.TrimSuffix(line, "\r")
			if len(line) < MinLen {
				continue
			}
			if _, dup := seen[line]; dup {
				continue
			}
			seen[line] = struct{}{}
			s.entries = append(s.entries, []byte(line))
		}
	}
	sort.Slice(s.entries, func(i, j int) bool {
		if len(s.entries[i]) != len(s.entries[j]) {
			return len(s.entries[i]) > len(s.entries[j])
		}
		return bytes.Compare(s.entries[i], s.entries[j]) < 0
	})
	if len(s.entries) > 0 {
		s.maxLen = len(s.entries[0])
	}
	for _, e := range s.entries {
		s.first[e[0]] = true
	}
	return s
}

// Empty reports whether the set tracks no values; callers skip wrapping
// writers entirely for an empty set (FR-008).
func (s *Set) Empty() bool { return len(s.entries) == 0 }

// MaskString replaces every occurrence of every entry in one shot (no carry).
func (s *Set) MaskString(in string) string {
	out, _ := s.scan([]byte(in), true)
	return string(out)
}

// scan masks buf, returning the emittable output and the retained tail: the
// leftmost suffix that is a proper prefix of some entry (nil when final is
// true — at end-of-stream an incomplete prefix is, by definition, not the
// secret, but completed entries within the tail are still masked).
func (s *Set) scan(buf []byte, final bool) (out, carry []byte) {
	if len(s.entries) == 0 {
		return buf, nil
	}
	out = make([]byte, 0, len(buf))
	i := 0
	for i < len(buf) {
		if !s.first[buf[i]] {
			j := i + 1
			for j < len(buf) && !s.first[buf[j]] {
				j++
			}
			out = append(out, buf[i:j]...)
			i = j
			continue
		}
		matched := false
		rest := buf[i:]
		// Entries are longest-first: a longer entry that might still complete
		// in the next chunk wins over a shorter full match; Flush re-scans the
		// carry in final mode, so a completed shorter secret can never leak.
		for _, e := range s.entries {
			if len(e) <= len(rest) {
				if bytes.Equal(rest[:len(e)], e) {
					out = append(out, Placeholder...)
					i += len(e)
					matched = true
					break
				}
			} else if !final && bytes.HasPrefix(e, rest) {
				return out, append([]byte(nil), rest...)
			}
		}
		if !matched {
			out = append(out, buf[i])
			i++
		}
	}
	return out, nil
}

func isSecretName(name string, declared, exempt map[string]struct{}) bool {
	if _, ok := declared[name]; ok {
		return true
	}
	if _, ok := exempt[name]; ok {
		return false
	}
	upper := strings.ToUpper(name)
	for _, p := range namePatterns {
		if strings.Contains(upper, p) {
			return true
		}
	}
	return false
}

func toNameSet(names []string) map[string]struct{} {
	if len(names) == 0 {
		return nil
	}
	m := make(map[string]struct{}, len(names))
	for _, n := range names {
		m[n] = struct{}{}
	}
	return m
}
