package recipe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sporo.dev/sporo/pkg/recipekit"
)

// The defect this file exists for, and how it hid.
//
// The neutrality scan bans PRODUCT names. The CLI computes that vocabulary from the project
// root, which defaults to ".". `filepath.Base(".")` is `"."` — so the forbidden word was a
// full stop, compiled to `\b\.\b`, which matches the decimal point in "2.5 days". Every
// recipe carrying any number was reported as naming a product of the fleet.
//
// The gate's own unit tests were all green, and they were green for a reason worth keeping in
// mind: they pass the vocabulary in EXPLICITLY. The bug lived entirely in the default the CLI
// derives — the path every real invocation takes and the only one no test covered. It was
// caught by running the gate on the real corpus. So the teeth now bite the DEFAULT, not the
// injected list.
func TestTheDefaultProductVocabularyIsNeverPunctuation(t *testing.T) {
	for _, root := range []string{".", "", "./", "../recipe"} {
		cfg, err := LoadConfig(root)
		if err != nil {
			t.Fatalf("root %q: %v", root, err)
		}
		for _, p := range cfg.Products {
			if !hasLetter(p) {
				t.Fatalf("root %q derived %q as a product name — a vocabulary of punctuation reds on every number in every recipe", root, p)
			}
		}
		// And prove it end-to-end through the public gate: the default vocabulary must not fire
		// on a decimal. (The neutrality scan itself now lives in pkg/recipekit; Lint is the seam
		// this package still owns, so the check rides through it.)
		body := strings.Replace(conformant, "Derive, never restate.",
			"An estimate of 2.5 days is a guess wearing a suit.", 1)
		if f := Lint("x.md", []byte(body), cfg.Products); len(f) != 0 {
			t.Fatalf("root %q: the default vocabulary reds on a plain number: %v", root, f)
		}
	}
}

func TestTheProjectsOwnNameIsBannedByDefault(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "payments-api")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	// A recipe written by someone standing in the repo leaks the repo's own name first —
	// so with no config at all, that is the one name the gate must already know. Routed
	// through the public Lint (the neutrality scan lives in pkg/recipekit now).
	body := strings.Replace(conformant, "Derive, never restate.",
		"Wire it the way payments-api does.", 1)
	red := false
	for _, x := range Lint("x.md", []byte(body), cfg.Products) {
		if strings.Contains(x.Msg, "product") {
			red = true
		}
	}
	if !red {
		t.Fatal("a project with no config must still ban its OWN name — it is the one most likely to leak")
	}
}

// A source the project does not have is a STATED absence. The distinction is the whole
// difference between "this project has no decision log" and "the harvest failed to read the
// decision log", and in an empty array those look identical.
func TestAMissingSourceIsStatedNotGuessed(t *testing.T) {
	dir := t.TempDir()
	cfg, err := LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Sources.Gates != "" || cfg.Sources.Decisions != "" {
		t.Fatalf("an empty repo has no records to probe; got %+v", cfg.Sources)
	}
	h := &Harvest{}
	if _, err := Gather(dir, "HEAD~1", "HEAD", cfg); err == nil {
		_ = h // a bare temp dir is not a git repo; the point below is the probe, not the log
	}
}

