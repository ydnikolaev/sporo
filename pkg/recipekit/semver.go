package recipekit

import (
	"strconv"
	"strings"
)

// SemverNewer reports whether candidate is a strictly higher semver than current. A version
// that does not parse as a triple is never "newer" — the caller decides what an unorderable
// version means; here it simply loses.
func SemverNewer(candidate, current string) bool {
	a, okA := SemverTriple(candidate)
	b, okB := SemverTriple(current)
	if !okA || !okB {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}

// SemverTriple parses a `MAJOR.MINOR.PATCH` version (a leading `v` and a `-pre`/`+build`
// suffix on the patch are tolerated) into its three integers, reporting false when the shape
// is not a triple.
func SemverTriple(v string) ([3]int, bool) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return [3]int{}, false
	}
	var out [3]int
	for i, p := range parts {
		if i == 2 {
			if j := strings.IndexAny(p, "-+"); j >= 0 {
				p = p[:j]
			}
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return [3]int{}, false
		}
		out[i] = n
	}
	return out, true
}
