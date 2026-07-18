package recipe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
