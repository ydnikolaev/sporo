package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sporo.dev/sporo/internal/recipe"
)

// runLint executes the flat `sporo lint` in-process, returning stdout, stderr, and the run error —
// a fresh lintCmd() per call keeps cases independent, mirroring runSeed.
func runLint(t *testing.T, args ...string) (out, errOut string, err error) {
	t.Helper()
	cmd := lintCmd()
	var o, e bytes.Buffer
	cmd.SetOut(&o)
	cmd.SetErr(&e)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return o.String(), e.String(), err
}

// finishedRecipe scaffolds a recipe and removes its draft mark — the scaffold is conformant by
// construction, so the corpus greens on wiring, not on a fixture's shape (the seed tests' precedent).
func finishedRecipe(t *testing.T, root string, slug string) {
	t.Helper()
	cfg, err := recipe.LoadConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	path, err := recipe.Scaffold(root, cfg, slug, "Do the thing", nil)
	if err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	finished := strings.Replace(string(src), "draft: true\n", "", 1)
	if err := os.WriteFile(path, []byte(finished), 0o644); err != nil {
		t.Fatal(err)
	}
}

// BL-006: the flat lint summary counts recipe INSTANCES and `_`-prefixed genre meta-docs apart, so
// it agrees with `sporo list` (which shows only instances). A corpus with a recipe and a meta-doc
// must report "1 recipe(s) and 1 meta-doc(s)", never a folded "2 recipe(s)".
func TestFlatLintCountsMetaDocsApartFromInstances(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".sporo"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := "home: recipes/\nproject: otherproj\n"
	if err := os.WriteFile(filepath.Join(root, ".sporo", "config.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	finishedRecipe(t, root, "widget")
	// A `_`-prefixed meta-doc with a valid provenance banner: held to its banner alone, so it lints
	// clean though it has none of the genre sections — but it counts as a meta-doc, not a recipe.
	banner := "<!-- SSOT SOURCE (otherproj). -->\n\n## Anything\n\nprose, not a recipe\n"
	if err := os.WriteFile(filepath.Join(root, "recipes", "_notes.md"), []byte(banner), 0o644); err != nil {
		t.Fatal(err)
	}

	// dir-arg form skips the registry sweep (an unsealed finished recipe would otherwise red it);
	// --root points the config (neutrality vocabulary) at this fixture, not the real repo.
	out, errOut, err := runLint(t, filepath.Join(root, "recipes"), "--root", root)
	if err != nil {
		t.Fatalf("lint (recipe + meta-doc): %v\nstderr: %s", err, errOut)
	}
	if !strings.Contains(out, "1 recipe(s) and 1 meta-doc(s) conformant and neutral") {
		t.Fatalf("lint summary must count the meta-doc apart from the recipe, got %q", out)
	}
}
