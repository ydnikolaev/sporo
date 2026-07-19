package recipe

import (
	"io/fs"
	"os"
	"path/filepath"
)

// WriteMirror regenerates outDir to hold exactly one `<slug>.md` per recipe, each the
// byte-for-byte `Export` handover form (banner stripped, adoption protocol appended). It is the
// generator behind the site's recipe `.md` mirror: committing the composed form — rather than
// recomposing it in the site's own code — keeps `Export` the ONE place the composition lives, so
// a `go generate` + git-diff gate can prove the served mirror never drifts from the binary's
// export. A file whose recipe no longer exists is removed, so a deletion reds the gate too, the
// same way a new recipe reds it until its file is committed. Returns the number of forms written.
func WriteMirror(corpus fs.FS, home, outDir string) (int, error) {
	entries := list(corpus, home)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return 0, err
	}
	want := make(map[string]bool, len(entries))
	for _, e := range entries {
		want[e.Slug+".md"] = true
	}
	// Drop stale forms first: a recipe removed from the corpus must not leave its export behind,
	// or the mirror would serve a recipe the corpus no longer has.
	existing, err := filepath.Glob(filepath.Join(outDir, "*.md"))
	if err != nil {
		return 0, err
	}
	for _, f := range existing {
		if !want[filepath.Base(f)] {
			if err := os.Remove(f); err != nil {
				return 0, err
			}
		}
	}
	for _, e := range entries {
		body, err := Export(corpus, home, e.Slug)
		if err != nil {
			return 0, err
		}
		if err := os.WriteFile(filepath.Join(outDir, e.Slug+".md"), []byte(body), 0o644); err != nil {
			return 0, err
		}
	}
	return len(entries), nil
}
