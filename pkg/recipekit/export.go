package recipekit

import "strings"

// StripBanner drops the leading provenance banner (and any blank lines above the first real
// content) from a recipe source. Inside the fleet the banner marks a file as SSOT-authored
// and warns against editing a synced copy; to a stranger it is noise about a repository they
// do not have — and the first line of a transferable document is the worst place to talk
// about yourself.
func StripBanner(s string) string {
	lines := strings.Split(s, "\n")
	for len(lines) > 0 && (strings.HasPrefix(lines[0], "<!-- SSOT SOURCE") || strings.TrimSpace(lines[0]) == "") {
		lines = lines[1:]
	}
	return strings.Join(lines, "\n")
}
