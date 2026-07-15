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
)

// Config is the recipe module's view of a project.
type Config struct {
	// Home is where this project's OWN recipes are authored (`<home:recipes>`).
	Home string
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
type Sources struct {
	// Gates is the gate registry — what this project guards, i.e. what a recipe's
	// `## Verification` section must teach the next builder to guard too.
	Gates string `yaml:"gates"`
	// Doctrine is the principles corpus — the payload section's raw material.
	Doctrine string `yaml:"doctrine"`
	// Decisions is the ADR / decision log: WHY the build went the way it did. Without it a
	// recipe reconstructs the rationale from commit messages, which is a lossy guess.
	Decisions string `yaml:"decisions"`
	// Knowledge is the durable knowledge base (a DKB, an architecture corpus) — the ground
	// the capability stands on.
	Knowledge string `yaml:"knowledge"`
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
	Project  string   `yaml:"project"`
	Home     string   `yaml:"home"`
	Products []string `yaml:"products"`
	Sources  Sources  `yaml:"sources"`
}

// LoadConfig folds a project's `.sporo/config.yaml` over the defaults. A MISSING config is
// fine (a repo that never ran `sporo init` can still be harvested — that is the point of the
// probes); a MALFORMED one is a hard error, because degrading to the defaults would discard
// the project's own product vocabulary and run a neutrality scan that bans nothing.
func LoadConfig(root string) (Config, error) {
	c := Config{Home: DefaultHome, Products: defaultProducts(root)}

	data, err := os.ReadFile(filepath.Join(root, ".sporo", "config.yaml"))
	if err != nil {
		c.Sources = probe(root, Sources{})
		return c, nil
	}
	var cfg projectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return c, fmt.Errorf(".sporo/config.yaml is malformed — fix the YAML (a broken config never degrades to the defaults, or the neutrality scan silently bans nothing): %w", err)
	}
	if strings.TrimSpace(cfg.Home) != "" {
		c.Home = cfg.Home
	}
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
