package recipe

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sporo.dev/sporo/pkg/recipekit"
)

// LintSeed checks one seed against the seed genre — the seed-kind sibling of Lint. It forwards to
// the generic engine with SeedShape and injects NO extra: unlike a recipe, a seed declares no
// exact-bound contracts (its trust contract — detect-first, per-step acceptance, a runnable Verify,
// no uncited pipe — is enforced inside SeedShape's own body checks, not a consumer-parsed fence),
// so there is no conform-layer lint half to inject.
func LintSeed(name string, src []byte, products []string) []Finding {
	return recipekit.LintShape(recipekit.SeedShape, name, src, products)
}

// LintSeedHome lints every seed in a project's seed home against the seed genre and appends the
// seed-scoped seal-coherence sweep — the seed counterpart of the recipe lint path (recipe.Lint per
// file + the coherence half of VerifyRegistry), scoped to KindSeed through the SHARED
// verifySealCoherence helper so a sealed seed is checked in the seed home and a flat recipe verb
// never sees it (INV-1). It returns the findings, the number of seeds checked, and the number of
// drafts skipped (the counts the CLI reports).
//
// A project that declares no seed home has no seed corpus to walk — a stated absence, returned as
// a clean error, never a crash (REQ-5): a repository that authors recipes but not seeds simply has
// nothing here. Drafts are exempt exactly as recipes are: reding on the state `sporo seed new`
// writes would train red-blindness. The `_`-prefixed genre meta-documents (the runner preamble, the
// authoring spec) are held to their banner alone — the engine's `_`-prefix early return does that.
//
// This is the coherence (registry → home) half only. VerifyRegistry's second, recipe-inline sweep —
// every FINISHED recipe in the home must be sealed — is deliberately not reproduced here: it is not
// part of the shared helper, and the seed corpus is meta-doc-only until S5 seals the first real
// seed, so there is nothing finished-but-unsealed to find yet.
func LintSeedHome(root string, cfg Config) ([]Finding, int, int, error) {
	home, ok := cfg.HomeFor(recipekit.KindSeed)
	if !ok {
		return nil, 0, 0, fmt.Errorf("this project declares no seed corpus — `sporo seed` authors seeds under a `homes: {seed: …}` home; declare one, then `sporo seed new` to start")
	}
	dir := filepath.Join(root, home)
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("sporo seed lint: no seed corpus at %s — author one there with `sporo seed new`, or declare the seed home in the project config", dir)
	}
	var findings []Finding
	n, drafts := 0, 0
	for _, e := range ents {
		if !isSeed(e.Name(), e.IsDir()) && !isSeedMeta(e.Name()) {
			continue
		}
		src, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, 0, 0, err
		}
		if IsDraft(src) {
			drafts++
			continue
		}
		n++
		findings = append(findings, LintSeed(e.Name(), src, cfg.Products)...)
	}
	reg, err := LoadRegistry(root)
	if err != nil {
		return nil, 0, 0, err
	}
	findings = append(findings, verifySealCoherence(root, home, reg, recipekit.KindSeed)...)
	return findings, n, drafts, nil
}

// isSeed: a `.md` in a seed home is a seed UNLESS it is a `_`-prefixed genre meta-document or the
// home's own README. It mirrors IsRecipe for the seed corpus — a second predicate rather than a
// parameterized shared one, because the recipe branch is INV-1-frozen and two four-clause
// predicates are cheaper to read than a genre-tagged abstraction with one caller each.
func isSeed(name string, isDir bool) bool {
	return !isDir && strings.HasSuffix(name, ".md") &&
		!strings.HasPrefix(name, "_") && name != "README.md"
}

// isSeedMeta reports the seed genre's own `_`-prefixed documents (the runner preamble, the authoring
// spec). The walk must still hand them to the linter — the meta-document IS checked for its banner —
// so isSeed excludes them and this second predicate lets them back in, banner-only.
func isSeedMeta(name string) bool { return len(name) > 0 && name[0] == '_' }
