package recipe

import (
	"fmt"
	"os"
	"path/filepath"

	"sporo.dev/sporo/pkg/recipekit"
)

// PublishResult is the verified subject a passed pre-flight yields: the sealed entity the corpus
// attest step may sign. It carries the bytes and the seal hash so a caller never re-reads or
// re-hashes to learn what was cleared; Entry.Kind names whether it is a recipe or a seed.
type PublishResult struct {
	Slug   string
	Entry  RegistryEntry
	Source []byte
	Hash   string // ContentHash of Source — equal to Entry.Hash, proven below
}

// PublishPreflight is the gate an entity — a recipe OR a seed — must clear before it may enter the
// corpus and be attested. It re-derives every claim from the bytes on disk and trusts nothing a
// caller asserts — the same discipline the corpus workflow re-runs server-side, because an
// attestation over an UNVERIFIED entity is exactly the decorative claim the whole provenance
// feature exists to avoid. It is kind-polymorphic: the sealed kind (from the registry) decides both
// the home the source is read from and the genre gate that must pass. Two things must hold, and a
// failure names which:
//
//   - SEALED — the entity is in the registry, and the committed file (in its kind's home) still
//     matches the seal: same id (a permanent identity, never edited), same version, same content
//     hash. A drifted or unsealed entity is refused; an attestation must bind the sealed bytes.
//   - GATE-PASSED — its genre gate passes (`Lint` for a recipe, `LintSeed` for a seed). An unknown
//     kind is REFUSED, never passed silently — a decorative pass is the hole this closes.
//
// It reads only the entity's own project (root): the corpus maintainer runs it against the corpus,
// a consumer against their own tree, and each checks the seal its own registry witnesses.
func PublishPreflight(root, slug string) (*PublishResult, error) {
	cfg, err := LoadConfig(root)
	if err != nil {
		return nil, err
	}

	// Entry first: the sealed kind is what tells us which home to read and which gate to run. A slug
	// with no registry entry is unsealed — the same refusal for a recipe and a seed.
	reg, err := LoadRegistry(root)
	if err != nil {
		return nil, err
	}
	entry, ok := reg.Recipes[slug]
	if !ok {
		return nil, fmt.Errorf("%q is not sealed — an attestation may only cover a sealed entity; author and seal it first (nothing by that slug is in the registry)", slug)
	}

	home, ok := cfg.HomeFor(entry.Kind)
	if !ok {
		return nil, fmt.Errorf("%q is sealed as kind %q, which this project declares no home for — a newer kind needs a newer binary or config", slug, entry.Kind)
	}
	name := slug + ".md"
	src, err := os.ReadFile(filepath.Join(root, home, name))
	if err != nil {
		return nil, fmt.Errorf("%q is sealed as a %s but its file is missing from %s — restore it before publishing", slug, entry.Kind, home)
	}

	// Seal coherence: the committed bytes must still be exactly what the seal witnesses. Kind-agnostic
	// (all read the entry); mirrors verifySealCoherence, kept as its own read here so publish depends
	// on the check, not the sweep.
	switch {
	case entry.ID != "" && fmValue(src, "id") != entry.ID:
		return nil, fmt.Errorf("%s %q frontmatter id %s does not match the sealed id %s — the id is a permanent identity, never edited; restore it before publishing", entry.Kind, slug, fmValue(src, "id"), entry.ID)
	case fmValue(src, "version") != entry.Version:
		return nil, fmt.Errorf("%s %q says version %s but the seal says %s — re-seal (`%s`) so the seal witnesses the version the document declares, then publish", entry.Kind, slug, fmValue(src, "version"), entry.Version, resealHint(entry.Kind, slug))
	case ContentHash(src) != entry.Hash:
		return nil, fmt.Errorf("%s %q has drifted from its seal without a version bump (still %s) — a published entity is attested by its sealed bytes; bump `version:` and re-seal before publishing", entry.Kind, slug, entry.Version)
	}

	// Genre gate, dispatched by kind — the same gate `sporo lint` / `sporo seed lint` runs. An
	// attestation says "these bytes came through the official pipeline"; the pipeline does not carry
	// an entity that fails its own gate, and an unknown kind is refused, not waved through.
	var findings []Finding
	switch entry.Kind {
	case recipekit.KindRecipe:
		findings = Lint(name, src, cfg.Products)
	case recipekit.KindSeed:
		findings = LintSeed(name, src, cfg.Products)
	default:
		return nil, fmt.Errorf("publish cannot gate kind %q — a kind this binary does not know must not pass the gate by default; upgrade sporo", entry.Kind)
	}
	if len(findings) > 0 {
		return nil, fmt.Errorf("%s %q does not pass the genre gate (%d finding(s)) — fix and re-lint before publishing; first: %s", entry.Kind, slug, len(findings), findings[0].Msg)
	}

	return &PublishResult{Slug: slug, Entry: entry, Source: src, Hash: ContentHash(src)}, nil
}

// resealHint names the seal command for a kind — `sporo seal` for a recipe, `sporo seed seal` for a
// seed — so a coherence-failure message tells the author the exact command to run.
func resealHint(kind, slug string) string {
	if kind == recipekit.KindSeed {
		return "sporo seed seal " + slug
	}
	return "sporo seal " + slug
}
