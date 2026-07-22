package main

import "testing"

// The compare URL the fallback prints is only correct if owner/repo are parsed right from BOTH
// remote forms — an author with an SSH remote and one with HTTPS must reach the same PR.
func TestParseGithubSlug(t *testing.T) {
	ok := []struct {
		remote, owner, repo string
	}{
		{"git@github.com:ydnikolaev/sporo.git", "ydnikolaev", "sporo"},
		{"git@github.com:ydnikolaev/sporo", "ydnikolaev", "sporo"},
		{"https://github.com/ydnikolaev/sporo.git", "ydnikolaev", "sporo"},
		{"https://github.com/ydnikolaev/sporo", "ydnikolaev", "sporo"},
		{"ssh://git@github.com/ydnikolaev/sporo.git", "ydnikolaev", "sporo"},
	}
	for _, c := range ok {
		owner, repo, err := parseGithubSlug(c.remote)
		if err != nil || owner != c.owner || repo != c.repo {
			t.Errorf("parseGithubSlug(%q) = %q/%q, %v; want %q/%q, nil", c.remote, owner, repo, err, c.owner, c.repo)
		}
	}

	// A non-github or malformed remote must error, not return a half-parsed slug that would print a
	// broken compare URL as if it were real.
	bad := []string{"git@gitlab.com:o/r.git", "https://github.com/onlyowner", "", "not a url"}
	for _, remote := range bad {
		if _, _, err := parseGithubSlug(remote); err == nil {
			t.Errorf("parseGithubSlug(%q) = nil error; want an error for a non-github/malformed remote", remote)
		}
	}
}
