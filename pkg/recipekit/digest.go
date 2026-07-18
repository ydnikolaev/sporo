package recipekit

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// ContentHash is the registry's currency: `sha256:<hex>` over exact bytes. The install layer
// prices its managed files in the same currency — one hash discipline, not two.
func ContentHash(b []byte) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(b))
}

// ExactContractsDigest hashes the bodies of the exact-bound fences in the contracts
// section, in order. Only the fence CONTENTS count: prose around a shape can be reworded
// freely, field names and structure cannot. An empty digest means the recipe promises
// nothing exact, and the seal imposes nothing.
func ExactContractsDigest(src []byte) string {
	con := SectionBody(strings.Split(string(src), "\n"), "## The contracts")
	var buf strings.Builder
	exact, inFence, capture := false, false, false
	for _, l := range con {
		switch {
		case reFence.MatchString(l):
			if !inFence {
				inFence, capture = true, exact
			} else {
				inFence, capture = false, false
			}
			exact = false
		case !inFence && reBinding.MatchString(l):
			exact = strings.Contains(l, "**Binding: exact**")
		default:
			if capture {
				buf.WriteString(l)
				buf.WriteByte('\n')
			}
		}
	}
	if buf.Len() == 0 {
		return ""
	}
	return ContentHash([]byte(buf.String()))
}

// SemverMajor reads the MAJOR of a semver triple; a malformed version reads as 0, and the
// lint gate (which requires a real triple) is where malformedness is reported.
func SemverMajor(v string) int {
	n := 0
	for _, r := range v {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	return n
}

// FrontmatterValue extracts one scalar from a recipe's frontmatter, tolerant of quotes. It
// reads the first `---` pair — scanning from line 0, not line 1, because it serves TWO
// document shapes: a source file (banner on line 0, frontmatter after) and an EXPORTED file
// (banner stripped, frontmatter IS line 0). A scan that assumed the banner missed every
// export's frontmatter entirely — `adopt` found that the hard way. Safe for sources too: a
// banner line is an HTML comment and never trims to `---`.
func FrontmatterValue(src []byte, key string) string {
	lines := strings.Split(string(src), "\n")
	start, end := -1, -1
	for i := 0; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			if start < 0 {
				start = i
			} else {
				end = i
				break
			}
		}
	}
	if start < 0 || end < 0 {
		return ""
	}
	line := KeyLine(lines[start+1:end], key)
	if line == "" {
		return ""
	}
	v := strings.TrimSpace(strings.TrimPrefix(line, key+":"))
	return strings.Trim(v, `"'`)
}
