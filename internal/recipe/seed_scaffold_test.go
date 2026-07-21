package recipe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sporo.dev/sporo/pkg/recipekit"
)

// The seed scaffold's two load-bearing properties, asserted as hard as any gate — the seed-kind
// mirror of scaffold_test.go: it is born a draft (so the tool never reds on its own output, and
// seal/export refuse it), and — minus the draft mark — it is SEED-GENRE-GREEN.

func seedScaffoldWorld(t *testing.T) (root string, cfg Config) {
	t.Helper()
	root = t.TempDir()
	// Products deliberately excludes every word that appears in the scaffold body or coach comments
	// (the CLI verbs it names, the fixture's tool) — the neutrality scan reds on a product name
	// wherever it appears, so a colliding product would fail the green test as if the scaffold were
	// malformed. `otherproj` is the recipe scaffold test's own precedent.
	return root, Config{
		Home:     ".sporo/recipes/",
		Homes:    map[string]string{recipekit.KindRecipe: ".sporo/recipes/", recipekit.KindSeed: "seeds/"},
		Products: []string{"otherproj"},
	}
}

func TestTheSeedScaffoldIsBornADraftAndRefusedBySealAndExport(t *testing.T) {
	root, cfg := seedScaffoldWorld(t)
	path, err := SeedScaffold(root, cfg, "my-tool", "")
	if err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !IsDraft(src) {
		t.Fatal("the seed scaffold must declare itself a draft — otherwise the gate reds on the state the tool itself wrote")
	}
	if _, err := SealKind(root, cfg, recipekit.KindSeed, "my-tool"); err == nil || !strings.Contains(err.Error(), "draft") {
		t.Fatalf("seal must refuse a draft seed — a draft has no version to promise: %v", err)
	}
	home, _ := cfg.HomeFor(recipekit.KindSeed)
	if _, err := SeedExport(realCorpus(t), filepath.Join(root, home), "my-tool"); err == nil || !strings.Contains(err.Error(), "draft") {
		t.Fatalf("export must refuse a draft — a stranger must never receive TODOs as if they were earned: %v", err)
	}
}

func TestTheSeedScaffoldMinusTheDraftMarkIsSeedGenreGreen(t *testing.T) {
	root, cfg := seedScaffoldWorld(t)
	path, err := SeedScaffold(root, cfg, "my-tool", "Install a capability worth having twice")
	if err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	finished := strings.Replace(string(src), "draft: true\n", "", 1)
	if f := LintSeed("my-tool.md", []byte(finished), cfg.Products); len(f) != 0 {
		t.Fatalf("the seed scaffold is the first conformant document the author reads — it may not fail its own gate:\n%v", f)
	}
	// The id is minted, not typed — present and a real ULID the moment the draft is born, or the
	// author hits a gate on a field they were never meant to touch. Lint above enforces the grammar;
	// here we assert it is minted and of ULID length.
	if id := fmValue(src, "id"); len(id) != 26 {
		t.Fatalf("the scaffold must mint a 26-char ULID id, got %q", id)
	}
}

func TestTheSeedScaffoldNeverOverwrites(t *testing.T) {
	root, cfg := seedScaffoldWorld(t)
	if _, err := SeedScaffold(root, cfg, "my-tool", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := SeedScaffold(root, cfg, "my-tool", ""); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("a second scaffold over a real draft would eat the author's work: %v", err)
	}
}

func TestSeedScaffoldRefusesAProjectWithNoSeedHome(t *testing.T) {
	root := t.TempDir()
	cfg := Config{Home: ".sporo/recipes/", Homes: map[string]string{recipekit.KindRecipe: ".sporo/recipes/"}}
	if _, err := SeedScaffold(root, cfg, "my-tool", ""); err == nil {
		t.Fatal("a project that declares no seed home has nowhere to scaffold a seed — it must error cleanly, never crash")
	}
}
