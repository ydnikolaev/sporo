package recipe

import (
	"path/filepath"
	"sort"
)

// FleetRow is one sealed recipe, located across the machine — the shape `sporo recipes` prints
// and, under `--json`, hands a script. The field names are a wire contract the future registry
// and `publish` may lean on; keep them stable.
type FleetRow struct {
	Repo       string `json:"repo"` // filepath.Base(root) — the human label
	Root       string `json:"root"` // absolute path, for scripts and disambiguation
	Slug       string `json:"slug"`
	ID         string `json:"id"`
	Version    string `json:"version"`
	Provenance string `json:"provenance"`
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
func FleetRecipes(roots []string) []FleetRow {
	var rows []FleetRow
	for _, root := range roots {
		reg, err := LoadRegistry(root)
		if err != nil {
			// A malformed registry in one repo must not blind the index to every other repo.
			continue
		}
		for slug, entry := range reg.Recipes {
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
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Repo != rows[j].Repo {
			return rows[i].Repo < rows[j].Repo
		}
		return rows[i].Slug < rows[j].Slug
	})
	return rows
}
