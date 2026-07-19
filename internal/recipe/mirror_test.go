package recipe

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// WriteMirror is the generator behind the site's committed recipe `.md` mirror. Its whole reason
// to exist is that the site serves what the binary exports WITHOUT recomposing it, so the two
// cannot drift — which only holds if every file it writes is byte-for-byte the `Export` form.
// These teeth assert exactly that, plus the two ways the set can change: a stale file is dropped,
// and the count reflects the corpus.
func TestWriteMirrorWritesTheExportFormPerRecipe(t *testing.T) {
	corpus := fstest.MapFS{
		"recipes/_adoption.md": {Data: []byte(adoptionFixture)},
		"recipes/alpha.md":     {Data: []byte("<!-- SSOT SOURCE -->\n\n---\nname: alpha\n---\n# Alpha\n")},
		"recipes/beta.md":      {Data: []byte("<!-- SSOT SOURCE -->\n\n---\nname: beta\n---\n# Beta\n")},
	}
	dir := t.TempDir()
	// A stale form from a recipe that no longer exists must not survive a regeneration.
	stale := filepath.Join(dir, "gone.md")
	if err := os.WriteFile(stale, []byte("orphan"), 0o644); err != nil {
		t.Fatal(err)
	}

	n, err := WriteMirror(corpus, "", dir)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("wrote %d forms, want 2 (alpha, beta — never the `_`-prefixed spec)", n)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatal("a form whose recipe was removed must be deleted, or the mirror serves a recipe the corpus lost")
	}
	for _, slug := range []string{"alpha", "beta"} {
		want, err := Export(corpus, "", slug)
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(filepath.Join(dir, slug+".md"))
		if err != nil {
			t.Fatalf("mirror is missing %s.md: %v", slug, err)
		}
		if string(got) != want {
			t.Fatalf("%s.md is not byte-for-byte the Export form — the whole point of committing it is that it is:\n got: %q\nwant: %q", slug, got, want)
		}
	}
}
