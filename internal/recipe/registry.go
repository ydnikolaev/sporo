package recipe

// The registry — `.sporo/registry.yaml` — is what makes a recipe's version REAL.
//
// The frontmatter carries `version:` because the exported file is the only thing the reader
// has; but a version that nothing checks is a comment. The registry is the check's other
// half: `sporo seal` records (version, content hash, provenance) for a recipe the author
// declares done, and from then on the pair is guarded — content that changes under a sealed
// version is a gate failure, not a mystery. A marketplace inherits its integrity story from
// exactly this: a recipe never silently mutates.
//
// Sealing is DELIBERATE, not automatic. A draft the author is still shaping has no entry and
// no obligations; the moment it is handed to anyone, it gets sealed, and every edit after
// that must announce itself as a new version. The loop closes the circle: report-backs cite
// the version they built, new scars produce the next one, `seal` stamps it.
//
// The registry also records the MANAGED files `sporo init` writes into a repo (the authoring
// skill, the seeded config). That is how `update` finds them again, skips the unchanged, and
// reports — never clobbers — one the user edited. Same file, two ledgers, one discipline:
// what sporo wrote, sporo may rewrite; what the author wrote, sporo may only guard.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"sporo.dev/sporo/pkg/recipekit"
)

// sealNow stamps the moment a seal is recorded. It is a var so tests can pin it; production reads
// the wall clock in UTC (a ledger timestamp belongs in one zone, not the sealer's local one).
var sealNow = func() string { return time.Now().UTC().Format(time.RFC3339) }

// RegistrySchema names the current on-disk shape, so a future binary can migrate an old
// registry instead of misreading it. Schema 2 adds `kind` to each entry; it is
// read-compatible with schema 1 — a schema-1 ledger (no kind fields) loads unchanged,
// every entry defaulting to `recipe` (see LoadRegistry).
const RegistrySchema = 2

// RegistryEntry is one sealed recipe: its permanent identity, the version the author
// declared, the content that version names, and whose build it is.
type RegistryEntry struct {
	// ID mirrors the recipe's frontmatter `id` — the ULID minted at `sporo new`. Unlike the
	// version (which climbs) and the hash (which changes with every edit), the id is FIXED:
	// it is the recipe's permanent identity across renames and releases, the key a marketplace
	// and a report-back thread hang on. The registry records it so the ledger — not just the
	// file — witnesses the identity. `omitempty` keeps registries sealed before ids existed
	// loadable; the next seal backfills them.
	ID string `yaml:"id,omitempty"`
	// Kind names the entity genre this seal is — a member of recipekit's closed vocabulary
	// (`recipe`, `seed`). `omitempty` keeps a schema-1 ledger (sealed before kinds existed)
	// loadable: a missing kind reads as `recipe` in LoadRegistry, so no re-seal or migration
	// step is forced. It is registry metadata, never part of the content hash.
	Kind string `yaml:"kind,omitempty"`
	// Version mirrors the recipe's own frontmatter at seal time. The two diverging is a
	// finding, not a preference — each is the other's witness.
	Version string `yaml:"version"`
	// Hash is `sha256:<hex>` of the recipe SOURCE file at seal time. The exported file
	// differs (banner stripped, protocol appended), so the source is the thing sealed.
	Hash string `yaml:"hash"`
	// Provenance says whose build this is: `local` (authored here), `community` (pulled
	// from the public corpus), or `team:<name>` (a private workspace). It is the extension
	// of the official/project split that Origin draws for the two corpora.
	Provenance string `yaml:"provenance"`
	// SealedAt is when this seal was recorded (RFC3339, UTC) — a machine-stamped fact, unlike
	// the frontmatter's author-typed `verified.date`. Set on the first seal and on every re-seal
	// after a version bump; preserved on an idempotent re-seal. `omitempty` keeps registries
	// sealed before the field existed loadable — the next seal backfills them.
	SealedAt string `yaml:"sealed_at,omitempty"`
	// ExactContracts digests the exact-bound fence bodies at seal time (empty when the
	// recipe has none). It is what lets seal enforce the fleet rule: an exact contract is
	// somebody else's parser, and a change to it under a minor bump ships a break wearing a
	// compatible version number.
	ExactContracts string `yaml:"exact_contracts,omitempty"`
}

