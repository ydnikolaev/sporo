package recipe

import (
	"fmt"
	"os"
	"path/filepath"

	"sporo.dev/sporo/pkg/recipekit"
)

// PublishResult is the verified subject a passed pre-flight yields: the sealed recipe the corpus
// attest step may sign. It carries the bytes and the seal hash so a caller never re-reads or
// re-hashes to learn what was cleared.
type PublishResult struct {
	Slug   string
	Entry  RegistryEntry
	Source []byte
	Hash   string // ContentHash of Source — equal to Entry.Hash, proven below
}

// PublishPreflight is the gate a recipe must clear before it may enter the corpus and be attested.
// It re-derives every claim from the bytes on disk and trusts nothing a caller asserts — the same
// discipline the corpus workflow re-runs server-side, because an attestation over an UNVERIFIED
// recipe is exactly the decorative claim the whole provenance feature exists to avoid. Two things
// must hold, and a failure names which:
//
//   - SEALED — the recipe is in the registry as a recipe, and the committed file still matches the
//     seal: same id (a permanent identity, never edited), same version, same content hash. A
//     drifted or unsealed recipe is refused; an attestation must bind the sealed bytes, not a draft.
//   - GATE-PASSED — the genre gate passes: a summary, the eleven sections in order, neutrality (no
//     leaked paths or product names), and earned scars.
//
// It reads only the recipe's own project (root): the corpus maintainer runs it against the corpus,
// a consumer against their own tree, and each checks the seal its own registry witnesses.
func PublishPreflight(root, slug string) (*PublishResult, error) {
	cfg, err := LoadConfig(root)
	if err != nil {
		return nil, err
	}
	name := slug + ".md"
	src, err := os.ReadFile(filepath.Join(root, cfg.Home, name))
	if err != nil {
		return nil, fmt.Errorf("no recipe %q in the recipes home — publish an authored recipe by its slug, not a path or a name that is not here", slug)
	}

	reg, err := LoadRegistry(root)
	if err != nil {
		return nil, err
	}
	entry, ok := reg.Recipes[slug]
	if !ok || entry.Kind != recipekit.KindRecipe {
		return nil, fmt.Errorf("recipe %q is not sealed — an attestation may only cover a sealed recipe; seal it first (`sporo seal %s`)", slug, slug)
	}

	// Seal coherence: the committed bytes must still be exactly what the seal witnesses. Mirrors
	// verifySealCoherence, kept as its own read here so publish depends on the check, not the sweep.
	switch {
	case entry.ID != "" && fmValue(src, "id") != entry.ID:
		return nil, fmt.Errorf("recipe %q frontmatter id %s does not match the sealed id %s — the id is a permanent identity, never edited; restore it before publishing", slug, fmValue(src, "id"), entry.ID)
	case fmValue(src, "version") != entry.Version:
		return nil, fmt.Errorf("recipe %q says version %s but the seal says %s — re-seal (`sporo seal %s`) so the seal witnesses the version the document declares, then publish", slug, fmValue(src, "version"), entry.Version, slug)
	case ContentHash(src) != entry.Hash:
		return nil, fmt.Errorf("recipe %q has drifted from its seal without a version bump (still %s) — a published recipe is attested by its sealed bytes; bump `version:` and re-seal before publishing", slug, entry.Version)
	}

	// Genre gate: the same shape/neutrality/scars gate `sporo lint` runs. An attestation says
	// "these bytes came through the official pipeline"; the pipeline does not carry a recipe that
	// does not pass its own gate.
	if findings := Lint(name, src, cfg.Products); len(findings) > 0 {
		return nil, fmt.Errorf("recipe %q does not pass the genre gate (%d finding(s)) — fix and re-lint before publishing; first: %s", slug, len(findings), findings[0].Msg)
	}

	return &PublishResult{Slug: slug, Entry: entry, Source: src, Hash: ContentHash(src)}, nil
}
