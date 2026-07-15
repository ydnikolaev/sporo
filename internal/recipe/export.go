package recipe

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// Origin says which corpus a recipe came from — and it is not decoration. An OFFICIAL recipe
// is somebody else's build, shipped read-only in the binary; a recipe in the PROJECT's own
// home is this repo's build, authored here and owned here. Printing them in one
// undifferentiated list would invite an author to edit the first kind, which the next update
// would silently eat.
type Origin string

const (
	Official Origin = "official" // shipped in the binary, read-only here
	Project  Origin = "project"  // authored in this repository, owned by it
)

// Entry is one recipe, located.
type Entry struct {
	Slug   string
	Origin Origin
}

// AdoptionDoc carries the two sections that are identical in every recipe: how to get the
// capability into a repository the author has never seen, and what to send back afterwards.
// They are NOT authored per recipe — that would be forty lines of boilerplate copied into
// every document, drifting apart the moment one of them is improved — so they live once and
// the delivery step appends them.
const AdoptionDoc = "_adoption.md"

// Export hands ONE self-contained file to an agent in a repository that has never heard
// of this harness. That is the whole delivery contract, and it is why the fleet corpus is
// compiled into the binary: a recipe that can only be read from a checkout of its origin
// is a recipe that never leaves home.
//
// The exported file is COMPOSED, not copied: the recipe's own body, then the adoption
// protocol. That composition is the point. A recipe is a payload with no consumption path —
// it says what the capability rests on and never says what to do when the reader's ground
// is missing half of it. The protocol is that path, and appending it here (rather than
// asking each author to restate it) is what keeps it one text for the whole corpus.
//
// The project's OWN home is searched first. A recipe about the repository you are standing
// in is the one you most want to hand out, and if a slug exists in both, the local build is
// the one its author verified.
//
// The provenance banner is stripped on the way out. Inside the fleet it marks a file as
// SSOT-authored and warns against editing a synced copy; to a stranger it is noise about
// a repository they do not have — and the first line of a transferable document is the
// worst place to talk about yourself.
func Export(corpus fs.FS, home, slug string) (string, error) {
	if strings.HasPrefix(slug, "_") {
		return "", fmt.Errorf("%q is the genre's own shape spec, not a recipe: it teaches how to "+
			"WRITE one, and exporting it to someone who wants to BUILD the thing helps nobody", slug)
	}
	protocol, err := adoption(corpus)
	if err != nil {
		return "", err
	}
	if home != "" {
		if b, err := os.ReadFile(filepath.Join(home, slug+".md")); err == nil {
			return strip(string(b)) + protocol, nil
		}
	}
	b, err := fs.ReadFile(corpus, path.Join("recipes", slug+".md"))
	if err != nil {
		var known []string
		for _, e := range list(corpus, home) {
			known = append(known, string(e.Origin)+":"+e.Slug)
		}
		return "", fmt.Errorf("no recipe %q (known: %s)", slug, strings.Join(known, ", "))
	}
	return strip(string(b)) + protocol, nil
}

// adoption reads the protocol from the compiled-in corpus and drops everything above its
// first heading — the banner and the note to the corpus's own authors are house business,
// and the reader is not in the house.
//
// It fails CLOSED. A binary whose corpus has lost the protocol would otherwise hand a
// stranger a recipe that looks complete and quietly omits the only section telling them
// what to do when their ground does not match the author's — which is the state every
// reader is actually in.
func adoption(corpus fs.FS) (string, error) {
	b, err := fs.ReadFile(corpus, path.Join("recipes", AdoptionDoc))
	if err != nil {
		return "", fmt.Errorf("the adoption protocol is missing from the corpus (%s): every exported "+
			"recipe carries it, and a recipe without it has no consumption path — the binary is "+
			"broken, not the recipe: %w", AdoptionDoc, err)
	}
	lines := strings.Split(string(b), "\n")
	for i, l := range lines {
		if strings.HasPrefix(l, "## ") {
			// A rule, and it is not decoration. The recipe ends in whatever its author chose —
			// often the appendix, which is deliberately the one section full of the author's own
			// coordinates. Landing straight from that into instructions the reader EXECUTES,
			// with nothing between them, is how a reader mistakes one for the other. The break
			// says: the document above is finished; what follows is addressed to you.
			return "\n---\n\n" + strings.Join(lines[i:], "\n"), nil
		}
	}
	return "", fmt.Errorf("the adoption protocol carries no section (%s)", AdoptionDoc)
}

// Genre prints the authoring spec from the compiled-in corpus, banner stripped. It exists
// because the spec is the one document a CONSUMER repository needs and does not have: the
// skill says "read the genre before you write a line", and in a repo that never checked out
// this source, the binary is the only place the genre lives. Export refuses `_`-prefixed
// documents for a good reason (a stranger asked for a capability, not a style guide), so the
// spec gets its own door instead of a hole in that rule.
func Genre(corpus fs.FS) (string, error) {
	b, err := fs.ReadFile(corpus, path.Join("recipes", "_authoring.md"))
	if err != nil {
		return "", fmt.Errorf("the genre spec is missing from the corpus — the binary is broken, not your repository: %w", err)
	}
	return strip(string(b)), nil
}

// List enumerates BOTH corpora — the fleet's and this project's. `_`-prefixed files are the
// genre's own meta-documents and are not recipes.
func List(corpus fs.FS, home string) ([]Entry, error) {
	if _, err := fs.ReadDir(corpus, "recipes"); err != nil {
		return nil, fmt.Errorf("read the recipe corpus: %w", err)
	}
	return list(corpus, home), nil
}

func list(corpus fs.FS, home string) []Entry {
	seen := map[string]bool{}
	var out []Entry
	if home != "" {
		if ents, err := os.ReadDir(home); err == nil {
			for _, e := range ents {
				if slug, ok := recipeSlug(e.Name(), e.IsDir()); ok {
					seen[slug] = true
					out = append(out, Entry{slug, Project})
				}
			}
		}
	}
	if ents, err := fs.ReadDir(corpus, "recipes"); err == nil {
		for _, e := range ents {
			if slug, ok := recipeSlug(e.Name(), e.IsDir()); ok && !seen[slug] {
				out = append(out, Entry{slug, Official})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Slug < out[j].Slug })
	return out
}

// IsRecipe: a `.md` in a recipes home is a recipe UNLESS it is one of the two documents that
// live there without instantiating the genre — the `_`-prefixed meta-document, and the home's
// own README (which `sporo init` may seed). Excluding the README is not tidiness: without it, a
// project that has merely RUN the tool has a red gate for a change nobody made, on a file the
// tool wrote. A gate that reds on the state it ships is the cry-wolf failure the validation
// discipline forbids — and this genre's own first recipe carries that exact scar.
func IsRecipe(name string, isDir bool) bool {
	return !isDir && strings.HasSuffix(name, ".md") &&
		!strings.HasPrefix(name, "_") && name != "README.md"
}

func recipeSlug(name string, isDir bool) (string, bool) {
	if !IsRecipe(name, isDir) {
		return "", false
	}
	return strings.TrimSuffix(name, ".md"), true
}

func strip(s string) string {
	lines := strings.Split(s, "\n")
	for len(lines) > 0 && (strings.HasPrefix(lines[0], "<!-- SSOT SOURCE") || strings.TrimSpace(lines[0]) == "") {
		lines = lines[1:]
	}
	return strings.Join(lines, "\n")
}