// Managed is one file `sporo init`/`update` wrote into this repository. Hash is the content
// as WRITTEN — if the file on disk no longer matches it, the user edited it, and update
// reports instead of rewriting.
type Managed struct {
	Path string `yaml:"path"`
	Hash string `yaml:"hash"`
	// Binary is the CLI version that wrote the file — the provenance stamp's machine half.
	Binary string `yaml:"binary"`
}

// ReviewSummary is one `review verify` outcome: how many verdicts were tallied, their mean,
// and the aggregate call. It lives beside the seal rather than inside it — RegistryEntry
// stays a comparable value (the seal either matches or it does not), while reviews accumulate.
type ReviewSummary struct {
	Date     string  `yaml:"date"`
	Version  string  `yaml:"version"`
	Verdicts int     `yaml:"verdicts"`
	Mean     float64 `yaml:"mean"`
	Verdict  string  `yaml:"verdict"`
}

// Registry is `.sporo/registry.yaml`, whole.
type Registry struct {
	Schema  int                        `yaml:"schema"`
	Recipes map[string]RegistryEntry   `yaml:"recipes,omitempty"`
	Reviews map[string][]ReviewSummary `yaml:"reviews,omitempty"`
	Managed []Managed                  `yaml:"managed,omitempty"`
	// Adopted is the READER-side ledger: the handed-over recipes this repository built
	// from, each with the version and source `sporo pull` re-checks. The author's ledger
	// (Recipes) records what this repo promises out; this one records what it took in.
	Adopted map[string]AdoptedEntry `yaml:"adopted,omitempty"`
}

func registryPath(root string) string {
	return filepath.Join(root, ".sporo", "registry.yaml")
}

// LoadRegistry reads the registry, or returns an empty one when the project has none — a
// repo that never sealed anything is not an error. A MALFORMED registry is a hard error for
// the same reason a malformed config is: degrading to "empty" would un-seal every recipe in
// the project without a word.
func LoadRegistry(root string) (Registry, error) {
	r := Registry{Schema: RegistrySchema, Recipes: map[string]RegistryEntry{}}
	data, err := os.ReadFile(registryPath(root))
	if err != nil {
		return r, nil //nolint:nilerr // absent registry → empty (malformed is the hard error below)
	}
	if err := yaml.Unmarshal(data, &r); err != nil {
		return r, fmt.Errorf(".sporo/registry.yaml is malformed — fix the YAML (a broken registry never degrades to an empty one, or every sealed recipe in this project is silently unsealed): %w", err)
	}
	if r.Recipes == nil {
		r.Recipes = map[string]RegistryEntry{}
	}
	// Default THEN validate: an absent kind (a schema-1 entry, or one written before the field
	// existed) reads as `recipe` — the read-compatibility that keeps an old ledger loadable
	// without a re-seal. Only after defaulting is the kind checked against the closed vocabulary,
	// so a genuinely unknown kind is a hard error (a foreign genre this binary cannot honour is
	// the same danger as a malformed registry — better refused than silently mis-sealed).
	for slug, entry := range r.Recipes {
		if entry.Kind == "" {
			entry.Kind = recipekit.KindRecipe
			r.Recipes[slug] = entry
		}
		if !recipekit.ValidKind(entry.Kind) {
			return r, fmt.Errorf(".sporo/registry.yaml records recipe %q with unknown kind %q — the kind is a member of a closed vocabulary this binary knows; a newer kind means a newer binary (or a corrupt ledger), not a degrade to empty", slug, entry.Kind)
		}
	}
	return r, nil
}

// Save writes the registry back. The parent `.sporo/` is created if missing: sealing the
// first recipe in a repo that never ran `init` should work, not scold.
func (r Registry) Save(root string) error {
	r.Schema = RegistrySchema
	b, err := yaml.Marshal(r)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(root, ".sporo"), 0o755); err != nil {
		return err
	}
	return os.WriteFile(registryPath(root), b, 0o644)
}

