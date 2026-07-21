package recipe

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sporo.dev/sporo/pkg/recipekit"
)

// FleetRow is one authored recipe, located across the machine — the shape `sporo recipes` prints
// and, under `--json`, hands a script. The field names are a wire contract the future registry
// and `publish` may lean on; keep them stable.
type FleetRow struct {
	Repo       string `json:"repo"` // filepath.Base(root) — the human label
	Root       string `json:"root"` // absolute path, for scripts and disambiguation
	Slug       string `json:"slug"`
	ID         string `json:"id"`
	Version    string `json:"version"`
	Provenance string `json:"provenance"`
	// Status is set only by the `--all` view: "sealed", "unsealed" (finished but not sealed), or
	// "draft". The default (registry-only) index leaves it empty — every row there is sealed.
	Status string `json:"status,omitempty"`
}

// AdoptedRow is one recipe this machine PULLED IN — the reader-side counterpart of FleetRow, for
// `sporo recipes --adopted`. It anchors the version and source `sporo pull` re-checks, not an id
// or provenance (an adopted recipe's identity lives in the repo that authored it).
type AdoptedRow struct {
	Repo    string `json:"repo"`
	Root    string `json:"root"`
	Slug    string `json:"slug"`
	Version string `json:"version"`
	Source  string `json:"source"`
	Date    string `json:"date"`
}

// FleetRecipes derives the machine-wide recipe index from the given project roots by reading each
// one's sealed registry. It is DERIVED, never stored: the answer is assembled from the registries
// live on every call, so it cannot drift the way a second ledger would.
//
// It reads the AUTHOR side (`reg.Recipes` — recipes made in each repo) and deliberately skips the
// reader side (`reg.Adopted` — recipes pulled in): "which recipes did I author locally" is the
// question this answers. A root with no registry (never sealed, or the repo is gone) contributes
// nothing — `LoadRegistry` already degrades a missing file to empty, so a stale project root drops
// out silently rather than erroring the whole index.
//
// The registry is one slug-keyed map across kinds, so a sealed SEED lives here too — but this is
// the authored-RECIPE index, so a non-recipe entry is skipped (INV-1 read-side; the seed CLI lists
// seeds, `sporo recipes` does not).
func FleetRecipes(roots []string) []FleetRow {
	var rows []FleetRow
	for _, root := range roots {
		reg, err := LoadRegistry(root)
		if err != nil {
			// A malformed registry in one repo must not blind the index to every other repo.
			continue
		}
		for slug, entry := range reg.Recipes {
			if entry.Kind != recipekit.KindRecipe {
				continue
			}
			rows = append(rows, FleetRow{
				Repo:       filepath.Base(root),
				Root:       root,
				Slug:       slug,
				ID:         entry.ID,
				Version:    entry.Version,
				Provenance: entry.Provenance,
			})
		}
	}
	sortFleet(rows)
	return rows
}

// FleetAll is the `--all` view: every AUTHORED recipe across the machine, sealed or not. Unlike
// FleetRecipes (which reads only the registry), it scans each project's recipes home — via that
// project's config — so a finished-but-unsealed recipe and a draft both show up, marked by Status.
// A sealed recipe takes its id/version from the registry (the witnessed values); an unsealed one
// takes them from its frontmatter (what it currently claims).
func FleetAll(roots []string) []FleetRow {
	var rows []FleetRow
	for _, root := range roots {
		cfg, err := LoadConfig(root)
		if err != nil {
			continue
		}
		reg, err := LoadRegistry(root)
		if err != nil {
			continue
		}
		home := filepath.Join(root, cfg.Home)
		ents, err := os.ReadDir(home)
		if err != nil {
			continue // no home in this repo — nothing to scan
		}
		for _, e := range ents {
			if !IsRecipe(e.Name(), e.IsDir()) {
				continue
			}
			slug := strings.TrimSuffix(e.Name(), ".md")
			row := FleetRow{Repo: filepath.Base(root), Root: root, Slug: slug, Provenance: "local"}
			// The lookup is kind-guarded: this walks the RECIPE home, so a registry entry for this
			// slug counts as its seal only when it is a recipe. A sealed seed sharing the slug (a
			// different home entirely) must not lend its seal/version/provenance to the recipe file —
			// fall through to the frontmatter/unsealed branch instead (INV-1 read-side).
			if entry, sealed := reg.Recipes[slug]; sealed && entry.Kind == recipekit.KindRecipe {
				row.ID, row.Version, row.Provenance, row.Status = entry.ID, entry.Version, entry.Provenance, "sealed"
			} else {
				src, err := os.ReadFile(filepath.Join(home, e.Name()))
				if err != nil {
					continue
				}
				row.ID, row.Version = fmValue(src, "id"), fmValue(src, "version")
				if IsDraft(src) {
					row.Status = "draft"
				} else {
					row.Status = "unsealed"
				}
			}
			rows = append(rows, row)
		}
	}
	sortFleet(rows)
	return rows
}

// FleetAdopted is the `--adopted` view: every recipe this machine pulled in, across all projects,
// read from each registry's reader-side ledger. Same derive-not-store discipline as FleetRecipes.
func FleetAdopted(roots []string) []AdoptedRow {
	var rows []AdoptedRow
	for _, root := range roots {
		reg, err := LoadRegistry(root)
		if err != nil {
			continue
		}
		for slug, entry := range reg.Adopted {
			rows = append(rows, AdoptedRow{
				Repo:    filepath.Base(root),
				Root:    root,
				Slug:    slug,
				Version: entry.Version,
				Source:  entry.Source,
				Date:    entry.Date,
			})
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Repo != rows[j].Repo {
			return rows[i].Repo < rows[j].Repo
		}
		return rows[i].Slug < rows[j].Slug
	})
	return rows
}

func sortFleet(rows []FleetRow) {
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Repo != rows[j].Repo {
			return rows[i].Repo < rows[j].Repo
		}
		return rows[i].Slug < rows[j].Slug
	})
}
