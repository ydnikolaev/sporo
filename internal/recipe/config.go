package recipe

// The config seam — what sporo cannot know about the project it is running in, and therefore
// reads from `.sporo/config.yaml` (or defaults when there is none).
//
// Two corpora, and the difference is the whole design:
//
//   - the OFFICIAL corpus (embedded in the binary) is READ-ONLY here. It is what other
//     projects' builds taught, shipped with the tool like any doctrine.
//   - the PROJECT home (default `.sporo/recipes/`) is where THIS repo's author writes a
//     recipe about THIS repo's build. It is project-owned: sporo never rewrites it, and a
//     future `push` is how one of them graduates into the shared corpus.
//
// Collapsing the two would mean a project's own recipe is either overwritten on the next
// update (if it lands in a managed place) or invisible to the gate (if it lands nowhere the
// gate looks). Both are silent failures.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"

	"sporo.dev/sporo/pkg/recipekit"
)

// Config is the recipe module's view of a project.
type Config struct {
	// Home is where this project's OWN recipes are authored (`<home:recipes>`).
	Home string
	// Homes maps each entity KIND to the corpus home this project authors that kind's instances in.
	// It is the per-kind generalization of Home: the flat `home:` key IS the recipe home (kept so
	// every existing caller is unchanged), and every other kind (seed, …) declares its own home
	// under `homes:`. Read it through HomeFor, never directly — HomeFor owns the recipe-always-Home
	// and absent-is-not-an-error contracts. This is the seam S3 (the seed CLI) and S4 (web) resolve
	// each kind's corpus through, which is why it is designed for N kinds from day one.
	Homes map[string]string
	// Products is the forbidden vocabulary of the neutrality scan: names that mean
	// something only inside one repository. It defaults to the project's own name —
	// the one name most likely to leak into a document written by someone standing in
	// it — and a project with siblings (a fleet, a monorepo's services) lists them.
	Products []string
	// Sources are the records the harvest reads besides git. Declared, not guessed:
	// a source that does not exist is a STATED absence in the harvest output, never a
	// crash and never a silent zero.
	Sources Sources
}

// Sources is the project's record — the places a recipe is distilled FROM. Every field is a
// path relative to the project root, and every one is optional: a project with no decision
// log still gets a harvest, and the harvest says so out loud.
// The json tags are not optional decoration: this struct is nested inside Harvest, which is
// emitted as JSON, and every sibling field there is tagged snake_case. Without matching json
// tags here the nested `sources` object alone would serialize with Go's PascalCase field
// names — one inconsistent island in an otherwise snake_case contract. Tag both dialects.
type Sources struct {
	// Gates is the gate registry — what this project guards, i.e. what a recipe's
	// `## Verification` section must teach the next builder to guard too.
	Gates string `yaml:"gates" json:"gates"`
	// Doctrine is the principles corpus — the payload section's raw material.
	Doctrine string `yaml:"doctrine" json:"doctrine"`
	// Decisions is the ADR / decision log: WHY the build went the way it did. Without it a
	// recipe reconstructs the rationale from commit messages, which is a lossy guess.
	Decisions string `yaml:"decisions" json:"decisions"`
	// Knowledge is the durable knowledge base (a DKB, an architecture corpus) — the ground
	// the capability stands on.
	Knowledge string `yaml:"knowledge" json:"knowledge"`
}

// probes are the locations a source is looked for when the config does not declare it.
// A probe is a CONVENIENCE, not a contract: it fires only when the key is absent, it never
// overrides a declaration, and whatever it finds (or fails to find) is reported. The list is
// short on purpose — an aggressive probe that guesses wrong is worse than a stated absence,
// because the author never learns the harvest read nothing.
var probes = map[string][]string{
	"gates":     {"harness/gates.yaml", ".agents/gates.yaml", "gates.yaml"},
	"doctrine":  {"doctrine", ".mate/doctrine", "docs/doctrine"},
	"decisions": {"docs/decisions", "docs/decisions.md", "docs/adr", "docs/architecture/decisions"},
	"knowledge": {"docs/architecture", ".agents/knowledge", "docs/knowledge"},
}

// DefaultHome is where a project's own recipes are authored when the config declares nothing.
// It sits under `.sporo/` — the per-repo home the tool owns — so a project that never wrote a
// config still has a place the gate looks and the author writes.
const DefaultHome = ".sporo/recipes/"

// projectConfig is `.sporo/config.yaml`. sporo IS the recipe tool, so there is no `recipe:`
// wrapper: the block is flat.
type projectConfig struct {
	Project  string            `yaml:"project"`
	Home     string            `yaml:"home"`
	Homes    map[string]string `yaml:"homes"`
	Products []string          `yaml:"products"`
	Sources  Sources           `yaml:"sources"`
}