// Seal records a recipe's (version, hash, provenance) in the registry — the flat recipe path,
// unchanged. It is SealKind bound to the recipe kind, so every existing caller keeps its
// signature and every rule SealKind spells out applies to a recipe exactly as before.
func Seal(root string, cfg Config, slug string) (RegistryEntry, error) {
	return SealKind(root, cfg, recipekit.KindRecipe, slug)
}

// SealKind records one entity's (version, hash, provenance, KIND) in the registry, resolving
// the source home per kind. The rules are the integrity story, so they are spelled out:
//
//   - an entity with no entry seals at whatever version its frontmatter declares;
//   - unchanged content re-seals as a no-op (idempotent, so a script can call it);
//   - CHANGED content under an UNCHANGED version is refused — that edit is exactly the
//     silent mutation the registry exists to catch; bump `version:` first;
//   - changed content under a new version seals the new pair;
//   - a slug already sealed under a DIFFERENT kind is refused (the ledger is slug-keyed, so
//     one slug is one kind for its life — see the collision guard below).
//
// Provenance defaults to `local` for a new entry and is PRESERVED on a re-seal: sealing a
// new version of a community recipe does not quietly claim it as yours. Kind is stamped on the
// first seal and preserved on every re-seal — the collision guard guarantees the loaded entry
// already carries this kind, so the re-seal branches keep it the way they keep provenance.
func SealKind(root string, cfg Config, kind, slug string) (RegistryEntry, error) {
	home, ok := cfg.HomeFor(kind)
	if !ok {
		// HomeFor's recipe-always-Home contract holds even for a hand-built Config whose Homes map
		// was never seeded (LoadConfig seeds it; a direct Config{Home:…} does not), so the flat
		// recipe path resolves to cfg.Home byte-for-byte as before. Any OTHER kind with no declared
		// home is a stated absence, never a crash (REQ-5): a project that authors recipes but not
		// seeds has no seed corpus to seal from.
		if kind != recipekit.KindRecipe {
			return RegistryEntry{}, fmt.Errorf("this project declares no home for kind %q — it authors no %s corpus to seal from; declare `homes: {%s: …}` in .sporo/config.yaml first", kind, kind, kind)
		}
		home = cfg.Home
	}
	src, err := os.ReadFile(filepath.Join(root, home, slug+".md"))
	if err != nil {
		return RegistryEntry{}, fmt.Errorf("no %s %q in this project's home (%s) — seal guards what exists: %w", kind, slug, home, err)
	}
	if IsDraft(src) {
		return RegistryEntry{}, fmt.Errorf("%s %q is still a draft — a draft has no version to promise; finish it, remove `draft: true`, then seal", kind, slug)
	}
	version := fmValue(src, "version")
	if version == "" {
		return RegistryEntry{}, fmt.Errorf("%s %q has no `version:` in its frontmatter — the seal records a version the document itself declares, so declare one (and `sporo lint` requires it)", kind, slug)
	}
	id := fmValue(src, "id")
	hash := ContentHash(src)

	reg, err := LoadRegistry(root)
	if err != nil {
		return RegistryEntry{}, err
	}
	digest := exactContractsDigest(src)
	entry, sealed := reg.Recipes[slug]
	// The registry is ONE slug-keyed map across kinds, so a slug is one kind for the life of the
	// ledger. Sealing it as a different kind would silently overwrite the other genre's entry, and
	// the coherence sweep would then hunt it in the wrong home — refuse it, naming both kinds. This
	// is the smaller, back-compatible choice over a schema-3 kind-qualified key.
	if sealed && entry.Kind != kind {
		return RegistryEntry{}, fmt.Errorf("%q is already sealed as a %s — refusing to seal it as a %s; a slug is one kind across the registry (it is slug-keyed), so give the %s a different slug", slug, entry.Kind, kind, kind)
	}
	// The id is a PERMANENT identity — it may be recorded and it may be backfilled onto an old
	// seal that predates ids, but it may never CHANGE. A changed id is not a new version of the
	// same entity; it is a different one wearing this slug, and letting it re-seal in place would
	// silently rewrite a marketplace permalink.
	if sealed && entry.ID != "" && id != "" && id != entry.ID {
		return RegistryEntry{}, fmt.Errorf("%s %q was sealed with id %s but its frontmatter now says %s — the id is a permanent identity, minted once and never edited; restore it (a genuinely new %s gets a new slug, not a rewritten id)", kind, slug, entry.ID, id, kind)
	}
	now := sealNow()
	switch {
	case !sealed:
		// The first seal stamps the kind it was handed. The re-seal and backfill branches below
		// mutate the loaded entry, which already carries this same kind (the collision guard above
		// guarantees it), so kind is preserved there the way provenance is.
		entry = RegistryEntry{ID: id, Kind: kind, Version: version, Hash: hash, Provenance: "local", SealedAt: now, ExactContracts: digest}
	case entry.Hash == hash && entry.Version == version && entry.ID == id:
		// Already sealed exactly so — idempotent by design. Backfill only a MISSING sealed_at (a
		// seal made before the field existed): recording when is the ledger catching up, not a
		// mutation, so it needs no version bump.
		if entry.SealedAt != "" {
			return entry, nil
		}
		entry.SealedAt = now
	case entry.Hash == hash && entry.Version == version:
		// Content and version match; only the id differs — an old seal being backfilled. Record
		// it without demanding a version bump: adding the id to the registry is not a mutation
		// of the entity, it is the ledger catching up to a field the file already carries.
		entry.ID = id
		if entry.SealedAt == "" {
			entry.SealedAt = now
		}
	case entry.Version == version:
		return RegistryEntry{}, fmt.Errorf("%s %q changed since it was sealed at %s, but `version:` still says %s — a sealed %s never silently mutates; bump the version, then seal", kind, slug, entry.Version, version, kind)
	default:
		// The fleet rule: an exact-bound contract is somebody else's parser. Changing one
		// under anything less than a major bump ships a break wearing a compatible version
		// number — the consumer upgrades "safely" and their feed dies.
		if entry.ExactContracts != "" && digest != entry.ExactContracts && semverMajor(version) <= semverMajor(entry.Version) {
			return RegistryEntry{}, fmt.Errorf("%s %q changes an exact-bound contract — every consumer in the fleet parses that shape, so this is a MAJOR version (%s → at least %d.0.0), not %s", kind, slug, entry.Version, semverMajor(entry.Version)+1, version)
		}
		// A re-seal after a version bump is a new seal event — restamp when.
		entry.ID, entry.Version, entry.Hash, entry.ExactContracts, entry.SealedAt = id, version, hash, digest, now
	}
	reg.Recipes[slug] = entry
	if err := reg.Save(root); err != nil {
		return RegistryEntry{}, err
	}
	return entry, nil
}

