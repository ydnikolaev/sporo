package main

import "testing"

// The docs surface is the source of truth the site's documentation page renders from, so its
// invariants are gates, not hopes. Two things must always hold, and each caught a real gap
// while being written (`list` had no group; the walk must not silently drop a verb).

// Every verb must live in exactly one taxonomy bucket. A new verb added to root() without a
// home in docGroups falls to "other" — which the docs page would render as an orphan section.
// This is the gate that makes "remember to group the new verb" mechanical.
func TestEveryVerbIsGrouped(t *testing.T) {
	s := buildSurface()
	for _, v := range s.Verbs {
		if v.Group == "other" {
			t.Errorf("verb %q is ungrouped — add it to a bucket in docGroups (docs.go)", v.Name)
		}
	}
}

// The surface must name exactly the commands the binary actually registers — no more (a stale
// hand list), no fewer (a walk that dropped one). This is the check the old hand-maintained
// verb map in internal/install could not make: it compares against the live cobra tree.
// The surface documents exactly the VISIBLE command tree: every registered non-hidden verb, and
// nothing else. Hidden verbs are build tools (`web-mirror`, run only by `go generate`) and must
// stay OUT — this is the bidirectional guard that a hidden command never leaks into the docs page,
// and that no visible verb is dropped.
func TestSurfaceCoversTheWholeCommandTree(t *testing.T) {
	inTree := map[string]bool{}
	for _, c := range root().Commands() {
		if c.Hidden {
			continue
		}
		inTree[c.Name()] = true
	}
	inSurface := map[string]bool{}
	for _, v := range buildSurface().Verbs {
		inSurface[v.Name] = true
	}
	for name := range inTree {
		if !inSurface[name] {
			t.Errorf("command %q is registered but missing from the docs surface", name)
		}
	}
	for name := range inSurface {
		if !inTree[name] {
			t.Errorf("docs surface reports %q, which the binary does not register", name)
		}
	}
}

// The two nested subcommands must survive the walk — the docs page lists `feedback add`,
// `feedback list`, `review verify`, and a walk that stopped at the top level would drop them.
func TestNestedSubcommandsAreCaptured(t *testing.T) {
	want := map[string][]string{"feedback": {"add", "list"}, "review": {"verify"}}
	got := map[string][]string{}
	for _, v := range buildSurface().Verbs {
		for _, sub := range v.Subcommands {
			got[v.Name] = append(got[v.Name], sub.Name)
		}
	}
	for parent, subs := range want {
		for _, sub := range subs {
			found := false
			for _, g := range got[parent] {
				if g == sub {
					found = true
				}
			}
			if !found {
				t.Errorf("%s %s missing from the surface (got %v)", parent, sub, got[parent])
			}
		}
	}
}

func TestGenreVersionFlagIsPartOfTheDocumentedSurface(t *testing.T) {
	for _, verb := range buildSurface().Verbs {
		if verb.Name != "genre" {
			continue
		}
		for _, flag := range verb.Flags {
			if flag.Name == "version" {
				return
			}
		}
		t.Fatal("genre --version is callable but absent from the generated docs surface")
	}
	t.Fatal("genre verb is absent from the generated docs surface")
}
