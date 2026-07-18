package recipe

import (
	"path/filepath"
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