// verifySealCoherence is the map-sweep half of the gate, scoped to one KIND and its home:
// every sealed entry OF THAT KIND must still say what its seal says. It is the shared mechanism
// both the recipe gate (VerifyRegistry) and the seed lint path call — the same checks scoped per
// kind, never one global sweep, so a sealed seed is never hunted in the recipe home and a sealed
// recipe is never hunted in the seed home (INV-1). Findings come back in lint's own currency.
//
// The genre nouns are templated off `kind` — `recipe`→"recipes home"/"sporo seal", every other
// kind→"<kind>s home"/"sporo <kind> seal" — so the recipe path reproduces its wording byte-for-byte
// while a seed reads coherently through its own namespace.
func verifySealCoherence(root, home string, reg Registry, kind string) []Finding {
	// The re-seal verb for this kind: recipes use the flat `sporo seal`, every other kind its
	// namespaced `sporo <kind> seal` (the seed CLI's `seal` subcommand).
	seal := "sporo seal"
	if kind != recipekit.KindRecipe {
		seal = "sporo " + kind + " seal"
	}
	var out []Finding
	for slug, entry := range reg.Recipes {
		if entry.Kind != kind {
			continue // another genre's seal — scoped out, so a flat recipe verb never sees a seed
		}
		name := slug + ".md"
		src, err := os.ReadFile(filepath.Join(root, home, name))
		if err != nil {
			out = append(out, Finding{File: name, Line: 0, Msg: fmt.Sprintf("sealed at %s but missing from the %ss home — a sealed %s that vanished is a broken promise, not a cleanup; unseal it deliberately (remove its registry entry) or restore it", entry.Version, kind, kind)})
			continue
		}
		version := fmValue(src, "version")
		switch {
		case entry.ID != "" && fmValue(src, "id") != entry.ID:
			out = append(out, Finding{File: name, Line: 0, Msg: fmt.Sprintf("frontmatter id %s does not match the sealed id %s — the id is a permanent identity, never edited; restore it (a genuinely different %s gets its own slug, not a rewritten id)", fmValue(src, "id"), entry.ID, kind)})
		case version != entry.Version:
			out = append(out, Finding{File: name, Line: 0, Msg: fmt.Sprintf("frontmatter says version %s but the seal says %s — re-seal (`%s %s`) so the registry witnesses the version the document declares", version, entry.Version, seal, slug)})
		case ContentHash(src) != entry.Hash:
			out = append(out, Finding{File: name, Line: 0, Msg: fmt.Sprintf("content drifted from its seal without a version bump (still %s) — a sealed %s never silently mutates; bump `version:`, then `%s %s`", entry.Version, kind, seal, slug)})
		}
	}
	return out
}