// A DECLARED source always wins over a probe — the probe is a convenience, and a convenience
// that overrides a declaration is a trap.
func TestADeclaredSourceBeatsTheProbe(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".sporo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "docs", "adr"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "rfcs"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := "project: svc\nsources:\n  decisions: rfcs\nproducts: [svc, svc-worker]\n"
	if err := os.WriteFile(filepath.Join(dir, ".sporo", "config.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if c.Sources.Decisions != "rfcs" {
		t.Fatalf("the declaration must win over the probe that would have found docs/adr; got %q", c.Sources.Decisions)
	}
	if len(c.Products) != 2 {
		t.Fatalf("a project with siblings declares them all; got %v", c.Products)
	}
}

// The per-kind home seam, as a truth table. The recipe home is the flat `home:` (back-compat, so
// every existing caller keeps reading Home); every other kind reads its `homes:` entry; a kind
// with no declared home is an ABSENCE, not an error — the difference between "this project has no
// seed corpus" and a crash. This repo's own config declares `home: recipes/` and `homes: {seed:
// seeds/}`, so it exercises the two declared rows against real values.
func TestHomeForResolvesEachKind(t *testing.T) {
	cfg, err := LoadConfig("../..")
	if err != nil {
		t.Fatal(err)
	}
	if home, ok := cfg.HomeFor(recipekit.KindRecipe); !ok || home != "recipes/" {
		t.Fatalf("recipe home is the flat `home:` key; got (%q, %v)", home, ok)
	}
	if home, ok := cfg.HomeFor(recipekit.KindSeed); !ok || home != "seeds/" {
		t.Fatalf("seed home is its `homes:` entry; got (%q, %v)", home, ok)
	}

	// A project that declares no `homes:` at all: the recipe home still resolves (it is the flat
	// key), and the seed home is a stated absence — ("", false), never a crash.
	bare, err := LoadConfig(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if home, ok := bare.HomeFor(recipekit.KindRecipe); !ok || home != DefaultHome {
		t.Fatalf("a config-less project still answers HomeFor(recipe) with the default; got (%q, %v)", home, ok)
	}
	if home, ok := bare.HomeFor(recipekit.KindSeed); ok || home != "" {
		t.Fatalf("an undeclared kind is absent, not an error; got (%q, %v)", home, ok)
	}
}

// A `homes:` key outside the closed kind vocabulary is a HARD error, not a silently-unreachable
// home. A typo'd kind means a fixable config, never a corpus the walk never reaches — the same
// default-then-validate discipline LoadRegistry applies to a sealed entry's kind.
func TestLoadConfigRejectsUnknownHomeKind(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".sporo"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := "project: svc\nhomes:\n  bogus: heaps/\n"
	if err := os.WriteFile(filepath.Join(dir, ".sporo", "config.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := LoadConfig(dir)
	if err == nil {
		t.Fatal("an unknown home kind must hard-error, not store a home nothing walks")
	}
	if !strings.Contains(err.Error(), "bogus") {
		t.Fatalf("the error must name the offending kind; got %v", err)
	}
}

// The corpus-walk seam: resolve the seed home through HomeFor, walk it, and lint every instance
// against SeedShape. This proves the home is reachable and the genre engine runs over it — the
// mechanism S2 delivers in place of a `sporo lint` CLI verb (deferred to S3). Today the corpus
// holds only the `_`-prefixed genre spec, so the instance loop is empty and MUST NOT fail on that:
// the walk proves the seam, not a populated corpus. The banner-only `_authoring.md` is asserted
// present (the walk reached the home) and is exempt from the instance shape by its `_` prefix.
func TestSeedCorpusWalksAndLintsGreen(t *testing.T) {
	const root = "../.."
	cfg, err := LoadConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	home, ok := cfg.HomeFor(recipekit.KindSeed)
	if !ok {
		t.Fatal("this repo declares a seed home; HomeFor(seed) must resolve it")
	}
	entries, err := os.ReadDir(filepath.Join(root, home))
	if err != nil {
		t.Fatalf("seed home %q must be walkable: %v", home, err)
	}
	foundGenreSpec := false
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".md") {
			continue
		}
		if name == "_authoring.md" {
			foundGenreSpec = true
		}
		if strings.HasPrefix(name, "_") {
			continue // a corpus document teaches the genre; it is not an instance of it
		}
		src, err := os.ReadFile(filepath.Join(root, home, name))
		if err != nil {
			t.Fatalf("read %q: %v", name, err)
		}
		if f := recipekit.LintShape(recipekit.SeedShape, name, src, cfg.Products); len(f) != 0 {
			t.Fatalf("seed %q must lint green against SeedShape: %v", name, f)
		}
	}
	if !foundGenreSpec {
		t.Fatal("the walk must reach the seed home's genre spec (_authoring.md) — the seam is unproven otherwise")
	}
}
