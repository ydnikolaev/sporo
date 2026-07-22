package recipe

import (
	"os"
	"path/filepath"
	"testing"

	"sporo.dev/sporo/pkg/recipekit"
)

// Pre-flight is the gate that keeps an unverified recipe out of the corpus — so the tests that
// matter are the REFUSALS: an attestation over an unsealed or drifted recipe would be exactly the
// decorative claim the provenance feature exists to avoid.

func TestPublishPreflightPassesASealedGatePassedRecipe(t *testing.T) {
	root, cfg := sealFixture(t)
	if _, err := Seal(root, cfg, "baseline"); err != nil {
		t.Fatal(err)
	}
	res, err := PublishPreflight(root, "baseline")
	if err != nil {
		t.Fatalf("a sealed, gate-passed recipe must pass pre-flight, got: %v", err)
	}
	if res.Slug != "baseline" || res.Hash != res.Entry.Hash {
		t.Fatalf("pre-flight returns the sealed subject with its seal hash; got %+v", res)
	}
}

func TestPublishPreflightRefusesAnUnsealedRecipe(t *testing.T) {
	root, _ := sealFixture(t) // written to the home, never sealed
	if _, err := PublishPreflight(root, "baseline"); err == nil {
		t.Fatal("an unsealed recipe must be refused — an attestation may only cover a sealed recipe")
	}
}

func TestPublishPreflightRefusesADriftedSeal(t *testing.T) {
	root, cfg := sealFixture(t)
	if _, err := Seal(root, cfg, "baseline"); err != nil {
		t.Fatal(err)
	}
	// The silent-mutation case: change the sealed file without bumping the version.
	path := filepath.Join(root, cfg.Home, "baseline.md")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(b, []byte("\nsilently added after the seal.\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := PublishPreflight(root, "baseline"); err == nil {
		t.Fatal("a recipe that drifted from its seal must be refused before publishing")
	}
}

// seedPublishFixture stands up a project whose seed home holds a genre-conformant seal, ready to seal
// and publish — the seed analogue of sealFixture. Both homes are declared so HomeFor(seed) resolves.
func seedPublishFixture(t *testing.T) (root string, cfg Config) {
	t.Helper()
	root = t.TempDir()
	cfg = Config{
		Home:  ".sporo/recipes/",
		Homes: map[string]string{recipekit.KindRecipe: ".sporo/recipes/", recipekit.KindSeed: "seeds/"},
	}
	writeSeedProject(t, root, "widget", fixtureSeed)
	return root, cfg
}

// writeSeedProject lays down the on-disk state PublishPreflight reads: a `.sporo/config.yaml` that
// declares the seed home (LoadConfig seeds only the recipe home by default, so HomeFor(seed) needs
// this) and the seed document in that home.
func writeSeedProject(t *testing.T, root, slug, doc string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, ".sporo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".sporo", "config.yaml"), []byte("home: .sporo/recipes/\nhomes:\n  seed: seeds/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	seedHome := filepath.Join(root, "seeds")
	if err := os.MkdirAll(seedHome, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(seedHome, slug+".md"), []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}
}

// publish is kind-polymorphic: a sealed, gate-passed SEED must pass the same pre-flight, read from
// the seed home and gated by the seed genre — not the recipe one.
func TestPublishPreflightPassesASealedGatePassedSeed(t *testing.T) {
	root, cfg := seedPublishFixture(t)
	if _, err := SealKind(root, cfg, recipekit.KindSeed, "widget"); err != nil {
		t.Fatal(err)
	}
	res, err := PublishPreflight(root, "widget")
	if err != nil {
		t.Fatalf("a sealed, gate-passed seed must pass pre-flight, got: %v", err)
	}
	if res.Entry.Kind != recipekit.KindSeed {
		t.Fatalf("pre-flight must carry the sealed kind so the corpus attests the right home; got %q", res.Entry.Kind)
	}
}

// The gate is dispatched by kind: a recipe-shaped document sealed as a seed passes the RECIPE gate
// but must be refused by the SEED gate — proving publish runs LintSeed for a seed, not Lint.
func TestPublishPreflightRefusesASeedThatFailsTheSeedGate(t *testing.T) {
	root := t.TempDir()
	cfg := Config{
		Home:  ".sporo/recipes/",
		Homes: map[string]string{recipekit.KindRecipe: ".sporo/recipes/", recipekit.KindSeed: "seeds/"},
	}
	// `conformant` is a recipe-shaped fixture: it passes Lint, but has none of the seed genre's
	// required shape, so LintSeed must find fault.
	writeSeedProject(t, root, "widget", conformant)
	if _, err := SealKind(root, cfg, recipekit.KindSeed, "widget"); err != nil {
		t.Fatal(err)
	}
	if _, err := PublishPreflight(root, "widget"); err == nil {
		t.Fatal("a seed that fails the seed genre gate must be refused — publish must run LintSeed for a seed, not Lint")
	}
}
