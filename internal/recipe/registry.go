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
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// sealNow stamps the moment a seal is recorded. It is a var so tests can pin it; production reads
// the wall clock in UTC (a ledger timestamp belongs in one zone, not the sealer's local one).
var sealNow = func() string { return time.Now().UTC().Format(time.RFC3339) }

// RegistrySchema names the current on-disk shape, so a future binary can migrate an old
// registry instead of misreading it.
const RegistrySchema = 1

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
		return r, nil
	}
	if err := yaml.Unmarshal(data, &r); err != nil {
		return r, fmt.Errorf(".sporo/registry.yaml is malformed — fix the YAML (a broken registry never degrades to an empty one, or every sealed recipe in this project is silently unsealed): %w", err)
	}
	if r.Recipes == nil {
		r.Recipes = map[string]RegistryEntry{}
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

// Seal records a recipe's (version, hash, provenance) in the registry. The rules are the
// integrity story, so they are spelled out:
//
//   - a recipe with no entry seals at whatever version its frontmatter declares;
//   - unchanged content re-seals as a no-op (idempotent, so a script can call it);
//   - CHANGED content under an UNCHANGED version is refused — that edit is exactly the
//     silent mutation the registry exists to catch; bump `version:` first;
//   - changed content under a new version seals the new pair.
//
// Provenance defaults to `local` for a new entry and is PRESERVED on a re-seal: sealing a
// new version of a community recipe does not quietly claim it as yours.
func Seal(root string, cfg Config, slug string) (RegistryEntry, error) {
	src, err := os.ReadFile(filepath.Join(root, cfg.Home, slug+".md"))
	if err != nil {
		return RegistryEntry{}, fmt.Errorf("no recipe %q in this project's home (%s) — seal guards what exists: %w", slug, cfg.Home, err)
	}
	if IsDraft(src) {
		return RegistryEntry{}, fmt.Errorf("recipe %q is still a draft — a draft has no version to promise; finish it, remove `draft: true`, then seal", slug)
	}
	version := fmValue(src, "version")
	if version == "" {
		return RegistryEntry{}, fmt.Errorf("recipe %q has no `version:` in its frontmatter — the seal records a version the document itself declares, so declare one (and `sporo lint` requires it)", slug)
	}
	id := fmValue(src, "id")
	hash := ContentHash(src)

	reg, err := LoadRegistry(root)
	if err != nil {
		return RegistryEntry{}, err
	}
	digest := exactContractsDigest(src)
	entry, sealed := reg.Recipes[slug]
	// The id is a PERMANENT identity — it may be recorded and it may be backfilled onto an old
	// seal that predates ids, but it may never CHANGE. A changed id is not a new version of the
	// same recipe; it is a different recipe wearing this one's slug, and letting it re-seal in
	// place would silently rewrite a marketplace permalink.
	if sealed && entry.ID != "" && id != "" && id != entry.ID {
		return RegistryEntry{}, fmt.Errorf("recipe %q was sealed with id %s but its frontmatter now says %s — the id is a permanent identity, minted once and never edited; restore it (a genuinely new recipe gets a new slug, not a rewritten id)", slug, entry.ID, id)
	}
	now := sealNow()
	switch {
	case !sealed:
		entry = RegistryEntry{ID: id, Version: version, Hash: hash, Provenance: "local", SealedAt: now, ExactContracts: digest}
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
		// of the recipe, it is the ledger catching up to a field the file already carries.
		entry.ID = id
		if entry.SealedAt == "" {
			entry.SealedAt = now
		}
	case entry.Version == version:
		return RegistryEntry{}, fmt.Errorf("recipe %q changed since it was sealed at %s, but `version:` still says %s — a sealed recipe never silently mutates; bump the version, then seal", slug, entry.Version, version)
	default:
		// The fleet rule: an exact-bound contract is somebody else's parser. Changing one
		// under anything less than a major bump ships a break wearing a compatible version
		// number — the consumer upgrades "safely" and their feed dies.
		if entry.ExactContracts != "" && digest != entry.ExactContracts && semverMajor(version) <= semverMajor(entry.Version) {
			return RegistryEntry{}, fmt.Errorf("recipe %q changes an exact-bound contract — every consumer in the fleet parses that shape, so this is a MAJOR version (%s → at least %d.0.0), not %s", slug, entry.Version, semverMajor(entry.Version)+1, version)
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

// VerifyRegistry is the coherence half of the gate: every sealed recipe's file must still
// say what the seal says. It returns findings in lint's own currency so the two run as one
// gate. An UNSEALED recipe is legal — a draft has no obligations yet — so absence from the
// registry is never a finding.
func VerifyRegistry(root string, cfg Config) ([]Finding, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return nil, err
	}
	var out []Finding
	for slug, entry := range reg.Recipes {
		name := slug + ".md"
		src, err := os.ReadFile(filepath.Join(root, cfg.Home, name))
		if err != nil {
			out = append(out, Finding{name, 0, fmt.Sprintf("sealed at %s but missing from the recipes home — a sealed recipe that vanished is a broken promise, not a cleanup; unseal it deliberately (remove its registry entry) or restore it", entry.Version)})
			continue
		}
		version := fmValue(src, "version")
		switch {
		case entry.ID != "" && fmValue(src, "id") != entry.ID:
			out = append(out, Finding{name, 0, fmt.Sprintf("frontmatter id %s does not match the sealed id %s — the id is a permanent identity, never edited; restore it (a genuinely different recipe gets its own slug, not a rewritten id)", fmValue(src, "id"), entry.ID)})
		case version != entry.Version:
			out = append(out, Finding{name, 0, fmt.Sprintf("frontmatter says version %s but the seal says %s — re-seal (`sporo seal %s`) so the registry witnesses the version the document declares", version, entry.Version, slug)})
		case ContentHash(src) != entry.Hash:
			out = append(out, Finding{name, 0, fmt.Sprintf("content drifted from its seal without a version bump (still %s) — a sealed recipe never silently mutates; bump `version:`, then `sporo seal %s`", entry.Version, slug)})
		}
	}
	// The other direction: every FINISHED recipe in the project's own home must be sealed. A
	// non-draft recipe declares itself done, and done means it promises a version; an unsealed
	// one is published in intent but unwitnessed by the registry, so "all recipes are sealed"
	// cannot be claimed. Drafts are exempt — they have no version to promise yet. (This runs
	// only in the own-home gate, alongside the coherence check above, never against a borrowed
	// corpus an explicit `sporo lint <dir>` might point at.)
	ents, err := os.ReadDir(filepath.Join(root, cfg.Home))
	if err != nil {
		return out, nil // no home to sweep — lint reports a missing corpus on its own
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
		out = append(out, Finding{e.Name(), 0, fmt.Sprintf("finished but not sealed — a non-draft recipe promises a version; `sporo seal %s` to record it, or mark it `draft: true` while it is still in flux", slug)})
	}
	return out, nil
}

// exactContractsDigest hashes the bodies of the exact-bound fences in the contracts
// section, in order. Only the fence CONTENTS count: prose around a shape can be reworded
// freely, field names and structure cannot. An empty digest means the recipe promises
// nothing exact, and the seal imposes nothing.
func exactContractsDigest(src []byte) string {
	con := sectionBody(strings.Split(string(src), "\n"), "## The contracts")
	var buf strings.Builder
	exact, inFence, capture := false, false, false
	for _, l := range con {
		switch {
		case reFence.MatchString(l):
			if !inFence {
				inFence, capture = true, exact
			} else {
				inFence, capture = false, false
			}
			exact = false
		case !inFence && reBinding.MatchString(l):
			exact = strings.Contains(l, "**Binding: exact**")
		default:
			if capture {
				buf.WriteString(l)
				buf.WriteByte('\n')
			}
		}
	}
	if buf.Len() == 0 {
		return ""
	}
	return ContentHash([]byte(buf.String()))
}

// semverMajor reads the MAJOR of a semver triple; a malformed version reads as 0, and the
// lint gate (which requires a real triple) is where malformedness is reported.
func semverMajor(v string) int {
	n := 0
	for _, r := range v {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + int(r-'0')
	}
	return n
}

// fmValue extracts one scalar from a recipe's frontmatter, tolerant of quotes. It reads the
// first `---` pair — scanning from line 0, not line 1, because it serves TWO document
// shapes: a source file (banner on line 0, frontmatter after) and an EXPORTED file (banner
// stripped, frontmatter IS line 0). A scan that assumed the banner missed every export's
// frontmatter entirely — `adopt` found that the hard way. Safe for sources too: a banner
// line is an HTML comment and never trims to `---`.
func fmValue(src []byte, key string) string {
	lines := strings.Split(string(src), "\n")
	start, end := -1, -1
	for i := 0; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			if start < 0 {
				start = i
			} else {
				end = i
				break
			}
		}
	}
	if start < 0 || end < 0 {
		return ""
	}
	line := keyLine(lines[start+1:end], key)
	if line == "" {
		return ""
	}
	v := strings.TrimSpace(strings.TrimPrefix(line, key+":"))
	return strings.Trim(v, `"'`)
}

// ContentHash is the registry's currency: `sha256:<hex>` over exact bytes. Exported because
// the install layer prices its managed files in the same currency — one hash discipline,
// not two.
func ContentHash(b []byte) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(b))
}