// HomeFor returns the corpus home this project authors the given kind's instances in, and whether
// one is declared. The recipe kind always resolves to Home — the flat `home:` key IS the recipe
// home, and no `homes:` entry can override it (back-compat; every existing caller reads Home). Any
// other declared kind resolves to its `homes:` entry; a kind with NO declared home returns
// ("", false) — an absence the caller reports, never a crash (REQ-5): a project that authors
// recipes but not seeds is a stated "no seed corpus". An unknown kind can only enter through
// `.sporo/config.yaml`, and LoadConfig rejects it there (mirroring LoadRegistry), so a well-formed
// Config holds only valid kinds and this stays a total two-outcome lookup.
func (c Config) HomeFor(kind string) (string, bool) {
	home, ok := c.Homes[kind]
	return home, ok
}

// LoadConfig folds a project's `.sporo/config.yaml` over the defaults. A MISSING config is
// fine (a repo that never ran `sporo init` can still be harvested — that is the point of the
// probes); a MALFORMED one is a hard error, because degrading to the defaults would discard
// the project's own product vocabulary and run a neutrality scan that bans nothing.
func LoadConfig(root string) (Config, error) {
	c := Config{Home: DefaultHome, Products: defaultProducts(root)}
	// Seed the recipe home now so even a config-less project (the early return below) answers
	// HomeFor(recipe). The config-present path re-asserts it after resolving `home:`.
	c.Homes = map[string]string{recipekit.KindRecipe: c.Home}

	data, err := os.ReadFile(filepath.Join(root, ".sporo", "config.yaml"))
	if err != nil {
		c.Sources = probe(root, Sources{})
		return c, nil //nolint:nilerr // absent config → documented defaults (malformed is the hard-error path below)
	}
	var cfg projectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return c, fmt.Errorf(".sporo/config.yaml is malformed — fix the YAML (a broken config never degrades to the defaults, or the neutrality scan silently bans nothing): %w", err)
	}
	if strings.TrimSpace(cfg.Home) != "" {
		c.Home = cfg.Home
	}
	// Per-kind homes fold over the recipe home. Each declared key is validated against the closed
	// kind vocabulary and an unknown one is a HARD error, not a silently-unreachable home — a typo'd
	// `homes:` key means a fixable config, never a corpus the walk never reaches (mirrors
	// LoadRegistry's default-then-validate on RegistryEntry.Kind). The recipe home is re-asserted
	// LAST so the flat `home:` owns it: a `homes: {recipe: …}` cannot override the back-compat key.
	for kind, home := range cfg.Homes {
		if !recipekit.ValidKind(kind) {
			return c, fmt.Errorf(".sporo/config.yaml declares a home for unknown kind %q — the kind is a member of a closed vocabulary this binary knows; a newer kind means a newer binary (or a typo'd `homes:` key), not a home the walk silently never reaches", kind)
		}
		c.Homes[kind] = home
	}
	c.Homes[recipekit.KindRecipe] = c.Home
	if cfg.Project != "" {
		c.Products = []string{cfg.Project}
	}
	if len(cfg.Products) > 0 {
		c.Products = cfg.Products
	}
	c.Sources = probe(root, cfg.Sources)
	return c, nil
}

// defaultProducts names the one product most likely to leak into a recipe written by someone
// standing in it: the project itself. It resolves the root ABSOLUTELY first — `filepath.Base(".")`
// is `"."`, and a `.` in the forbidden vocabulary compiles to `\b\.\b`, which matches the
// decimal point in "2.5 days". The gate then reds on every number in every recipe and calls
// it a product name. Its own unit tests could not see this: they pass the vocabulary in
// explicitly, so the defect lived entirely in the default the CLI computes — which is the
// path every real invocation takes.
func defaultProducts(root string) []string {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil
	}
	name := filepath.Base(abs)
	if !hasLetter(name) {
		return nil // a root with no nameable project bans nothing, rather than banning punctuation
	}
	return []string{name}
}

func hasLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

// probe fills the sources the config left undeclared, and only those. It returns paths that
// EXIST; an unfound source stays empty, and the harvest reports the absence rather than
// inventing a zero.
func probe(root string, declared Sources) Sources {
	get := map[string]*string{
		"gates":     &declared.Gates,
		"doctrine":  &declared.Doctrine,
		"decisions": &declared.Decisions,
		"knowledge": &declared.Knowledge,
	}
	for key, field := range get {
		if strings.TrimSpace(*field) != "" {
			continue // declared wins, always — even over a probe that would have found more
		}
		for _, cand := range probes[key] {
			if _, err := os.Stat(filepath.Join(root, cand)); err == nil {
				*field = cand
				break
			}
		}
	}
	return declared
}