// VerifyRegistry is the coherence half of the gate: every sealed recipe's file must still
// say what the seal says. It returns findings in lint's own currency so the two run as one
// gate. An UNSEALED recipe is legal — a draft has no obligations yet — so absence from the
// registry is never a finding. It is kind-scoped to RECIPES: a sealed seed lives in the same
// slug-keyed ledger but is swept by the seed lint path against the seed home, never here.
func VerifyRegistry(root string, cfg Config) ([]Finding, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return nil, err
	}
	out := verifySealCoherence(root, cfg.Home, reg, recipekit.KindRecipe)
	// The other direction: every FINISHED recipe in the project's own home must be sealed. A
	// non-draft recipe declares itself done, and done means it promises a version; an unsealed
	// one is published in intent but unwitnessed by the registry, so "all recipes are sealed"
	// cannot be claimed. Drafts are exempt — they have no version to promise yet. (This runs
	// only in the own-home gate, alongside the coherence check above, never against a borrowed
	// corpus an explicit `sporo lint <dir>` might point at.)
	ents, err := os.ReadDir(filepath.Join(root, cfg.Home))
	if err != nil {
		return out, nil //nolint:nilerr // no home to sweep — lint reports a missing corpus on its own
	}
	for _, e := range ents {
		if !IsRecipe(e.Name(), e.IsDir()) {
			continue
		}
		slug := strings.TrimSuffix(e.Name(), ".md")
		if _, sealed := reg.Recipes[slug]; sealed {
			continue
		}
		src, err := os.ReadFile(filepath.Join(root, cfg.Home, e.Name()))
		if err != nil || IsDraft(src) {
			continue
		}
		out = append(out, Finding{File: e.Name(), Line: 0, Msg: fmt.Sprintf("finished but not sealed — a non-draft recipe promises a version; `sporo seal %s` to record it, or mark it `draft: true` while it is still in flux", slug)})
	}
	return out, nil
}

// exactContractsDigest forwards to recipekit — the digest is a pure function of the source,
// and the registry shares it with a future server via the extracted package.
func exactContractsDigest(src []byte) string { return recipekit.ExactContractsDigest(src) }

// semverMajor reads the MAJOR of a semver triple; a malformed version reads as 0.
func semverMajor(v string) int { return recipekit.SemverMajor(v) }

// fmValue extracts one scalar from a recipe's frontmatter, tolerant of quotes.
func fmValue(src []byte, key string) string { return recipekit.FrontmatterValue(src, key) }

// ContentHash is the registry's currency: `sha256:<hex>` over exact bytes. Exported because
// the install layer prices its managed files in the same currency — one hash discipline,
// not two.
func ContentHash(b []byte) string { return recipekit.ContentHash(b) }
