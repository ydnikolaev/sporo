package recipe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// seedRegistry writes a .sporo/registry.yaml directly — the honest way to test the fleet index.
// Driving it through `new`+`seal` would fight the draft refusal and bury the test in scaffold
// boilerplate; the index only reads what Save writes, so Save is the right seam.
func seedRegistry(t *testing.T, root string, recipes map[string]RegistryEntry) {
	t.Helper()
	reg := Registry{Recipes: recipes}
	if err := reg.Save(root); err != nil {
		t.Fatal(err)
	}
}

func TestFleetRecipesMergesAndSortsAcrossRoots(t *testing.T) {
	// Two roots with predictable basenames so we can assert the (Repo, Slug) sort.
	alpha := filepath.Join(t.TempDir(), "alpha-repo")
	zeta := filepath.Join(t.TempDir(), "zeta-repo")
	seedRegistry(t, alpha, map[string]RegistryEntry{
		"gamma-recipe": {ID: "01ARZ3NDEKTSV4RRFFQ69G5FAV", Version: "1.0.0", Provenance: "local"},
		"beta-recipe":  {ID: "01BX5ZZKBKACTAV9WEVGEMMVRZ", Version: "2.1.0", Provenance: "local"},
	})
	seedRegistry(t, zeta, map[string]RegistryEntry{
		"omega-recipe": {ID: "01CX5ZZKBKACTAV9WEVGEMMV00", Version: "0.3.0", Provenance: "community"},
	})

	rows := FleetRecipes([]string{zeta, alpha}) // pass out of order; result must be sorted
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows across two repos, got %d: %+v", len(rows), rows)
	}
	// Sorted by (Repo, Slug): alpha-repo/beta, alpha-repo/gamma, zeta-repo/omega.
	want := []struct{ repo, slug, id, version string }{
		{"alpha-repo", "beta-recipe", "01BX5ZZKBKACTAV9WEVGEMMVRZ", "2.1.0"},
		{"alpha-repo", "gamma-recipe", "01ARZ3NDEKTSV4RRFFQ69G5FAV", "1.0.0"},
		{"zeta-repo", "omega-recipe", "01CX5ZZKBKACTAV9WEVGEMMV00", "0.3.0"},
	}
	for i, w := range want {
		if rows[i].Repo != w.repo || rows[i].Slug != w.slug || rows[i].ID != w.id || rows[i].Version != w.version {
			t.Fatalf("row %d = %+v, want %v", i, rows[i], w)
		}
	}
}

func TestFleetRecipesSkipsAdoptedAndMissingRoots(t *testing.T) {
	root := t.TempDir()
	reg := Registry{
		Recipes: map[string]RegistryEntry{
			"authored-here": {ID: "01ARZ3NDEKTSV4RRFFQ69G5FAV", Version: "1.0.0", Provenance: "local"},
		},
		// Adopted recipes are the READER side — a fleet-of-authored index must not surface them.
		Adopted: map[string]AdoptedEntry{
			"pulled-in": {Version: "9.9.9"},
		},
	}
	if err := reg.Save(root); err != nil {
		t.Fatal(err)
	}
	rows := FleetRecipes([]string{root, filepath.Join(root, "does-not-exist")})
	if len(rows) != 1 || rows[0].Slug != "authored-here" {
		t.Fatalf("only the authored recipe must appear (no adopted, no crash on a missing root): %+v", rows)
	}
}

func TestFleetRecipesEmptyInputIsEmpty(t *testing.T) {
	if rows := FleetRecipes(nil); len(rows) != 0 {
		t.Fatalf("no roots means no rows, got %+v", rows)
	}
}

func TestFleetAllMarksSealedUnsealedAndDraft(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, ".sporo", "recipes") // the default home LoadConfig resolves
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(home, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("sealed-one.md", conformant)   // in the registry below → sealed
	write("unsealed-one.md", conformant) // finished, absent from the registry → unsealed
	write("draft-one.md", strings.Replace(conformant, "effort: reference", "effort: reference\ndraft: true", 1))
	seedRegistry(t, root, map[string]RegistryEntry{
		"sealed-one": {ID: "01ARZ3NDEKTSV4RRFFQ69G5FAV", Version: "1.0.0", Provenance: "local"},
	})

	status := map[string]string{}
	for _, r := range FleetAll([]string{root}) {
		status[r.Slug] = r.Status
	}
	if status["sealed-one"] != "sealed" || status["unsealed-one"] != "unsealed" || status["draft-one"] != "draft" {
		t.Fatalf("--all must mark each recipe's seal status; got %+v", status)
	}
}

func TestFleetAdoptedReadsTheReaderSideOnly(t *testing.T) {
	root := filepath.Join(t.TempDir(), "consumer-repo")
	reg := Registry{
		Recipes: map[string]RegistryEntry{"authored-here": {Version: "1.0.0"}},
		Adopted: map[string]AdoptedEntry{
			"pulled-in": {Version: "2.3.0", Source: "https://example.test/x.md", Date: "2026-07-18"},
		},
	}
	if err := reg.Save(root); err != nil {
		t.Fatal(err)
	}
	rows := FleetAdopted([]string{root})
	if len(rows) != 1 || rows[0].Slug != "pulled-in" || rows[0].Version != "2.3.0" || rows[0].Source != "https://example.test/x.md" {
		t.Fatalf("--adopted must surface the reader side only (no authored recipe): %+v", rows)
	}
}
