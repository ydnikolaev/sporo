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

	"gopkg.in/yaml.v3"
)

// RegistrySchema names the current on-disk shape, so a future binary can migrate an old
// registry instead of misreading it.
const RegistrySchema = 1

// RegistryEntry is one sealed recipe: the version the author declared, the content that
// version names, and whose build it is.
type RegistryEntry struct {
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
	version := fmValue(src, "version")
	if version == "" {
		return RegistryEntry{}, fmt.Errorf("recipe %q has no `version:` in its frontmatter — the seal records a version the document itself declares, so declare one (and `sporo lint` requires it)", slug)
	}
	hash := ContentHash(src)

	reg, err := LoadRegistry(root)
	if err != nil {
		return RegistryEntry{}, err
	}
	entry, sealed := reg.Recipes[slug]
	switch {
	case !sealed:
		entry = RegistryEntry{Version: version, Hash: hash, Provenance: "local"}
	case entry.Hash == hash && entry.Version == version:
		return entry, nil // already sealed exactly so — idempotent by design
	case entry.Version == version:
		return RegistryEntry{}, fmt.Errorf("recipe %q changed since it was sealed at %s, but `version:` still says %s — a sealed recipe never silently mutates; bump the version, then seal", slug, entry.Version, version)
	default:
		entry.Version, entry.Hash = version, hash
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
		case version != entry.Version:
			out = append(out, Finding{name, 0, fmt.Sprintf("frontmatter says version %s but the seal says %s — re-seal (`sporo seal %s`) so the registry witnesses the version the document declares", version, entry.Version, slug)})
		case ContentHash(src) != entry.Hash:
			out = append(out, Finding{name, 0, fmt.Sprintf("content drifted from its seal without a version bump (still %s) — a sealed recipe never silently mutates; bump `version:`, then `sporo seal %s`", entry.Version, slug)})
		}
	}
	return out, nil
}

// fmValue extracts one scalar from a recipe's frontmatter, tolerant of quotes. It reads the
// same window Lint reads (the first `---` pair after the banner), so the two never disagree
// about where the frontmatter is.
func fmValue(src []byte, key string) string {
	lines := strings.Split(string(src), "\n")
	start, end := -1, -1
	for i := 1; i < len(lines); i++ {
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
