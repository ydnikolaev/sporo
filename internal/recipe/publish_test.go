package recipe

import (
	"os"
	"path/filepath"
	"testing"
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
